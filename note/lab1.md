### lab1 part 3

这一部分让我们写一个mapreduce里的schedule，这个函数负责由master.run调用。

schedule负责给worker分配任务，任务的列表由arg给出, 目前空闲的的worker从registerChan取得。

考虑到调用需要in parallel，一个直接的想法是，我们用循环遍历每一个的任务，为每一个任务启动一个goroutine，从registerChan
里取一个worker，并且通过RPC把任务分配给他，然后等待返回。任务结束后，由于这个worker又变回空闲状态，我们把它放回registerChan。
那么怎么保证主进程在所有的go routine完成任务之后才退出呢？在python里我们有p.wait(), 在这里我们可以用sync.WaitGroup: 我们初始化一个
名叫tasks的sync.WaitGroup变量，告诉他总共有n个任务要做。 然后每个go routine 完成一个任务后，就把tasks减一。等到tasks变成0的时候，
我们的主进程就知道它可以退出了。简单来说，sync.WaitGroup就像(或者说就是？)一个可以在进程之间共享的变量。

根据这个思路, 我们实现第一个版本的代码。 

```go
	var tasks sync.WaitGroup
	for i := 0; i < ntasks; i ++ {
		tasks.Add(1) // increase the number of task in queue

		go func(i int) {  // do task in parallel
			rpcAddress :=<- registerChan // work's rpc address
			taskArg := new(DoTaskArgs)
                        ...  // format taskArg here
			ok := call(rpcAddress, "Worker.DoTask", taskArg, nil)
			if ok  {
				registerChan <- rpcAddress

			} else{
				fmt.Printf("Cleanup: RPC %s error\n", rpcAddress)
			}
			tasks.Done() // Once done, decrease the number of task in queue
		}(i)
	}
	tasks.Wait()
	fmt.Printf("Schedule: %v done\n", phase)
```

但是我们发现以上代码并不能通过测试用例。观察log输出，我们发现第19个任务执行完毕之后，程序就stuck住了。 那么为什么最后一个任务没有能成功执行呢？其实，观察日志我们可以发现，最后一个goroutine的`call(rpcAddress, "Worker.DoTask", taskArg, nil)`是被执行了， 但代码被阻塞在了 `registerChan <- rpcAddress` 这一行， 而`tasks.Done()`没有被执行，所以主进程也就没有退出。

为什么往registerChan存消息会阻塞goroutine呢？ 这是因为registerChan是一个无缓冲的信道，它的存消息和取消息都是阻塞的。也就是说, 无缓冲的信道在取消息和存消息的时候都会挂起当前的goroutine，除非另一端已经准备好。 所以当最后一个goroutine往registerChan存入消息时，由于没人将这个消息取走，所以最后一个任务被阻塞住了。而对于前19个任务，因为他们存入registerChan的消息都被后来的进程取走，所以他们都执行完毕并退出了。

参考[Go语言的goroutines、信道和死锁goroutine](https://blog.csdn.net/smilesundream/article/details/80209026）

一个简单的修改办法：


```go
	var tasks sync.WaitGroup
	for i := 0; i < ntasks; i ++ {
		tasks.Add(1) // increase the number of task in queue

		go func(i int) {  // do task in parallel
			rpcAddress :=<- registerChan // work's rpc address
			taskArg := new(DoTaskArgs)
                        ...  // format taskArg here
			ok := call(rpcAddress, "Worker.DoTask", taskArg, nil)
			tasks.Done() // Once done, decrease the number of task in queue
			if ok  {
				registerChan <- rpcAddress

			} else{
				fmt.Printf("Cleanup: RPC %s error\n", rpcAddress)
			}
		}(i)
	}
	tasks.Wait()
	fmt.Printf("Schedule: %v done\n", phase)
```

可以调换tasks.Done()的位置， 就可以通过测试了～
