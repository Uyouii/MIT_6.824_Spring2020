package mr

import (
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
	isMapDone    bool
	workerNum    int
	workers      map[int]WorkerStatus
	workersMux   sync.Mutex
	reduceNum    int
	mapFiles     []string
	curMapId     int32
	doneMapFiles []string
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

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	m.reduceNum = nReduce
	m.mapFiles = files

	m.Server()
	m.Run()

	return &m
}

func (m *Master) Run() {
	// 1. map
	for len(m.doneMapFiles) < len(m.mapFiles) {
		freeWorkerId := m.GetFreeWorker()
		// no free worker
		if freeWorkerId < 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
	}
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
	resp.workerId = workerId
	m.workerNum += 1

	m.workersMux.Lock()
	m.workers[workerId] = WORKER_STATUS_FREE
	m.workersMux.Unlock()

	return nil
}

// func (m *Master) GetMapTask()
