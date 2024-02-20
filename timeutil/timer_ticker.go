package timeutil

import (
    "github.com/whencome/gotil"
    "sync"
    "time"
)

type TimerTicker struct {
    C      chan string
    ticker *time.Ticker
    buff   sync.Map // 数据缓存
    locker sync.RWMutex
}

func NewTimerTicker(n int) *TimerTicker {
    if n < 0 {
        n = 0
    }
    ticker := &TimerTicker{
        C:      make(chan string, n),
        ticker: nil,
        buff:   sync.Map{},
        locker: sync.RWMutex{},
    }
    go ticker.start()
    return ticker
}

func (t *TimerTicker) start() {
    t.ticker = time.NewTicker(time.Duration(1) * time.Second)
    for {
        curTime := <-t.ticker.C
        now := curTime.Unix()
        t.buff.Range(func(key, value interface{}) bool {
            v := gotil.Int64(value)
            if v < now {
                t.buff.Delete(key)
                return true
            }
            if v > now {
                return true
            }
            t.C <- gotil.String(key)
            return true
        })
    }
}

func (t *TimerTicker) Add(k string, v int64) {
    now := time.Now().Unix()
    if v < now {
        return
    }
    t.buff.Store(k, v)
}

func (t *TimerTicker) Stop() {
    t.ticker.Stop()
    close(t.C)
    t.buff = sync.Map{}
}
