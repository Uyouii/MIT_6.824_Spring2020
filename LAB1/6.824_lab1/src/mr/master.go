package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type WorkerStatus int32

const (
	WORKER_STATUS_FREE WorkerStatus = 0
	WORKER_STATUS_BUSY WorkerStatus = 1
	WORKER_STATUS_DOWN WorkerStatus = 2
)

type Master struct {
	mapTaskDone     bool
	workerNum       int
	workers         map[int]WorkerStatus
	workersMux      sync.Mutex
	reduceNum       int
	reduceHandleBit []int
	reduceTaskDone  bool
	mapFiles        []string
	curHandleMapId  int
	doneMapFiles    []string
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) Server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false

	// Your code here.

	return ret
}

func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	m.reduceNum = nReduce
	m.mapFiles = make([]string, len(files))
	copy(m.mapFiles, files)
	m.reduceHandleBit = make([]int, nReduce)
	m.workers = make(map[int]WorkerStatus)

	m.Server()
	go m.Run()

	return &m
}

func (m *Master) Run() {
	// 1. map
	for len(m.doneMapFiles) < len(m.mapFiles) {
		time.Sleep(100 * time.Millisecond)
		continue
	}
	m.mapTaskDone = true

	// 2. reduce
}

// get free worker
func (m *Master) GetFreeWorker() int {

	defer m.workersMux.Unlock()

	m.workersMux.Lock()
	for workerId, workerStatus := range m.workers {
		if workerStatus == WORKER_STATUS_FREE {
			return workerId
		}
	}
	return -1
}

// ========================== rpc ==========================

func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (m *Master) WorkerRegister(req *WorkerRegisterReq, resp *WorkerRegisterResp) error {
	workerId := m.workerNum
	resp.WorkerId = workerId
	m.workerNum += 1

	m.workersMux.Lock()
	m.workers[workerId] = WORKER_STATUS_FREE
	m.workersMux.Unlock()
	fmt.Printf("worler %d register\n", workerId)
	return nil
}

func (m *Master) GetMapTask(req *GetMapTaskReq, resp *GetMapTaskResp) error {
	resp.MapTaskDone = m.mapTaskDone
	resp.ReduceNum = m.reduceNum

	if m.curHandleMapId >= len(m.mapFiles)-1 {
		return nil
	}

	m.workersMux.Lock()
	m.workers[req.WorkerId] = WORKER_STATUS_BUSY
	m.workersMux.Unlock()

	resp.FileName = m.mapFiles[m.curHandleMapId]
	m.curHandleMapId += 1

	fmt.Printf("worler %d get map task: %v\n", req.WorkerId, resp.FileName)

	return nil
}
