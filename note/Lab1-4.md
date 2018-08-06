### Lab1 Part4

目标: 在part 3 的基础上加上Failure的处理

想法很直接，如果任务执行失败，我们就把这个任务再交给其他的worker来执行。为了记录哪些是代办的任务（包括尚未执行和之前执行失败的），
我们定义一个名为todo的信道。理想情况下，我们希望不断新建goroutine来执行任务，一直到todo信道没有任务再被添加进去。对此，我们可以
在适当的时候关闭todo信道（表示没有值会被发送了），并且对信道循环v`for i := range todo` 。

那么我们什么时候可以关闭todo信道呢？从前面的分析可以知道，必须要等到没有代办的任务之后，我们才可以关闭它，也就是说，要等到所有
tasks都成功执行才可以。怎么知道所有任务都被执行了呢？ 这就可以用我们在task3中用到的`var tasks sync.WaitGroup`

Code Snippet：

```go
	var tasks sync.WaitGroup // wait until all the goroutine has done
	var todo = make(chan int)
	go func() {
		for  i := 0; i < ntasks; i ++ {
			todo <- i
			tasks.Add(1) // increase the number of task in queue
		}
		tasks.Wait() // If tasks have all be done, we can then close todo channel
		close(todo)
	}()

	for i:= range todo {

		go func(i int) {  // do task in parallel
                    ... 
			ok := call(rpcAddress, "Worker.DoTask", taskArg, nil)
			if ok {
				tasks.Done() // Once done, decrease the number of task in queueÇ
				registerChan <- rpcAddress
			} else{
				// re-assign the task given to the failed worker to another worker
				todo <- i
			}

		}(i)
	}
	tasks.Wait()
```
