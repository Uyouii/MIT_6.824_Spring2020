package mr

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/rpc"
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

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue, reducef func(string, []string) string) {

	// 1. map phase
	workerId := Register()

	mapTaskDone := false

	for mapTaskDone == false {
		getMapTaskResp := GetMapTask(workerId)
		mapTaskDone = getMapTaskResp.mapTaskDone
		if mapTaskDone {
			break
		}
		fileName := getMapTaskResp.fileName

		if len(fileName) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// DoMap(workerId, mapf, fileName)
	}
}

func DoMap(workerId int, mapf func(string, string) []KeyValue, fileName string) {
	fmt.Printf("Begin DoMap, workerid: %d, fiel: %v\n", workerId, fileName)
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
	fmt.Printf("worker %d register\n", resp.workerId)
	return resp.workerId
}

func GetMapTask(workerId int) GetMapTaskResp {
	req := GetMapTaskReq{workerId}
	resp := GetMapTaskResp{}
	call("Master.GetMapTask", &req, &resp)
	return resp
}
