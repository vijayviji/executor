package executor

import (
	"strconv"
	"sync"
	"testing"
)

func simpleRunnerFn(taskData interface{}, runnerThreadName string, runnerTaskID uint64) interface{} {
	taskDataStr := taskData.(string)
	//fmt.Println("Doing task", runnerTaskID, "in thread", runnerThreadName)
	return taskDataStr
}

func simpleInitiatorFn(taskData interface{}, initiatorThreadName string, initiatorTaskID uint64) interface{} {
	privateData := taskData.(map[string]interface{})
	runner := privateData["runner"].(*Executor)
	t := privateData["test"].(*testing.T)
	wg := privateData["wg"].(*sync.WaitGroup)
	iterations := privateData["iterationsPerInitiator"].(uint64)

	var fut []Future
	taskDataStrs := make([]string, iterations)
	for i := uint64(0); i < iterations; i++ {
		taskDataStrs[i] = initiatorThreadName + strconv.FormatUint(initiatorTaskID, 10) +
			strconv.FormatUint(i, 10)

		fut = append(fut, runner.Submit(simpleRunnerFn, taskDataStrs[i]))
	}

	for i := uint64(0); i < iterations; i++ {
		output := fut[i].Get().(string)
		if output != taskDataStrs[i] {
			t.Fatal("Output mismatch. Expected " + taskDataStrs[i] + ". Got " + output)
		}
	}

	wg.Done()
	return nil
}

func TestExecutor_NewFixedThreadPool(t *testing.T) {
	const (
		// 1 Million iterations of tasks in a single thread.
		iterationsPerInitiator = uint64(1 * 1000 * 1000)

		// 10 threads in a pool. We've two pools: initiators and runners, each with 10 threads.
		// Every initiator will create 'iterationsPerInitiator' tasks and queue it to runners.
		// So, total tasks done = poolSize * iterationsPerInitiator
		poolSize = uint32(10)
	)

	simpleInitiator := NewFixedThreadPool("TestTaskInitiator", poolSize, 1)
	simpleRunner := NewFixedThreadPool("TestTaskRunner", poolSize, 1000)

	wg := sync.WaitGroup{}
	for i := uint32(0); i < poolSize; i++ {
		wg.Add(1)
		simpleInitiator.Submit(simpleInitiatorFn, map[string]interface{}{
			"runner":                 &simpleRunner,
			"test":                   t,
			"wg":                     &wg,
			"iterationsPerInitiator": iterationsPerInitiator,
		})
	}

	wg.Wait()
	simpleInitiator.Shutdown()
	simpleRunner.Shutdown()
}
