package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type EmptyReq struct {
}

type EmptyResp struct {
}

type WorkerRegisterReq struct {
}

type WorkerRegisterResp struct {
	WorkerId int
}

type GetMapTaskReq struct {
	WorkerId int
}

type GetMapTaskResp struct {
	MapTaskDone bool
	FileName    string
	ReduceNum   int
}

type MapTaskDoneReq struct {
	WorkerId int
	FileName string
}

type GetReduceTaskReq struct {
	WorkerId int
}

type GetReduceTaskResp struct {
	ReduceId       int
	ReduceTaskDone bool
}

type ReduceTaskDoneReq struct {
	WorkerId int
	ReduceId int
}
