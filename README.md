# executor
`Executor` is similar to Executors in Java language. A simple thread pool implemented for golang. This gives you fine control over native threads.

An executor is backed by thread pool. `runtime.LockOSThread()` is
internally used to make the thread concept available for go programmers.
This is very useful, if you want to contain resources for certain type of tasks.

For example, if you want some CPU hunger tasks not starve other simple tasks, you can create two separate thread pools (or executors) and contain
the CPU resource to CPU hunger tasks.

Enjoy!