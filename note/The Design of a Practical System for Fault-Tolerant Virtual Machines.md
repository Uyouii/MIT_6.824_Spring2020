## 3. Practical Implementation of FT

### 3.1 Starting and Restarting FT VMs

This mechanism will also be uesd when re-starting a backup VM after a failuer has occurred.

modified _FT VMotion_ clones a VM to a remote host rather than migrating it. The FT VMotion also sets up a logging channel. Like normal VMotion, FT VMotion typically interrupts the execution of the primary VM by less than a second.

choosing a server on which to run it: Fault-tolerant VMs run in a cluster of servers that have access to shared storage. 

### 3.2 Managing the Logging Channel

maintain a large buffer for logging entries: primary VM produces log entries into the log buffer, backup VM consumes log entries from its log buffer.

the backup sends ack back to the primary each time that it reads some log entries form the network into its log buffer.

![image](https://note.youdao.com/yws/api/personal/file/WEB4d656d6ffccd85a093ce1718242b896d?method=download&shareKey=9087aece3c8e953fbdbaeb23ba0c44ea)

> FT Logging Buffers and Channel

backup VM encounters an empty log buffer when it needs to read the next log entry, it will stop execution until a new log entry is available. If the primary VM encounters a full log buffer when it needs to write a log entry, it must stop execution until log entries can be flushed out. 

in general, the backup VM must be able to replay an execution at roughly the same speed as the primary VM is recording the execution.

if the primary VM fails, the backup VM must "catch up" by replaying all the log entries that it has already acknowledged before it live and starts communicating with eht external world. The time to finish replaying is basically the execution lag time at the point of the failure.

send additional information to determine the real-time execution lag between the primary and backup VMs. Typically the execution lag is less than 100 milliseconds. 

use a slow feedback loop, which will try to gradually pinpoint the appropriate CPU limit for the primary VM that will allow the backup MV to match its execution.

### 3.3 Operation on FT MVs

in general, most operations on the VM should be initiated only on the primary VM. VMware FT then sends any necessary control entry to cause the appropriate change on the backup VM. The onlu operation that can be done independently on the primary and backup VMs is VMotion.

### 3.4 Implementation Issues for Disk IOs

first, simultaneous disk operations that access the same disk location can lead to non-determinism. solution is generally to detect any such IO races and force sucn racing disk operations to execute sequentially in the same way on the primary and backup.

second, a disk operation can also race with a memory access by an application in a VM. e.g. if an app/OS in a VM is reading a memory block at the same time a disk read is occurring to that block. use _bounce buffers_. 

a bounce buffer is a temporary buffer that has the same size as the memory being accessed by a disk operation. A disk read operation is modified to read the specified data to the bounce buffer, and the data is copied to guest memory only as the IO completion is delivered. Similarly, for a disk write operation, the data to be sent is first copied to the bounce buffer, and the disk write is modified to write data from the bounce buffer. 

third, there are some issues associated with disk IOs that are outstanding on the primary when a failure happens, and the backup takes over. There is no way for the newly-promeoted primary VM to be sure if the disk IOs were issued to the disk or completed successfully. re-issue the pending IOs during the go-live process of the backup VM. 

### 3.5 Implementatin Issues for Network IO

The biggest change of the networking emulation code for FT is the disabling of the asynchronous network optimizations.

the elimination of the asynchronous updates of the network device combined with the delaying of sending packets provide some performance challenges for networking.

two approaches to improving VM network performance
1. implement clustering optimizations to reduce VM traps and interrupts.
2. reducing the delay for transmitted packets. the key to reducing the transmit delay is to reduce the time required to send a log message to the backup and get an acknowledgement.
    - ensuring that sending and receving log entries and acknowledgements can all be done without any thread context switch. 

## 4. Design Alternatives 

### 4.1 Shared vs. Non-shared Disk

in default design, the primary and backup VMs share the same virtual disks. theerefore, the content of the shared disks is naturally correct and available if a failover occurs.

### 4.2 Executing Disk Reads on the Backup VM

in default design, the backup VM nerver reads from its virutal disk. Since the disk read is considered an input, it is natural to send the results of the disk read to the backup VM via the logging channel.


