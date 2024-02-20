package timeutil

import (
    "github.com/whencome/gotil"
    "sync"
    "time"
)

type DelayTicker struct {
    C        chan string
    Interval int64    // 单位：秒
    buff     sync.Map // 数据缓存
    ticker   *time.Ticker
}

func NewDelayTicker(interval int64, n int) *DelayTicker {
    if n <= 0 {
        n = 0
    }
    ticker := &DelayTicker{
        C:        make(chan string, n),
        Interval: interval,
        buff:     sync.Map{},
        ticker:   nil,
    }
    go ticker.start()
    return ticker
}

func (t *DelayTicker) start() {
    t.ticker = time.NewTicker(time.Duration(1) * time.Second)
    for {
        curTime := <-t.ticker.C
        now := curTime.Unix()
        t.buff.Range(func(key, value interface{}) bool {
            v := gotil.Int64(value)
            if now-v >= t.Interval {
                t.C <- gotil.String(key)
                t.buff.Delete(key)
            }
            return true
        })
    }
}

func (t *DelayTicker) Add(k string) {
    now := time.Now().Unix()
    t.buff.Store(k, now)
}

func (t *DelayTicker) Stop() {
    t.ticker.Stop()
    close(t.C)
    t.buff = sync.Map{}
}
