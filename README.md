# executor
`Executor` is similar to Executors in Java language. A simple thread pool implemented for golang. This gives you fine control over native threads.

An executor is backed by thread pool. `runtime.LockOSThread()` is
internally used to make the thread concept available for go programmers.
This is very useful, if you want to contain resources for certain type of tasks.

For example, if you want some CPU hunger tasks not starve other simple tasks, you can create two separate thread pools (or executors) and contain
the CPU resource to CPU hunger tasks.

Enjoy!

## APIs
* To import
    ```go
    import "github.com/vijayviji/executor"
    ```
* Create an executor backed by a single thread:
   ```go
   ex := executor.NewSingleThreadExecutor("executorName", 200)
   // 200 = Task Queue Size
   ```

* Create an executor backed by a pool of threads:
    ```go
    ex := executor.NewFixedThreadPool("executorName", 10, 2000)
    // 10 = No. of threads
    // 2000 = Task Queue Size of each thread
    // All the tasks queued to this executor will be pushed to task queues of all the threads backing this executor in
    // a round-robin fashion.
    ```

* Queue a task to an executor:
    ```go
    dataForTask = "Dummy data"

    future := ex.Submit(func(taskData interface{}, threadName string, taskID uint64) interface{} {
        dataFromTask := taskData.(string)
        fmt.Println("data for this task: ", dataFromTask)
        return "OKKK"
    }, dataForTask)
    :
    :
    result := future.Get()
    // result will be "OKKK"

    // taskStatus can be any of executor.TaskNotStarted, executor.TaskStarted, executor.TaskDone
    taskStatus := future.GetStatus()

* Shutdown an executor (and all its backing threads):
    ```go
    ex.Shutdown()
    // This function would only return after all the threads are shut down.
    ```

* Running test on executor.go
    ```make
    make test
    ```
