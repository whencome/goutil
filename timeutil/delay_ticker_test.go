package timeutil

import (
    "testing"
    "time"
)

// TestDelayTicker 测试延时触发器的基本功能
func TestDelayTicker(t *testing.T) {
    dt := NewDelayTicker(1, 10)
    defer dt.Stop()

    dt.Add("k1")
    select {
    case k := <-dt.C:
        if k != "k1" {
            t.Fatalf("unexpected key: %s", k)
        }
    case <-time.After(5 * time.Second):
        t.Fatal("wait delay ticker result timeout")
    }
}

// TestDelayTickerStop 测试停止操作的安全性（重复停止、停止后不再触发）
func TestDelayTickerStop(t *testing.T) {
    dt := NewDelayTicker(1, 1)
    dt.Add("k1")
    dt.Stop()
    // 重复停止不应panic
    dt.Stop()

    deadline := time.After(3 * time.Second)
    for {
        select {
        case _, ok := <-dt.C:
            if !ok {
                return
            }
        case <-deadline:
            t.Fatal("channel not closed after stop")
        }
    }
}
