package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
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
	workerId := Register()

	mapTaskDone := false

	for mapTaskDone == false {
		getMapTaskResp := GetMapTask(workerId)
		fmt.Printf("Worker %d get map task, done: %v file: %v\n", workerId, getMapTaskResp, getMapTaskResp.FileName)
		mapTaskDone = getMapTaskResp.MapTaskDone
		if mapTaskDone {
			break
		}
		fileName := getMapTaskResp.FileName
		reduceNum := getMapTaskResp.ReduceNum

		if len(fileName) == 0 {
			time.Sleep(2 * time.Millisecond)
			continue
		}
		err := DoMap(workerId, mapf, reduceNum, fileName)
		if err != nil {
			log.Fatalf("do map failed, %v", err)
		}
	}
}

func DoMap(workerId int, mapf func(string, string) []KeyValue, reduceNum int, fileName string) error {
	fmt.Printf("Begin DoMap, workerid: %d, fiel: %v\n", workerId, fileName)

	content, err := ReadDataFromFile(fileName)
	if err != nil {
		log.Fatalf("cannot read %v", fileName)
		return err
	}

	filedata := mapf(fileName, string(content))

	// sort.Sort(ByKey(intermediate))
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

func GetRecudeTempFileName(reduceIndex int, originFileName string) string {
	reducdeFileName := fmt.Sprintf("reduce-out-%v-%d", originFileName, reduceIndex)
	reducdeFileName = "tempReduceFile/" + reducdeFileName
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

	fmt.Fprintf(ofile, "%v", err)

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

func Register() int {
	req := WorkerRegisterReq{}
	resp := WorkerRegisterResp{}
	call("Master.WorkerRegister", &req, &resp)
	fmt.Printf("worker %d register\n", resp.WorkerId)
	return resp.WorkerId
}

func GetMapTask(workerId int) GetMapTaskResp {
	req := GetMapTaskReq{}
	req.WorkerId = workerId
	resp := GetMapTaskResp{}
	call("Master.GetMapTask", &req, &resp)
	return resp
}
