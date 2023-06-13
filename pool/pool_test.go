package pool

import (
    "fmt"
    "testing"
    "time"
)

// 定义一个Print任务对象
type PrintJob struct {
    Seq int
    Cnt chan<- struct{}
}

func NewPrintJob(ch chan<- struct{}, i int) *PrintJob {
    return &PrintJob{
        Seq: i,
        Cnt: ch,
    }
}

func (j *PrintJob) Do() {
    fmt.Printf("%d - now is %s \n", j.Seq, time.Now().Format("2006-01-02 15:04:05.000"))
    time.Sleep(time.Second * 1)
    j.Cnt <- struct{}{}
}

func TestWorkerPool(t *testing.T) {
    // 定义任务总数（用于测试）
    taskSize := 100
    finished := 0
    // 用于进程阻塞
    ch := make(chan struct{})
    // 用于计数
    cntCh := make(chan struct{})
    // 初始化worker pool并运行
    p := NewWorkerPool(3, 5)
    p.Run()

    // 生产任务
    go func() {
        for i := 0; i < taskSize; i++ {
            job := NewPrintJob(cntCh, i)
            p.JobQueue <- job
        }
    }()

    // 计数并退出任务
    go func() {
        for {
            select {
            case <-cntCh:
                finished++
                if finished >= taskSize {
                    close(ch)
                }
            }
        }
    }()
    <-ch
    t.Logf("executed %d tasks\n", finished)
    t.Log("success")
}
