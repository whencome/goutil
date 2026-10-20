package timeutil

import (
    "sync"
    "time"

    "github.com/whencome/goutil"
)

// DelayTicker 延时触发器
// 说明：往DelayTicker中添加一个key，当该key的存活时间超过Interval（秒）后，会将该key发送到C通道；
// 此对象为并发安全对象，可直接在多个goroutine中使用；请使用NewDelayTicker创建实例。
// 注意：字段C与Interval请勿在运行过程中修改（应对仅在启动前设置）。
type DelayTicker struct {
    C        chan string
    Interval int64    // 单位：秒
    buff     sync.Map // 数据缓存
    ticker   *time.Ticker

    mu        sync.Mutex
    startOnce sync.Once
    stopOnce  sync.Once
    stopCh    chan struct{}
    wg        sync.WaitGroup
}

// NewDelayTicker 创建一个延时触发器
// interval - 延时时间，单位：秒
// n - 结果通道的缓冲大小
func NewDelayTicker(interval int64, n int) *DelayTicker {
    if n <= 0 {
        n = 0
    }
    ticker := &DelayTicker{
        C:        make(chan string, n),
        Interval: interval,
        buff:     sync.Map{},
        ticker:   nil,
        stopCh:   make(chan struct{}),
    }
    ticker.start()
    return ticker
}

// start 启动后台检查协程
// 说明：此方法可以重复调用，但只有第一次调用会真正启动协程（内部使用sync.Once保证）
func (t *DelayTicker) start() {
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

// loop 定时检查缓存中的数据是否已经到期
func (t *DelayTicker) loop() {
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
                if now-v >= t.Interval {
                    t.buff.Delete(key)
                    // 发送时同时监听停止信号，避免Stop时阻塞或向已关闭的通道发送数据
                    select {
                    case t.C <- goutil.String(key):
                    case <-t.stopCh:
                        return false
                    }
                }
                return true
            })
        }
    }
}

// Add 添加一个需要延时处理的key，如果key已存在，则刷新其时间
func (t *DelayTicker) Add(k string) {
    now := time.Now().Unix()
    t.buff.Store(k, now)
}

// Stop 停止触发器并释放资源
// 说明：此方法可以重复调用；停止后会关闭C通道，因此在Stop之后不应再向通道读取数据以外的操作；
// Stop会等待后台协程退出后再关闭通道，不会出现"send on closed channel"的panic
func (t *DelayTicker) Stop() {
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
