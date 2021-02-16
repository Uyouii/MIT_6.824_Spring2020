package raft

//
// this is an outline of the API that raft must expose to the service (or tester).
// see comments below for each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester) in the same server.
//

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

const ELECTION_TIME int = 300
const HEART_BEAT_INTERVAL int = 200
const SEND_REQUEST_TIME_OUT int = 300
const SHOW_LOG bool = false
const RAFT_RUN_LOOP int = 100

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g., snapshots) on the applyCh;
// at that point you can add fields to ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type RaftPeerState int32

const (
	PEER_STATE_DEFAULT      RaftPeerState = 0
	PEER_STATE_TRY_ELECTION RaftPeerState = 1
	PEER_STATE_CANDIDATE    RaftPeerState = 2
	PEER_STATE_FOLLOWER     RaftPeerState = 3
	PEER_STATE_LEADER       RaftPeerState = 4
)

type RaftEntry struct {
	Cmd   interface{}
	Index int
	Term  int
}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what state a Raft server must maintain.
	currentTerm  int // latest term server has seen (initialized to 0 on first boot, increases monotonically)
	electionTerm int
	termMux      sync.Mutex
	votedFor     int         // candidateId that received vote in current term (or null if none)
	logs         []RaftEntry // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1
	logMux       sync.Mutex

	// Volatile state on all servers
	commitIndex int // index of highest log entry known to be committed (initialized to 0, increases monotonically)
	lastApplied int // index of highest log entry applied to state machine (initialized to 0, increases monotonically)

	// Volatile state on leaders, Reinitialized after election
	nextIndex  []int //  for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	matchIndex []int // for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

	stateMap           map[int]RaftPeerState
	stateMux           sync.Mutex
	heartBeatTimeStamp int64
}

func (rf *Raft) Init(me int, peers []*labrpc.ClientEnd, persister *Persister) {
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	rf.dead = 0

	rf.currentTerm = 0
	rf.votedFor = -1
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.electionTerm = 0

	rf.heartBeatTimeStamp = 0

	rf.stateMap = make(map[int]RaftPeerState)
	rf.stateMap[0] = PEER_STATE_DEFAULT

	rf.Print(fmt.Sprintf("Init, id %d", rf.me))
}

func (rf *Raft) Print(log string) {
	if SHOW_LOG {
		fmt.Printf("-------- %s\n", log)
	}
}

// return currentTerm and whether this server believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	term := rf.GetTerm()
	return term, rf.GetTermState(term) == PEER_STATE_LEADER
}

func (rf *Raft) GetStateString() string {
	state, isleader := rf.GetState()
	return fmt.Sprintf("id: %d, state: %d, term: %d, isleader: %t", rf.me, state, rf.GetTerm(), isleader)
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	rf.Print(fmt.Sprintf("Kill, id: %d", rf.me))
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) SetHeartTimeStamp(time int64) {
	atomic.StoreInt64(&rf.heartBeatTimeStamp, time)
}

func (rf *Raft) GetHeartTimeStamp() int64 {
	return atomic.LoadInt64(&rf.heartBeatTimeStamp)
}

func (rf *Raft) GetRandomTimeOut() int {
	nowTime := time.Now().UnixNano()
	rand.Seed(nowTime)
	return rand.Intn(ELECTION_TIME) + ELECTION_TIME
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int // candidate’s term
	CandidateId  int // candidate requesting vote
	LastLogIndex int // index of candidate’s last log entry
	LastLogTerm  int // term of candidate’s last log entry
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	if rf.GetTerm() >= args.Term {
		reply.VoteGranted = false
		return
	}

	ret := rf.ChangeSelfTerm(args.Term)
	if ret {
		rf.SetTermState(args.Term, PEER_STATE_FOLLOWER)
		reply.VoteGranted = true
		reply.Term = args.Term
		return
	}

	reply.VoteGranted = false
}

func (rf *Raft) ChangeSelfTerm(term int) bool {
	rf.termMux.Lock()
	defer rf.termMux.Unlock()

	if rf.currentTerm >= term {
		return false
	}
	rf.currentTerm = term

	return true
}

func (rf *Raft) GetTerm() int {
	rf.termMux.Lock()
	defer rf.termMux.Unlock()
	return rf.currentTerm
}

func (rf *Raft) SetTermState(term int, state RaftPeerState) {
	rf.stateMux.Lock()
	rf.stateMap[term] = state
	rf.stateMux.Unlock()
}

func (rf *Raft) GetTermState(term int) RaftPeerState {
	rf.stateMux.Lock()
	defer rf.stateMux.Unlock()
	return rf.stateMap[term]
}

type AppendEntriesArgs struct {
	Term         int // leader’s term
	LeaderId     int // so follower can redirect clients
	PrevLogIndex int // index of log entry immediately preceding new ones
	PrevLogTerm  int // term of prevLogIndex entry
	LeaderCommit int // leader’s commitIndex
	Entries      []RaftEntry
}

