package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// 1. map phase
	workerId := CallRegister()

	mapTaskDone := false

	for mapTaskDone == false {

		time.Sleep(1000 * time.Millisecond)

		getMapTaskResp := CallGetMapTask(workerId)
		fmt.Printf("Worker %d get map task, done: %v \n", workerId, getMapTaskResp)
		mapTaskDone = getMapTaskResp.MapTaskDone
		if mapTaskDone {
			break
		}
		fileName := getMapTaskResp.FileName
		reduceNum := getMapTaskResp.ReduceNum

		if len(fileName) == 0 {
			continue
		}
		err := DoMap(workerId, mapf, reduceNum, fileName)
		if err != nil {
			log.Fatalf("do map failed, %v", err)
		} else {
			CallMapTaskDone(workerId, fileName)
		}
	}

	reduceTaskDone := false

	// 2. recude phase
	for reduceTaskDone == false {

		time.Sleep(1000 * time.Millisecond)

		getReduceTaskResp := CallGetReduceTask(workerId)
		fmt.Printf("Worker %d get reduce task, done: %v\n", workerId, getReduceTaskResp)
		reduceTaskDone = getReduceTaskResp.ReduceTaskDone
		if reduceTaskDone {
			break
		}
		reduceId := getReduceTaskResp.ReduceId
		if reduceId < 0 {
			continue
		}

		err := DoReduce(workerId, reducef, reduceId)
		if err != nil {
			log.Fatalf("do reduce failed, %v", err)
		} else {
			CallReduceTaskDone(workerId, reduceId)
		}
	}
}

func DoReduce(workerId int, reducef func(string, []string) string, recudeId int) error {
	fmt.Printf("Begin DoMap, workerid: %d, reduceId: %v\n", workerId, recudeId)
	strAllRecudeFile := fmt.Sprintf("reduce-out-%d-*", recudeId)
	reduceFiles, err := filepath.Glob(strAllRecudeFile)
	if err != nil {
		log.Fatalf("Glob read %v", strAllRecudeFile)
		return err
	}
	// fmt.Printf("reduceFiles: %v\n", reduceFiles)
	allReduceData := []KeyValue{}

	for _, fileName := range reduceFiles {
		content, err := ReadDataFromFile(fileName)
		if err != nil {
			log.Fatalf("cannot read %v", fileName)
			return err
		}

		kvDataList := []KeyValue{}
		err = json.Unmarshal([]byte(content), &kvDataList)
		if err != nil {
			log.Fatalf("Unmarshal %v", fileName)
			return err
		}

		// fmt.Println(kvDataList)
		allReduceData = append(allReduceData, kvDataList...)
	}

	sort.Sort(ByKey(allReduceData))

	outFileName := fmt.Sprintf("mr-out-%d", recudeId)

	ofile, _ := os.Create(outFileName)

	i := 0
	for i < len(allReduceData) {
		j := i + 1
		for j < len(allReduceData) && allReduceData[j].Key == allReduceData[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, allReduceData[k].Value)
		}
		output := reducef(allReduceData[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", allReduceData[i].Key, output)

		i = j
	}

	ofile.Close()

	for _, fileName := range reduceFiles {
		os.Remove(fileName) // clean up
	}

	return nil
}

func DoMap(workerId int, mapf func(string, string) []KeyValue, reduceNum int, fileName string) error {
	fmt.Printf("Begin DoMap, workerid: %d, fiel: %v\n", workerId, fileName)

	content, err := ReadDataFromFile(fileName)
	if err != nil {
		log.Fatalf("cannot read %v", fileName)
		return err
	}

	filedata := mapf(fileName, string(content))

	// sort.Sort(ByKey(allReduceData))
	reduceData := make([][]KeyValue, reduceNum)

	for _, kvData := range filedata {
		reduceIndex := ihash(kvData.Key) % reduceNum
		reduceData[reduceIndex] = append(reduceData[reduceIndex], kvData)
	}

	for index, kvDataList := range reduceData {
		res, err := json.Marshal(kvDataList)
		if err != nil {
			log.Fatalf("kvdata to json failed %v", err)
		}
		strKvDataJson := string(res)

		reducdeFileName := GetRecudeTempFileName(index, fileName)

		err = WriteDataToFile(strKvDataJson, reducdeFileName)
		if err != nil {
			log.Fatalf("write json data to file failed, filename: %v, err: %v", reducdeFileName, err)
		}
	}

	return nil
}

func GetRecudeTempFileName(reduceIndex int, fullPathFileName string) string {
	originFileName := fullPathFileName
	loc := strings.LastIndex(fullPathFileName, "/")
	if loc > 0 {
		originFileName = fullPathFileName[loc+1:]
	}
	reducdeFileName := fmt.Sprintf("reduce-out-%d-%v", reduceIndex, originFileName)
	return reducdeFileName
}

func ReadDataFromFile(fileName string) ([]byte, error) {
	mapfile, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("cannot open %v", fileName)
		return nil, err
	}
	content, err := ioutil.ReadAll(mapfile)
	if err != nil {
		log.Fatalf("cannot read %v", fileName)
		return nil, err
	}
	mapfile.Close()
	return content, nil
}

func WriteDataToFile(data string, fileName string) error {
	ofile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("cannot create file: %v err: %v", fileName, err)
	}

	fmt.Fprintf(ofile, "%v", data)

	ofile.Close()

	return nil
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func CallExample() {
	args := ExampleArgs{}
	args.X = 99
	reply := ExampleReply{}

	call("Master.Example", &args, &reply)
	fmt.Printf("reply.Y %v\n", reply.Y)
}

func CallRegister() int {
	req := WorkerRegisterReq{}
	resp := WorkerRegisterResp{}
	call("Master.WorkerRegister", &req, &resp)
	fmt.Printf("worker %d register\n", resp.WorkerId)
	return resp.WorkerId
}

func CallGetMapTask(workerId int) GetMapTaskResp {
	req := GetMapTaskReq{}
	req.WorkerId = workerId
	resp := GetMapTaskResp{}
	call("Master.GetMapTask", &req, &resp)
	return resp
}

func CallMapTaskDone(workerId int, fileName string) {
	req := MapTaskDoneReq{}
	req.WorkerId = workerId
	req.FileName = fileName
	resp := EmptyResp{}
	call("Master.MapTaskDone", &req, &resp)
}

func CallGetReduceTask(workerId int) GetReduceTaskResp {
	req := GetReduceTaskReq{}
	req.WorkerId = workerId
	resp := GetReduceTaskResp{}
	call("Master.GetReduceTask", &req, &resp)
	return resp
}

func CallReduceTaskDone(workerId int, reduceId int) {
	req := ReduceTaskDoneReq{}
	req.WorkerId = workerId
	req.ReduceId = reduceId
	resp := EmptyResp{}
	call("Master.ReduceTaskDone", &req, &resp)
	return
}
