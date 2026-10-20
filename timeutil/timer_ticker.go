package timeutil

import (
    "sync"
    "time"

    "github.com/whencome/goutil"
)

// TimerTicker 定时触发器
// 说明：往TimerTicker中添加一个key与指定的触发时间，当到达该时间（精确到秒）时，会将该key发送到C通道；
// 此对象为并发安全对象，可直接在多个goroutine中使用；请使用NewTimerTicker创建实例。
type TimerTicker struct {
    C      chan string
    ticker *time.Ticker
    buff   sync.Map // 数据缓存

    mu        sync.Mutex
    startOnce sync.Once
    stopOnce  sync.Once
    stopCh    chan struct{}
    wg        sync.WaitGroup
}

// NewTimerTicker 创建一个定时触发器
// n - 结果通道的缓冲大小
func NewTimerTicker(n int) *TimerTicker {
    if n < 0 {
        n = 0
    }
    ticker := &TimerTicker{
        C:      make(chan string, n),
        ticker: nil,
        buff:   sync.Map{},
        stopCh: make(chan struct{}),
    }
    ticker.start()
    return ticker
}

// start 启动后台检查协程
// 说明：此方法可以重复调用，但只有第一次调用会真正启动协程（内部使用sync.Once保证）
func (t *TimerTicker) start() {
    t.startOnce.Do(func() {
        t.mu.Lock()
        if t.stopCh == nil {
            t.stopCh = make(chan struct{})
        }
        if t.ticker == nil {
            t.ticker = time.NewTicker(time.Duration(1) * time.Second)
        }
        t.mu.Unlock()
        t.wg.Add(1)
        go t.loop()
    })
}

// loop 定时检查缓存中的数据是否到达触发时间
func (t *TimerTicker) loop() {
    defer t.wg.Done()
    for {
        select {
        case <-t.stopCh:
            return
        case curTime, ok := <-t.ticker.C:
            if !ok {
                return
            }
            now := curTime.Unix()
            t.buff.Range(func(key, value interface{}) bool {
                v := goutil.Int64(value)
                if v < now {
                    // 已过期，直接清理
                    t.buff.Delete(key)
                    return true
                }
                if v > now {
                    // 还未到触发时间
                    return true
                }
                // 到达触发时间，删除后发送
                t.buff.Delete(key)
                // 发送时同时监听停止信号，避免Stop时阻塞或向已关闭的通道发送数据
                select {
                case t.C <- goutil.String(key):
                case <-t.stopCh:
                    return false
                }
                return true
            })
        }
    }
}

// Add 添加一个定时任务
// k - 任务标识，v - 触发时间（unix时间戳，单位：秒）
// 说明：若给定的触发时间已经过去，则此任务会被忽略
func (t *TimerTicker) Add(k string, v int64) {
    now := time.Now().Unix()
    if v < now {
        return
    }
    t.buff.Store(k, v)
}

// Stop 停止触发器并释放资源
// 说明：此方法可以重复调用；停止后会关闭C通道，因此在Stop之后不应再向通道读取数据以外的操作；
// Stop会等待后台协程退出后再关闭通道，不会出现"send on closed channel"的panic
func (t *TimerTicker) Stop() {
    t.stopOnce.Do(func() {
        t.mu.Lock()
        stopCh := t.stopCh
        ticker := t.ticker
        t.mu.Unlock()
        if stopCh == nil {
            // 未启动，直接关闭结果通道
            if t.C != nil {
                close(t.C)
            }
            return
        }
        close(stopCh)
        if ticker != nil {
            ticker.Stop()
        }
        // 等待后台协程退出后关闭结果通道
        t.wg.Wait()
        // 清理缓存数据
        t.buff.Range(func(key, _ interface{}) bool {
            t.buff.Delete(key)
            return true
        })
        if t.C != nil {
            close(t.C)
        }
    })
}
