package mapreduce

import (
	"fmt"
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

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)
	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//

	finishJob := make(chan bool, ntasks)
	jobQueue := make(chan int)

	go func() {
		for i := 0; i < ntasks; i++ {
			jobQueue <- i
		}
	}()

	// worker goroutine once finished either queue worker for new job or give task to anotehr worker
	work := func(worker string, index int) {
		var resp bool
		switch phase {
		case mapPhase:
			resp = call(worker, "Worker.DoTask", DoTaskArgs{JobName: jobName, Phase: phase, TaskNumber: index, NumOtherPhase: n_other, File: mapFiles[index]}, nil)
		case reducePhase:
			resp = call(worker, "Worker.DoTask", DoTaskArgs{JobName: jobName, Phase: phase, TaskNumber: index, NumOtherPhase: n_other}, nil)
		}

		if resp {
			finishJob <- true
			registerChan <- worker
		} else {
			jobQueue <- index
		}
	}

	end := false
	jobFinished := 0
	for !end {
		select {
		case job := <-jobQueue:
			worker := <-registerChan
			go work(worker, job)
		case worker := <-registerChan:
			select {
			case job := <-jobQueue:
				go work(worker, job)
			case finished := <-finishJob:
				finishJob <- finished
			}
		case <-finishJob:
			jobFinished++
			if jobFinished == ntasks {
				end = true
			}
		}
	}

	fmt.Printf("Schedule: %v done\n", phase)
}
