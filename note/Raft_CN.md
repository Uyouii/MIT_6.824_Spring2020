

[In Search of an Understandable Consensus Algorithm](https://pdos.csail.mit.edu/6.824/papers/raft-extended.pdf)

Raft是一个共识算法。Raft是提供与Paxos相同的容错性及性能，但比Paxos更好理解的算法，同时为构建实际应用提供更好的基础。

Raft相比与其他一致性算法一些新特性：

- **Strong leader**: Raft uses a stronger from of leadership than other consensus algorithms. e.g. log entries only flow from the leader to other servers.
- **Leader election**: Raft uses randomized timers to elect leaders. This adds only a small amount of mechanism to the heartbeats already required for ay consensus algorithm, while resolving conflicts simply and rapidly.
- **Membership changes**: Raft's mechanism for changing the set of servers in the cluster uses a new _joint consensus_ approach where the majorities of two different configurations overlap during transitions. This allows the cluster to continue operating normally during configuration changes.

## 2 Replicated state machines

Replicated state machines are typically implemented using a replicated log.

![image](https://note.youdao.com/yws/api/personal/file/WEBe739089bd15e43c79f449eab3839941f?method=download&shareKey=4b7217a032db60840a3f7d70034f603d)

> Replicated state machine architecture. The consensus algorithm manages a replicated log containing state machine commands from clients. The state machines process identical sequences of commands from the logs, so they produce the same outputs.

Each server stores a log containing a series of commands, which its state machine executes in order, so each state machine processes the same sequence of commands. Since the state machines are deterministic, ecah computes the same state and the same sequence of outputs.

Keeping the replicated log consistent is the job of the consensus algorithm. The consensus module on a server receives commands from clients and adds them to its log. It communicates with the consensus modules on other servers to ensure that every log eventually contains the same requests in the same order, even if some servers fail. Once commands are properly replicated, each server’s state machine processes them in log order, and the out- puts are returned to clients. 


一致性算法在实际的系统中通常有以下特性：

- 在不复杂的情况下都能确保安全性(safty, 从不返回错误的结果)，包括 网络延迟，partitions, and packet loss, duplication, and reordering.
- They are fully functional(available) as long as any majority of the servers are opreational and can communicate with each other and with clients. Servers are assumes to fail by stoppong; they may later recover from state on stable storage and rejoin the cluster.
- They do not depend on timing to ensure the consistency of the logs: faulty clocks and extreme message delays can, at worst, cause availability problems.
- in the common case, a command can complete as soon as a majority of the cluster has responded to a single round of remote procedure calls; a minority of slow servers need not impact overall system performance.

## 3 What's wrong with Paxos


Paxos first defines a protocol capable of reaching agreement on a single decision, such as a single replicated log entry. We refer to this subset as _single-decree Paxos_. Paxos then combines multiple instances of this protocol to facilitate a series of decisions such as a log (_multi-Paxos_).

Paxos drawbacks:

1. Paxos is exceptionally difficult to understand.
2. Paxos does not provide a good foundation for building practical implementations.
3. Paxos architecture is a poor one for building practical systems.

## 5 The Raft consensus algorithm

Raft implements consensus by first electing a **_distinguished leader_**, then giving the leader complete responsibility for managing the replicated log. The leader accepts log entries from clients, replicates them on other servers, and tells servers when it si safe to apply log entries to their state machines. 

A leader can fail or become disconnected form the other servers, in which case a new leader is elected.

Raft decomposesthe consensus problem into three relatively independent subproblems:

1. **Leader election**: a new leader must be chosen when an existing leader fails.
2. **Log replication**: the leader must accept log entries from clients and replicate them across the cluster, forcing the other logs to agree with its own.
3. **Safety**: the key safety property for Raft is the State Machine Safety Property: if any server has applied a particular log entry to its state machine, then no other server may apply a different command for the same log index.

![image](https://note.youdao.com/yws/api/personal/file/WEBb5b3d7c90c5cddc2c5d5596dfd28bf0c?method=download&shareKey=dd050d25aac6c27770f39c81cb071464)

> A condensed summary of the Raft consensus algorithm



- **Election Safety**: at most one leader can be elected in a given term. **5.2**
- **Leader Append-Only**: a leader never overwrites or deletes entries in its log; it only appends new entries. **5.3**
- **Log Matching** : if two logs contain an entry with the same index and term, then the logs are identical in all entries up through the given index. **5.3**
- **Leader Completeness**: if a log entry is committed in a given term, then that entry will be present in the logs of the leaders for all higher-numbered terms. **5.4**
- **State Maching Safety**: if a server has applied a log entry at a given index to its state machine, no other server wiil ever apply a different log entry for the same index. **5.4.3**

>  Raft guarantees that each of these properties is true at all times. 


### 5.1 Raft basics

A Raft cluster contains several servers; five is a typical number, which allows the system to 
tolerate two failures.

each server is in one of three states: **_leader, follower_**, or **_candidate_**. 

In normal operation there is exactly one leader and all of the other servers are followers. Followers are _passive_: they issue no requests on their own but simply respond to requests from leaders and candidates. 

<font color=#FB8E03>The leader handles all client requests</font> (if a client contacts a follower, the follower redirects it to the leader). The third state, candidate, is used to elect a new leader.

![image](https://note.youdao.com/yws/api/personal/file/WEBac243c2253bc4acafd81e921418fa947?method=download&shareKey=9cdee868fbdcd78f4d0e533eabd4f2e5)

> Server states. Followers only respond to requests from other servers. If a follower receives no communication, it becomes a candidate and initiates an election. A candidate that receives votes from a majority of the full cluster becomes the new leader. Leaders typically operate until they fail.

Raft divides time into **_terms_** of arbitrary length. Term are numbered with consecutive inegers.  Each term begins with an **_election_**, in which one or more candidates attempt to become leader.

In some situations an election will result in a split vote. In this case the term will end with no leader; a new term (with a new election) will begin shortly. Raft ensures that there is at most one leader in a given term.

![image](https://note.youdao.com/yws/api/personal/file/WEB5ebdb7c28d9dd2a490a140a855315b46?method=download&shareKey=77f2cee319d1a8116defd336fc8a2c18)

> Time is divided into terms, and each term begins with an election. After a successful election, a single leader manages the cluster until the end of the term. Some elections fail, in which case the term ends without choosing a leader. The transitions between terms may be observed at different times on different servers.

Different servers may observe the transitions between terms at different times, and in some situations a server may not observe an election or even entire terms. Terms act as a logical clock in Raft, and they allow servers to detect obsolete information such as stale leaders.

Each server stores a _current_ term number, which increases monotonically over time. Current terms are exchanged whenever servers communicate. 

If a candidate or leader discovers that its term is out of date, it immediately reverts to follower state. If a server receives a request with a stale term number, it rejects the request.

Raft servers communicate using remote procedure calls (RPCs), and the basic consensus algorithm requires only two types of RPCs:

1. RequestVote RPCs, initiated by candidates during elections
2. AppendEntries RPCs are initiated by leaders to replicate log entries and to provide a form of heartbeat 

Servers retry RPCs if they do not receive a response in a timely manner, and they issue RPCs in parallel for best performance.

### 5.2 Leader election

Raft uses a heartbeat mechanism to trigger elader election. 

When servers start up, they begin as followers. A server remains in follower state as long as it receives valid RPCs from a leader or candidate. Leaders send periodic heartbeats (AppendEntries RPCs that carry no log entries) to all followers in order to maintain their authority.  If a follower receives no communication over a period of time called the **_election timeout_**, then it assumes there is no viable leader and begins an election to choose a new leader.

To begin an election, a follower increments its current term and transitions to candidate state. It then votes for itself and issues RequestVote RPCs in parallel to each of the other servers in the cluster. 

A candidate continues in this state until one of three things happens:

1. it wins the election
2. another server establishes itself as leader
3. a period of time goes by with no winner

A candidate wins an election if it receives votes from a majority of the servers in the full cluster for the same term. Each server will vote for at most one candidate in a given term, _**on a first-come-first-served basis**_. 

The majority rule ensures that at most one candidate can win the election for a particular term. Once a candidate wins an election, it becomes leader. It then sends heartbeat messages to all of the other servers to establish its authority and prevent new elections.

While waiting for votes, a candidate may receive an **AppendEntries RPC** from another server claiming to be leader. If the leader’s term (included in its RPC) is at least as large as the candidate’s current term, then the candidate recognizes the leader as legitimate and returns to follower state. If the term in the RPC is smaller than the candidate’s current term, then the candidate rejects the RPC and continues in candidate state.

The third possible outcome is that a candidate neither wins nor loses the election: if many followers become candidates at the same time, votes could be split so that no candidate obtains a majority. When this happens, each candidate will time out and start a new election by incrementing its term and initiating another round of RequestVote RPCs. However, without extra measures split votes could repeat indefinitely.

Raft uses **randomized election timeouts** to ensure that split votes are rare and that they are resolved quickly. To prevent split votes in the first place, election timeouts are chosen randomly from a fixed interval (e.g., 150–300ms). This spreads out the servers so that in most cases only a single server will time out; it wins the election and sends heartbeats before any other servers time out.

The same mechanism is used to handle split votes. Each candidate restarts its randomized election timeout at the start of an election, and it waits for that timeout to elapse before starting the next election; this reduces the likelihood of another split vote in the new election.

### 5.3 Log replication

Each client request contains a command to be executed by the replicated state machines. The leader appends the command to its log as a new entry, then issues AppendEntries RPCs in parallel to each of the other servers to replicate the entry. When the entry has been safely replicated, the leader applies the entry to its state machine and returns the result of that execution to the client. If followers crash or run slowly, or if network packets are lost, the leader retries AppendEntries RPCs indefinitely (even after it has responded to the client) until all followers eventually store all log entries.

Each log entry stores a state machine command along with the term number when the entry was received by the leader.Each log entry also has an integer index identifying its position in the log.

![image](https://note.youdao.com/yws/api/personal/file/WEB403f827c3cb3aa84f948765e42bb42de?method=download&shareKey=4e48f424e263f500ebb92d5c31b21dcb)

>  Logs are composed of entries, which are numbered sequentially. Each entry contains the term in which it was created (the number in each box) and a command for the state machine. An entry is considered **_committed_** if it is safe for that entry to be applied to state machines.

The leader decides when it is safe to apply a log entry to the state machines; such an entry is called **_<u>committed</u>_**.  Raft guarantees that committed entries are durable and will eventually be executed by all of the available state machines. <u>A log entry is committed once the leader that created the entry has replicated it on a majority of the servers</u>. This also commits all preceding entries in the leader's log,  including entries created by previous leaders.

The leader keeps track of the highest index it knows to be committed, and it includes that index in future AppendEntries RPCs (including heartbeats) so that the other servers eventually find out. Once a follower learns that a log entry is committed, it applies the entry to its local state machine (in log order).

Raft maintains the following properties:

- If two entries in different logs have the same index and term, then they store the same command.
- If two entries in different logs have the same index and term, then the logs are identical in all preceding entries.

The first property follows from the fact that a leader creates at most one entry with a given log index in a given term, and log entries never change their position in the log. 


The second property is guaranteed by a simple consistency check performed by AppendEntries. When sending an AppendEntries RPC, the leader includes the index and term of the entry in its log that immediately precedes the new entries. If the follower does not find an entry in its log with the same index and term, then it refuses the new entries. The consistency check acts as an induction step: the initial empty state of the logs satisfies the ***Log Matching Property*​**, and the consistency check preserves the ***Log Matching Property*** whenever logs are extended. As a result, whenever AppendEntries returns successfully, the leader knows that the follower’s log is identical to its own log up through the new entries. 

During normal operation, the logs of the leader and followers stay consistent, so the AppendEntries consistency check never fails. However, leader crashes can leave the logs inconsistent (the old leader may not have fully replicated all of the entries in its log). These inconsistencies can compound over a series of leader and follower crashes. A follower may be missing entries that are present on the leader, it may have extra entries that are not present on the leader, or both. Missing and extraneous entries in a log may span multiple terms.

![image](https://note.youdao.com/yws/api/personal/file/WEBba0d2f63e4fa6189033b55992f20f6ee?method=download&shareKey=5e16c868e8dbdbe3a3b4013f03cc775c)

> When the leader at the top comes to power, it is possible that any of scenarios (a–f) could occur in follower logs. Each box represents one log entry; the number in the box is its term. A follower may be missing entries (a–b), may have extra uncommitted entries (c–d), or both (e–f). For example, scenario (f) could occur if that server was the leader for term 2, added several entries to its log, then crashed before committing any of them; it restarted quickly, became leader for term 3, and added a few more entries to its log; before any of the entries in either term 2 or term 3 were committed, the server crashed again and remained down for several terms.

In Raft, <u>the leader handles inconsistencies by forcing the followers’ logs to duplicate its own</u>. This means that conflicting entries in follower logs will be overwritten with entries from the leader’s log.

To bring a follower’s log into consistency with its own, the leader must find the latest log entry where the two logs agree, delete any entries in the follower’s log after that point, and send the follower all of the leader’s entries after that point. All of these actions happen in response to the consistency check performed by AppendEntries RPCs. 

The leader maintains a **_nextIndex_** for each follower, which is the index of the next log entry the leader will send to that follower. When a leader first comes to power, it initializes all nextIndex values to the index just after the last one in its log. If a follower’s log is inconsistent with the leader’s, the AppendEntries consistency check will fail in the next AppendEntries RPC. After a rejection, the leader decrements nextIndex and retries the AppendEntries RPC. Eventually nextIndex will reach a point where the leader and follower logs match. When this happens, AppendEntries will succeed, which removes any conflicting entries in the follower’s log and appends entries from the leader’s log(if any). Once AppendEntries succeeds, the follower’s log is consistent with the leader’s, and it will remain that way for the rest of the term.

If desired, the protocol can be optimized to reduce the number of rejected AppendEntries RPCs. For example, when rejecting an AppendEntries request, the follower can include the term of the conflicting entry and the first index it stores for that term. With this information, the leader can decrement nextIndex to bypass all of the conflicting entries in that term; one AppendEntries RPC will be required for each term with conflicting entries, rather than one RPC per entry.

With this mechanism, a leader does not need to take any special actions to restore log consistency when it comes to power. It just begins normal operation, and the logs auto-matically converge in response to failures of the AppendEntries consistency check. A leader never overwrites or deletes entries in its own log.

Raft can accept, replicate, and apply new log entries as long as a majority of the servers are up; in the normal case a new entry can be replicated with a single round of RPCs to a majority of the cluster; and a single slow follower will not impact performance.

### 5.4 Safety

the mechanisms described so far are not quite sufficient to ensure that each state machine executes exactly the same commands in the same order. For example, a follower might be unavailable while the leader commits several log entries, then it could be elected leader and overwrite these entries with new ones; as a result, different state machines might execute different command sequences.

This section completes the Raft algorithm by adding a restriction on which servers may be elected leader. The restriction ensures that the leader for any given term contains all of the entries committed in previous terms. Given the election restriction, we then make the rules for commitment more precise. Finally, we present a proof sketch for the ***Leader Completeness Property*** and show how it leads to correct behavior of the replicated state machine.

#### 5.4.1 Election restriction

In any leader-based consensus algorithm, the leader must eventually store all of the committed log entries.

In some consensus algorithms, such as Viewstamped Replication, a leader can be elected even if it doesn’t initially contain all of the committed entries. These algorithms contain additional mechanisms to identify the missing entries and transmit them to the new leader, either during the election process or shortly afterwards. Unfortunately, this results in considerable additional mechanism and complexity. 

Raft uses a simpler approach where it guarantees that all the committed entries from previous terms are present on each new leader from the moment of its election, without the need to transfer those entries to the leader. This means that log entries only flow in one direction, from leaders to followers, and leaders never overwrite existing entries in their logs.

Raft uses the voting process to prevent a candidate from winning an election unless its log contains all committed entries. A candidate must contact a majority of the cluster in order to be elected, which means that every committed entry must be present in at least one of those servers.  If the candidate’s log is at least as up-to-date as any other log in that majority, then it will hold all the committed entries. The RequestVote RPC implements this restriction: the RPC includes information about the candidate’s log, and the voter denies its vote if its own log is more up-to-date than that of the candidate.

Raft determines which of two logs is more up-to-date by comparing the index and term of the last entries in the logs.  If the logs have last entries with different terms, then the log with the later term is more up-to-date. If the logs end with the same term, then whichever log is longer is more up-to-date.

#### 5.4.2 Committing entries from previous terms

a leader knows that an entry from its current term is committed once that entry is stored on a majority of the servers. If a leader crashes before committing an entry, future leaders will attempt to finish replicating the entry. However, a leader cannot immediately conclude that an entry from a previous term is committed once it is stored on a majority of servers. 

![image](https://note.youdao.com/yws/api/personal/file/WEB295332b460f243a197de1e9ccee57a69?method=download&shareKey=b1b6fe97e59fa5e6ee2120b0a0addd4f)

> Figure 8: A time sequence showing why a leader cannot de- termine commitment using log entries from older terms. In (a) S1 is leader and partially replicates the log entry at index 2. In (b) S1 crashes; S5 is elected leader for term 3 with votes from S3, S4, and itself, and accepts a different entry at log index 2. In (c) S5 crashes; S1 restarts, is elected leader, and continues replication. At this point, the log entry from term 2 has been replicated on a majority of the servers, but it is not committed. If S1 crashes as in (d), S5 could be elected leader (with votes from S2, S3, and S4) and overwrite the entry with its own entry from term 3. However, if S1 replicates an en- try from its current term on a majority of the servers before crashing, as in (e), then this entry is committed (S5 cannot win an election). At this point all preceding entries in the log are committed as well.

Figure 8  illustrates a situation where an old log entry is stored on a majority of servers, yet can still be overwritten by a future leader.

To eliminate problems, Raft never commits log entries from previous terms by counting replicas. Only log entries from the leader’s current term are committed by counting replicas; once an entry from the current term has been committed in this way, then all prior entries are committed indirectly because of the Log Matching Property.

Raft incurs this extra complexity in the commitment rules because log entries retain their original term numbers when a leader replicates entries from previous terms. 

#### 5.4.3 Safety argument

Given the complete Raft algorithm, we can now ar- gue more precisely that the Leader Completeness Prop- erty holds. We assume that the Leader Completeness Property does not hold, then we prove a contradiction. Suppose the leader for term T (leaderT) commits a log entry from its term, but that log entry is not stored by the leader of some future term. Consider the smallest term U > T whose leader (leaderU) does not store the entry.


![image](https://note.youdao.com/yws/api/personal/file/WEBf0d7b97ea095a0e0ae267cf3e00ddd34?method=download&shareKey=e6523ac6fe88ab2fe3ac31d8149d7545)

> Figure 9: If S1 (leader for term T) commits a new log entry from its term, and S5 is elected leader for a later term U, then there must be at least one server (S3) that accepted the log entry and also voted for S5.

1. The committed entry must have been absent from leaderU’s log at the time of its election (leaders never delete or overwrite entries).
2. leaderT replicated the entry on a majority of the cluster, and leaderU received votes from a majority of the cluster. Thus, at least one server (“the voter”) both accepted the entry from leaderT and voted for leaderU, as shown in Figure 9. The voter is key to reaching a contradiction.
3. The voter must have accepted the committed entry from leaderT before voting for leaderU; otherwise it would have rejected the AppendEntries request from leaderT (its current term would have been higher than T).
4. The voter still stored the entry when it voted for leaderU, since every intervening leader contained the entry, leaders never remove entries, and followers only remove entries if they conflict with the leader.
5. The voter granted its vote to leaderU, so leaderU’s log must have been as up-to-date as the voter’s. This leads to one of two contradictions.
6. First, if the voter and leaderU shared the same last log term, then leaderU’s log must have been at least as long as the voter’s, so its log contained every entry in the voter’s log. This is a contradiction, since the voter contained the committed entry and leaderU was assumed not to.
7. Otherwise, leaderU’s last log term must have been larger than the voter’s. Moreover, it was larger than T, since the voter’s last log term was at least T (it con- tains the committed entry from term T). The earlier leader that created leaderU’s last log entry must have contained the committed entry in its log (by assump- tion). Then, by the Log Matching Property, leaderU’s log must also contain the committed entry, which is a contradiction.
    8.This completes the contradiction. Thus, the leaders of all terms greater than T must contain all entries from term T that are committed in term T.
    9.The Log Matching Property guarantees that future leaders will also contain entries that are committed indirectly.

> Leader Completeness: If a log entry is committed in a given term, then that entry is committed in a given term, then that entry will be present in the logs of the leaders for all higher-numbered terms.

> State Machine Safety: if a server has applied a log entry at a given index to its state machine, no other server will be apply a different log entry for the same index.

Given the Leader Completeness Property, we can prove the State Machine Safety Property, which states that if a server has applied a log entry at a given index to its state machine, no other server will ever apply a different log entry for the same index. At the time a server applies a log entry to its state machine, its log must be identical to the leader’s log up through that entry and the entry must be committed.  Now consider the lowest term in which any server applies a given log index; the Log Completeness Property guarantees that the leaders for all higher terms will store that same log entry, so servers that apply the index in later terms will apply the same value. Thus, the State Machine Safety Property holds.

Finally, Raft requires servers to apply entries in log index order. Combined with the State Machine Safety Property, this means that all servers will apply exactly the same set of log entries to their state machines, in the same order.

### 5.5 Follower and candidate crashes

If a follower or candidate crashes, then fu- ture RequestVote and AppendEntries RPCs sent to it will fail.  Raft handles these failures by retrying indefinitely; if the crashed server restarts, then the RPC will complete successfully. If a server crashes after completing an RPC but before responding, then it will receive the same RPC again after it restarts. Raft RPCs are idempotent(幂等的), so this causes no harm. 

### 5.6 Timing and availability

One of our requirements for Raft is that safety must not depend on timing: the system must not produce incorrect results just because some event happens more quickly or slowly than expected. However, availability (the ability of the system to respond to clients in a timely manner) must inevitably depend on timing. 

Leader election is the aspect of Raft where timing is most critical. Raft will be able to elect and maintain a steady leader as long as the system satisfies the following timing requirement:

_boradcastTime << electionTimeout << MTBF_

In this inequality _broadcastTime_ is the average time it takes a server to send RPCs in parallel to every server in the cluster and receive their responses; _electionTimeout_ is the election timeout described in Section 5.2; and MTBF is the average time between failures for a single server. 

The broadcast time should be an order of mag- nitude less than the election timeout so that leaders can reliably send the heartbeat messages required to keep fol- lowers from starting elections; given the randomized approach used for election timeouts, this inequality also makes split votes unlikely. 

The election timeout should be a few orders of magnitude less than MTBF so that the system makes steady progress. When the leader crashes, the system will be unavailable for roughly the election timeout; we would like this to represent only a small fraction of overall time.

The broadcast time and MTBF are properties of the un- derlying system, while the election timeout is something we must choose. Raft’s RPCs typically require the recip- ient to persist information to stable storage, so the broad- cast time may range from 0.5ms to 20ms, depending on storage technology. As a result, the election timeout is likely to be somewhere between 10ms and 500ms. Typical server MTBFs are several  months or more, which easily satisfies the timing requirement.