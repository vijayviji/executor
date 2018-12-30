package main

import (
	"fmt"
	"runtime"
	"sync"
	"github.com/vijayviji/executor"
)

type AcceptorData struct {
	workers Executor
	wg      *sync.WaitGroup
}

func initiatorTask(data interface{}, aThreadID int, aTaskID uint64) {
	ad := data.(AcceptorData)

	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		ad.workers.submit(func(_ interface{}, wThreadID int, wTaskID uint64) {
			fmt.Printf("Task received from: initiator-%d-%d; Executed by: worker-%d-%d\n",
				aThreadID, aTaskID, wThreadID, wTaskID)
			wg.Done()
		})
	}

	wg.Wait()
	ad.wg.Done()
}

func main() {
	initiators := executor.Executor{}
	workers := executor.Executor{}

	runtime.GOMAXPROCS(18)
	workers.newFixedThreadPool(7, 100, nil)

	var wg sync.WaitGroup
	ad := AcceptorData{
		workers: workers,
		wg:      &wg,
	}

	initiators.newFixedThreadPool(7, 1, ad)
	for i := uint32(0); i < 7; i++ {
		initiators.submit(initiatorTask)
		wg.Add(1)
	}

	wg.Wait()

	fmt.Println("Exiting!")
}
