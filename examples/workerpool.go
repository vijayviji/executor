package main

import (
	"fmt"
	"github.com/vijayviji/executor"
	"runtime"
	"sync"
)

type acceptorData struct {
	workers executor.Executor
	wg      *sync.WaitGroup
}

func initiatorTask(taskData interface{}, aThreadName string, aTaskID uint64) interface{}  {
	ad := taskData.(*acceptorData)

	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		ad.workers.Submit(func(_ interface{}, wThreadName string, wTaskID uint64) interface{} {
			fmt.Printf("Task %d received from: %s; Executed by %s as task %d\n",
				aTaskID, aThreadName, wThreadName, wTaskID)
			wg.Done()

			return nil
		}, nil)
	}

	wg.Wait()
	ad.wg.Done()

	return nil
}

func main() {
	runtime.GOMAXPROCS(18)
	workers := executor.NewFixedThreadPool("worker", 7, 100)

	var wg sync.WaitGroup
	ad := acceptorData{
		workers: workers,
		wg:      &wg,
	}

	initiators := executor.NewFixedThreadPool("initiator", 7, 1)
	for i := uint32(0); i < 7; i++ {
		initiators.Submit(initiatorTask, &ad)
		wg.Add(1)
	}

	wg.Wait()

	fmt.Println("Exiting!")
}
