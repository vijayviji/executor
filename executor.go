package executor

import (
	"runtime"
	"syscall"
)

type ExecutorTask func(interface{}, int, uint64)

type Executor struct {
	taskQueue chan ExecutorTask
	data      interface{}
	threadCnt uint32
}

func taskLoop(ex *Executor) {
	runtime.LockOSThread()

	taskID := uint64(0)
	threadID := syscall.Gettid()

	for task := range ex.taskQueue {
		// this is uint64. no need to wrap.
		taskID += 1

		// threadID and taskID combined to give a unique id for every task.
		task(ex.data, threadID, taskID)
	}
}

func (ex *Executor) newSingleThreadExecutor(qSize uint32, data interface{}) {
	ex.newFixedThreadPool(1, qSize, data)
}

func (ex *Executor) newFixedThreadPool(poolSize uint32, qSize uint32, data interface{}) {
	ex.data = data
	ex.taskQueue = make(chan ExecutorTask, qSize)

	for i := uint32(0); i < poolSize; i++ {
		go taskLoop(ex)
	}
}

func (ex *Executor) submit(task ExecutorTask) {
	ex.taskQueue <- task
}
