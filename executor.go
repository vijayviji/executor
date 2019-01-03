/*
 * Executor package similar to executor in Java language. An executor is backed by thread pool. LockOSThread() is
 * internally used to make the thread concept available for go programmers.
 *
 * This is very useful, if you want to contain resources for certain type of tasks. For example, if you want some
 * CPU hunger tasks not starve other simple tasks, you can create two separate thread pools (or executors) and contain
 * the CPU resource to CPU hunger tasks.
 */

package executor

import (
	"fmt"
	"runtime"
	"strconv"
	"sync/atomic"
)

// TaskFn - The task fn of executor should have this signature.
//
// taskData - Data specific to a task can be passed while submitting tasks.
// threadName - As syscall.gettid() doesn't work in all platforms, this fake threadName is passed to indicate the thread
// taskID - Every task gets a unique id.
type TaskFn func(taskData interface{}, threadName string, taskID uint64) interface{}

// TaskStatus - status of a task
type TaskStatus int

const (
	// TaskNotStarted - Task not started
	TaskNotStarted TaskStatus = iota

	// TaskStarted - Task is running
	TaskStarted TaskStatus = iota

	// TaskDone - Task is done
	TaskDone TaskStatus = iota
)

// This represents a task.
type executorTask struct {
	fn         TaskFn           // function of this task
	returnChan chan interface{} // channel used by the task to communicate the return value to the future.
	taskData   interface{}      // private task data
	taskStatus TaskStatus       // task's current status
}

// An executorThread runs in a goroutine tied to a thread (that thread doesn't run any other
// goroutine). In essence, an executorThread represents a thread. It can execute multiple tasks.
type executorThread struct {
	name      string            // name of this thread
	taskQueue chan executorTask // task queue for this thread
	quit      chan bool         // to signal to the thread to quit
	quitDone  chan bool         // to signal to the executor that thread quitting is done.
}

// Executor - This represents an executor which is backed by multiple threads.
type Executor struct {
	name              string           // name of this executor
	threads           []executorThread // threads backing this executor
	nextThreadToQueue uint64           // next thread to queue the incoming task. RR policy.
}

// Future - While submitting the task to the executor, the user gets a future which can be used to inspect the task
// status and get the return value.
type Future struct {
	task executorTask // task represented by this future
}

// Runs a task.
func (task *executorTask) run(th *executorThread, taskID uint64) {
	task.taskStatus = TaskStarted

	// executorName and taskID combined will give a unique id for every task.
	task.returnChan <- task.fn(task.taskData, th.name, taskID)

	task.taskStatus = TaskDone
}

// The main loop executed by the executor. Listens to the task queue and executes the tasks.
func (th *executorThread) startTaskLoop() {
	runtime.LockOSThread()

	taskID := uint64(0)

	fmt.Println("Starting task loop for thread", th.name)
	for {
		select {
		case task := <-th.taskQueue:
			// this is uint64. no need to wrap.
			taskID++

			task.run(th, taskID)

		case <-th.quit:
			fmt.Println("Quitting thread", th.name)
			th.quitDone <- true
			return
		}
	}
}

// Shutdown - Shuts down an executor (and all the threads backing it)
func (ex *Executor) Shutdown() {
	fmt.Println("Quitting executor", ex.name)

	for _, th := range ex.threads {
		th.quit <- true
	}

	// wait for all the threads to quit
	for _, th := range ex.threads {
		<-th.quitDone
	}

	fmt.Println("Quitting executor", ex.name, "done.")
}

// NewSingleThreadExecutor - Creates an executor backed by single thread
func NewSingleThreadExecutor(name string, qSize uint32) Executor {
	return NewFixedThreadPool(name, 1, qSize)
}

// NewFixedThreadPool - Creates an executor backed by multiple threads.
//
// qSize - Queue size of every thread in the pool.
func NewFixedThreadPool(name string, poolSize uint32, qSize uint32) Executor {
	ex := Executor{
		name:              name,
		nextThreadToQueue: 0,
	}

	for i := uint32(0); i < poolSize; i++ {

		th := executorThread{
			/*
			 * syscall.Gettid() is not defined for mac os. Let's use simple counter and the name to define unique
			 * ID for executor.
			 */
			name: ex.name + "-thread-" + strconv.FormatUint(uint64(i+1), 10),

			taskQueue: make(chan executorTask, qSize),
			quit:      make(chan bool),
			quitDone:  make(chan bool),
		}

		ex.threads = append(ex.threads, th)

		go th.startTaskLoop()
	}

	return ex
}

// Submit - Creates a task out of taskFn and taskData, and submits it to the executor. This returns a future which can
// be used to inspect the result.
// Submit for a single Executor can be called from multiple threads. It's THREAD-SAFE
func (ex *Executor) Submit(taskFn TaskFn, taskData interface{}) Future {
	task := executorTask{
		fn:         taskFn,
		returnChan: make(chan interface{}, 1),
		taskData:   taskData,
		taskStatus: TaskNotStarted,
	}

	val := atomic.AddUint64(&ex.nextThreadToQueue, 1)
	val %= uint64(len(ex.threads))
	ex.threads[val].taskQueue <- task

	future := Future{
		task: task,
	}
	return future
}

// Gets the status of a task
func (task *executorTask) getStatus() TaskStatus {
	return task.taskStatus
}

// Get - Gets the return value of a task this future is pointing to.
func (fut *Future) Get() interface{} {
	result := <-fut.task.returnChan

	return result
}

// GetStatus - Gets the status of the task this future is pointing to.
func (fut *Future) GetStatus() interface{} {
	return fut.task.getStatus()
}
