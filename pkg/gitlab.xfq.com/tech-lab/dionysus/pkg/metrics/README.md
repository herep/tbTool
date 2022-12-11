# metrics 

metrics包用来统计服务运行时的资源使用情况及http请求回应等指标信息，对外暴露两类接口metrics和pprof.

## 访问方式 
- 因为metrics请求耗时较长，效率较低，因此安全考虑metrics使用与服务不同的端口, 可使用flag及环境变量进行配置，默认为9120.
- metrics 访问样式 http://{host}:{metrics_port}/metrics
- pprof 访问样式 http://{host}:{metrics_port}/dio/pprof

## 统计指标 
- metrics接口http及go相关指标：
	- 服务的http有关指标：request数量，错误数量，延时，字节数，response字节数。
	- go_gc_duration_seconds：go垃圾回收的时间间隔
	- go_goroutines：goroutine数量
	- go_memstats_xxx：go 内存分配相关指标。
	- go_threads： go线程数
	- process_xxx：进程相关指标，cpu, 文件描述符，虚拟内存，运行时间等。
	- promhttp_metrics_xxx：metrics这个请求的本身的统计指标，如请求次数等

- pprof展示程序底层的资源使用情况，包含多个子目录：
	- cpu（CPU Profiling）: {host}:{metrics_port}/dio/pprof/profile，默认进行 30s 的 CPU Profiling，得到一个分析用的 profile 文件
	- block（Block Profiling）：{host}:{metrics_port}/dio/pprof/block，查看导致阻塞同步的堆栈跟踪
	- goroutine：{host}:{metrics_port}/dio/pprof/goroutine，查看当前所有运行的 goroutines 堆栈跟踪
	- heap（Memory Profiling）: {host}:{metrics_port}/dio/pprof/heap，查看活动对象的内存分配情况
	- mutex（Mutex Profiling）：{host}:{metrics_port}/dio/pprof/mutex，查看导致互斥锁的竞争持有者的堆栈跟踪
	- threadcreate：{host}:{metrics_port}/dio/pprof/threadcreate，查看创建新OS线程的堆栈跟踪