type AppendEntriesReply struct {
	Term    int  // currentTerm, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	term := rf.GetTerm()

	if term > args.Term {
		reply.Term = term
		reply.Success = false
		return
	}

	if term < args.Term {
		rf.ChangeSelfTerm(args.Term)
	}

	rf.SetTermState(args.Term, PEER_STATE_FOLLOWER)

	// heart beat
	if len(args.Entries) == 0 {
		rf.SetHeartTimeStamp(int64(time.Now().UnixNano() / 1e6))
	}

	reply.Term = args.Term
	reply.Success = true
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	beginMsTime := time.Now().UnixNano() / 1e6

	rf.Print(fmt.Sprintf("sendRequestVote, candidateid: %d, reciver: %d, reqterm: %d",
		args.CandidateId, server, args.Term))

	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)

	endMsTime := time.Now().UnixNano() / 1e6
	intervalMsTime := endMsTime - beginMsTime

	if ok == false {
		rf.Print(fmt.Sprintf("sendRequestVote failed, candidateid: %d, reciver: %d, reqterm: %d, cost: %d",
			args.CandidateId, server, args.Term, intervalMsTime))
	} else {
		rf.Print(fmt.Sprintf("sendRequestVotesendRequestVote succ, candidateid: %d, receiver: %d, reqterm: %d, selfterm: %d, granted: %t, cost: %d",
			args.CandidateId, server, args.Term, rf.GetTerm(), reply.VoteGranted, intervalMsTime))
	}

	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {

	rf.Print(fmt.Sprintf("sendAppendEntries,  leader: %d, reciver: %d, leaderterm: %d",
		args.LeaderId, server, args.Term))

	beginMsTime := time.Now().UnixNano() / 1e6
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	endMsTime := time.Now().UnixNano() / 1e6
	intervalMsTime := endMsTime - beginMsTime

	if ok == false {
		rf.Print(fmt.Sprintf("sendAppendEntries failed,  leader: %d, reciver: %d, leaderterm: %d, cost: %d",
			args.LeaderId, server, args.Term, intervalMsTime))
	} else {
		rf.Print(fmt.Sprintf("sendAppendEntries succ, leader: %d, reciver: %d, leaderterm: %d, meterm: %d, ret: %t, cost : %d",
			args.LeaderId, server, args.Term, rf.GetTerm(), reply.Success, intervalMsTime))
	}

	return ok
}

func (rf *Raft) AppendLog(command interface{}) RaftEntry {

	rf.logMux.Lock()
	defer rf.logMux.Unlock()

	entry := RaftEntry{
		Cmd:   command,
		Term:  rf.GetTerm(),
		Index: len(rf.logs) + 1,
	}

	rf.logs = append(rf.logs, entry)

	return entry
}

//
// the service using Raft (e.g. a k/v server) wants to start agreement on the next command to be appended to Raft's log.
// if this server isn't the leader, returns false.
// otherwise start the agreement and return immediately.
// there is no guarantee that this command will ever be committed to the Raft log, since the leader may fail or lose an election.
// even if the Raft instance has been killed, this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {

	fmt.Printf("Start, commamd: %d, %v\n", command, rf.GetStateString())

	term, isLeader := rf.GetState()
	index := -1

	if !isLeader {
		return index, term, isLeader
	}

	entry := rf.AppendLog(command)
	go rf.SendEntry(&entry)

	return index, term, isLeader
}

func (rf *Raft) SendEntry(entry *RaftEntry) {
	// entryTerm := entry.Term

}

func (rf *Raft) BroadcastAppendEntry(args *AppendEntriesArgs) {
	var wg sync.WaitGroup

	for index := 0; index < len(rf.peers); index++ {
		if index == rf.me {
			continue
		}

		wg.Add(1)

		var reply AppendEntriesReply

		reqDone := int32(0)

		go func(index int) {
			ret := rf.sendAppendEntries(index, args, &reply)
			if atomic.LoadInt32(&reqDone) == 0 {
				atomic.StoreInt32(&reqDone, 1)
				wg.Done()
			}
			if ret && reply.Success == false && reply.Term > rf.GetTerm() {
				rf.ChangeSelfTerm(reply.Term)
				rf.SetTermState(reply.Term, PEER_STATE_FOLLOWER)
				return
			}
		}(index)

		go func(index int) {
			time.Sleep(time.Duration(SEND_REQUEST_TIME_OUT) * time.Millisecond)
			done := atomic.LoadInt32(&reqDone)
			if done == 0 {
				atomic.StoreInt32(&reqDone, 1)
				wg.Done()
			}
		}(index)
	}

	wg.Wait()
}

