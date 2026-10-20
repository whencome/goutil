package pool

import (
    stdlog "log"
    "runtime/debug"
    "sync"
)

// --------------------------- Job ---------------------
type Job interface {
    Do()
}

// runJob 执行任务，并捕获任务执行过程中的panic
// 说明：历史实现中，任何一个任务发生panic都会导致worker协程退出，进而造成任务堆积甚至进程崩溃，
// 此处统一recover并记录日志，保证worker可以继续处理后续任务
func runJob(job Job) {
    if job == nil {
        return
    }
    defer func() {
        if r := recover(); r != nil {
            stdlog.Printf("worker pool: job panic: %v, stack: %s", r, string(debug.Stack()))
        }
    }()
    job.Do()
}

// --------------------------- Worker ---------------------
type Worker struct {
    JobQueue chan Job
}

// NewWorker 创建一个worker
func NewWorker() Worker {
    return Worker{JobQueue: make(chan Job)}
}

// Run 启动worker，worker会把自己的任务通道注册到worker队列中
func (w Worker) Run(wq chan chan Job) {
    go func() {
        for {
            // 注册自身，等待分发器投递任务
            wq <- w.JobQueue
            job, ok := <-w.JobQueue
            if !ok {
                return
            }
            runJob(job)
        }
    }()
}

// --------------------------- WorkerPool ---------------------
type WorkerPool struct {
    size        int
    JobQueue    chan Job
    WorkerQueue chan chan Job

    runOnce sync.Once
}

// NewWorkerPool 创建一个worker池
// workerSize - worker数量
// queueSize - worker注册队列的缓冲大小
// 注意：JobQueue为无缓冲通道，因此Add必须在Run之后调用，否则会阻塞
func NewWorkerPool(workerSize, queueSize int) *WorkerPool {
    if workerSize <= 0 {
        workerSize = 1
    }
    if queueSize < 0 {
        queueSize = 0
    }
    return &WorkerPool{
        size:        workerSize,
        JobQueue:    make(chan Job),
        WorkerQueue: make(chan chan Job, queueSize),
    }
}

// Run 启动worker池
// 说明：此方法可以重复调用，但只有第一次调用会真正启动（内部使用sync.Once保证），
// 避免重复启动导致的协程与任务重复执行问题
func (wp *WorkerPool) Run() {
    wp.runOnce.Do(func() {
        // 初始化worker
        for i := 0; i < wp.size; i++ {
            worker := NewWorker()
            worker.Run(wp.WorkerQueue)
        }
        // 循环获取可用的worker,往worker中写job
        go func() {
            for {
                job, ok := <-wp.JobQueue
                if !ok {
                    return
                }
                worker, ok := <-wp.WorkerQueue
                if !ok {
                    return
                }
                worker <- job
            }
        }()
    })
}

// Add 向worker池中添加任务
// 注意：此方法在worker池未启动时会一直阻塞，请确保先调用Run
func (wp *WorkerPool) Add(job Job) {
    wp.JobQueue <- job
}
