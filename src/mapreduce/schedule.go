package mapreduce

import (
	"fmt"
	"sync"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	//fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//
	var tasks sync.WaitGroup // wait until all the goroutine has done
	var todo = make(chan int)
	go func() {
		for  i := 0; i < ntasks; i ++ {
			todo <- i
			tasks.Add(1) // increase the number of task in queue
		}
		tasks.Wait()
		close(todo)
	}()

	for i:= range todo {

		go func(i int) {  // do task in parallel
			rpcAddress :=<- registerChan // work's rpc address
			taskArg := new(DoTaskArgs)
			taskArg.Phase = phase
			if phase == mapPhase {
				taskArg.File = mapFiles[i]
			}
			taskArg.JobName = jobName
			taskArg.NumOtherPhase = n_other
			taskArg.TaskNumber = i
			ok := call(rpcAddress, "Worker.DoTask", taskArg, nil)
			if ok {
				tasks.Done() // Once done, decrease the number of task in queueÇ
				registerChan <- rpcAddress
			} else{
				// re-assign the task given to the failed worker to another worker
				todo <- i
				fmt.Printf("Cleanup: RPC %s error\n", rpcAddress)
			}

		}(i)
	}
	tasks.Wait()
	fmt.Printf("Schedule: %v done\n", phase)

}