func (rf *Raft) StartHeatBeat(term int) {

	for term == rf.GetTerm() &&
		rf.GetTermState(term) == PEER_STATE_LEADER &&
		!rf.killed() {

		args := AppendEntriesArgs{
			Term:     rf.GetTerm(),
			LeaderId: rf.me,
		}

		beginMsTime := time.Now().UnixNano() / 1e6

		rf.BroadcastAppendEntry(&args)

		endMsTime := time.Now().UnixNano() / 1e6
		intervelMsTime := endMsTime - beginMsTime

		if intervelMsTime < int64(HEART_BEAT_INTERVAL) {
			time.Sleep(time.Duration(int64(HEART_BEAT_INTERVAL)-intervelMsTime) * time.Millisecond)
		}
	}
}

func (rf *Raft) BeLeader(term int) {
	rf.Print(fmt.Sprintf("BeLeader, id: %d, currentTerm: %d", rf.me, rf.GetTerm()))
	rf.SetTermState(term, PEER_STATE_LEADER)
	go rf.StartHeatBeat(term)
}

func (rf *Raft) CheckHeartBeat(term int) {
	if term == 0 {
		return
	}

	state := rf.GetTermState(term)
	if state == PEER_STATE_FOLLOWER {
		curMsTime := time.Now().UnixNano() / 1e6
		if curMsTime-rf.GetHeartTimeStamp() >= int64(SEND_REQUEST_TIME_OUT) {

			rf.TryElection(term + 1)

		}
	}
}

func (rf *Raft) TryElection(newTrem int) {

	state := rf.GetTermState(newTrem)
	if state != PEER_STATE_DEFAULT {
		return
	}

	rf.SetTermState(newTrem, PEER_STATE_TRY_ELECTION)

	timeOut := rf.GetRandomTimeOut()
	time.Sleep(time.Duration(timeOut) * time.Millisecond)

	if rf.GetTerm() >= newTrem {
		return
	}

	rf.SetTermState(newTrem, PEER_STATE_CANDIDATE)
	rf.ChangeSelfTerm(newTrem)

	rf.Print(fmt.Sprintf("TryElection Begin, id: %d, currentTerm: %d",
		rf.me, newTrem))

	var wg sync.WaitGroup

	voteCount := int32(0)

	for index := 0; index < len(rf.peers); index++ {
		if index == rf.me {
			continue
		}

		wg.Add(1)
		args := RequestVoteArgs{
			Term:        newTrem,
			CandidateId: rf.me,
		}
		var reply RequestVoteReply

		reqDone := int32(0)

		go func(index int) {
			ret := rf.sendRequestVote(index, &args, &reply)
			if atomic.LoadInt32(&reqDone) == 0 {
				atomic.StoreInt32(&reqDone, 1)
				wg.Done()
			} else {
				return
			}
			if ret && reply.VoteGranted && reply.Term == newTrem {
				atomic.AddInt32(&voteCount, 1)
			}
		}(index)

		go func(index int) {
			time.Sleep(time.Duration(SEND_REQUEST_TIME_OUT) * time.Millisecond)
			done := atomic.LoadInt32(&reqDone)
			if done == 0 {
				atomic.StoreInt32(&reqDone, 1)
				wg.Done()
			}
		}(index)
	}

	wg.Wait()

	count := atomic.LoadInt32(&voteCount)

	rf.Print(fmt.Sprintf("TryElection, id: %d, term: %d, voteCount: %d, peerscount: %d",
		rf.me, rf.GetTerm(), count, len(rf.peers)))

	if count >= int32((len(rf.peers)-1)/2) && newTrem == rf.GetTerm() {
		rf.BeLeader(newTrem)
	} else {
		rf.SetTermState(rf.GetTerm(), PEER_STATE_FOLLOWER)
	}

}

func (rf *Raft) CheckStatus() {
	for !rf.killed() {
		rf.TryElection(1)
		rf.CheckHeartBeat(rf.GetTerm())

		time.Sleep(time.Duration(RAFT_RUN_LOOP) * time.Millisecond)

	}
}

func (rf *Raft) PrintStatus() {
	for !rf.killed() {
		time.Sleep(300 * time.Millisecond)
		term := rf.GetTerm()

		rf.Print(fmt.Sprintf("PrintStatus, id: %d, term: %d, state: %d",
			rf.me, term, rf.GetTermState(term)))
	}
}

func (rf *Raft) Run() {
	rf.Print(fmt.Sprintf("Raft Run, id: %d", rf.me))

	go rf.CheckStatus()
	go rf.PrintStatus()
}

//
// the service or tester wants to create a Raft server.
// the ports of all the Raft servers (including this one) are in peers[].
// this server's port is peers[me].
// all the servers' peers[] arrays have the same order.
// persister is a place for this server to save its persistent state,
// and also initially holds the most recent saved state, if any.
// applyCh is a channel on which the tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}

	rf.Init(me, peers, persister)
	go rf.Run()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
