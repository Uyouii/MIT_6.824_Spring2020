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
	workerNum  int
	workers    map[int]WorkerStatus
	workersMux sync.Mutex

	mapTaskDone       bool
	mapFiles          []string
	toSolveMapFileMux sync.Mutex
	toSolveMapFile    []string
	doneMapFilesMux   sync.Mutex
	doneMapFiles      []string

	reduceNum           int
	toSolveReduceIdsMux sync.Mutex
	toSolveReduceIds    []int
	reduceTaskDone      bool
	doneReduceTasksMux  sync.Mutex
	doneReduceTasks     []int
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
	return m.reduceTaskDone
}

func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	m.reduceNum = nReduce
	m.mapFiles = make([]string, len(files))
	copy(m.mapFiles, files)
	m.workers = make(map[int]WorkerStatus)
	m.toSolveMapFile = make([]string, len(files))
	copy(m.toSolveMapFile, files)

	for i := 0; i < m.reduceNum; i++ {
		m.toSolveReduceIds = append(m.toSolveReduceIds, i)
	}

	m.Server()
	go m.Run()

	return &m
}

func (m *Master) Run() {
	// 1. map
	for len(m.doneMapFiles) < len(m.mapFiles) {
		time.Sleep(100 * time.Millisecond)
		// fmt.Printf("cur done mapFiles size: %d, total file size: %d\n", len(m.doneMapFiles), len(m.mapFiles))
		continue
	}
	m.mapTaskDone = true

	// 2. reduce
	for len(m.doneReduceTasks) < m.reduceNum {
		time.Sleep(100 * time.Millisecond)
		continue
	}
	m.reduceTaskDone = true
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

func (m *Master) SetWorkerState(workerId int, status WorkerStatus) {
	if workerId >= len(m.workers) {
		return
	}
	m.workersMux.Lock()
	m.workers[workerId] = status
	m.workersMux.Unlock()
}

func (m *Master) AddDoneMapFiles(fileName string) error {
	m.doneMapFilesMux.Lock()
	m.doneMapFiles = append(m.doneMapFiles, fileName)
	m.doneMapFilesMux.Unlock()
	return nil
}

func (m *Master) AddDoneReduceId(reduceId int) error {
	m.doneReduceTasksMux.Lock()
	m.doneReduceTasks = append(m.doneReduceTasks, reduceId)
	m.doneReduceTasksMux.Unlock()
	return nil
}

func (m *Master) CheckMapTaskRight(fileName string) error {
	time.Sleep(10000 * time.Millisecond)
	bFound := false
	for _, doneFileName := range m.doneMapFiles {
		if doneFileName == fileName {
			bFound = true
			break
		}
	}
	if bFound == false {
		fmt.Printf("file wrong to handle: %v\n", fileName)
		m.toSolveMapFileMux.Lock()
		m.toSolveMapFile = append(m.toSolveMapFile, fileName)
		m.toSolveMapFileMux.Unlock()
	}
	return nil
}

func (m *Master) CheckReduceTaskRight(reduceId int) error {
	time.Sleep(5000 * time.Millisecond)
	bFound := false
	for _, doneReduceId := range m.doneReduceTasks {
		if doneReduceId == reduceId {
			bFound = true
			break
		}
	}
	if bFound == false {
		fmt.Printf("reduce wrong to handle: %v\n", reduceId)
		m.toSolveReduceIdsMux.Lock()
		m.toSolveReduceIds = append(m.toSolveReduceIds, reduceId)
		m.toSolveReduceIdsMux.Unlock()
	}
	return nil
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

	m.SetWorkerState(workerId, WORKER_STATUS_FREE)

	fmt.Printf("worler %d register\n", workerId)
	return nil
}

func (m *Master) GetMapTask(req *GetMapTaskReq, resp *GetMapTaskResp) error {
	resp.MapTaskDone = m.mapTaskDone
	resp.ReduceNum = m.reduceNum

	if len(m.toSolveMapFile) == 0 {
		return nil
	}

	m.SetWorkerState(req.WorkerId, WORKER_STATUS_BUSY)

	m.toSolveMapFileMux.Lock()
	resp.FileName = m.toSolveMapFile[0]
	m.toSolveMapFile = m.toSolveMapFile[1:]
	m.toSolveMapFileMux.Unlock()

	go m.CheckMapTaskRight(resp.FileName)

	fmt.Printf("worler %d get map task: %v\n", req.WorkerId, resp.FileName)

	return nil
}

func (m *Master) MapTaskDone(req *MapTaskDoneReq, resp *EmptyResp) error {

	m.SetWorkerState(req.WorkerId, WORKER_STATUS_FREE)

	fmt.Printf("worler %d done map task: %v\n", req.WorkerId, req.FileName)

	m.AddDoneMapFiles(req.FileName)

	return nil
}

func (m *Master) GetReduceTask(req *GetReduceTaskReq, resp *GetReduceTaskResp) error {
	resp.ReduceTaskDone = m.reduceTaskDone

	if len(m.toSolveReduceIds) == 0 {
		resp.ReduceId = -1
		return nil
	}

	m.SetWorkerState(req.WorkerId, WORKER_STATUS_BUSY)

	m.toSolveReduceIdsMux.Lock()
	resp.ReduceId = m.toSolveReduceIds[0]
	m.toSolveReduceIds = m.toSolveReduceIds[1:]
	m.toSolveReduceIdsMux.Unlock()

	go m.CheckReduceTaskRight(resp.ReduceId)

	fmt.Printf("worler %d get reduce task: %v\n", req.WorkerId, resp.ReduceId)

	return nil
}

func (m *Master) ReduceTaskDone(req *ReduceTaskDoneReq, resp *EmptyResp) error {

	m.SetWorkerState(req.WorkerId, WORKER_STATUS_FREE)

	fmt.Printf("worker %d done reduce task: %v\n", req.WorkerId, req.ReduceId)

	m.AddDoneReduceId(req.ReduceId)

	return nil
}
