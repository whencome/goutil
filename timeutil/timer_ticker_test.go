package timeutil

import (
    "fmt"
    "testing"
    "time"
)

// TestTimerTicker 测试定时触发器的基本功能
func TestTimerTicker(t *testing.T) {
    tt := NewTimerTicker(10)
    defer tt.Stop()

    now := time.Now().Unix()
    expectedTime := now + 2
    key := fmt.Sprintf("data_at_%d", expectedTime)
    tt.Add(key, expectedTime)

    select {
    case got := <-tt.C:
        if got != key {
            t.Fatalf("unexpected key: got %s, expect %s", got, key)
        }
        // 触发时间与实际时间可能有1秒的误差（受检查时机影响）
        diff := time.Now().Unix() - expectedTime
        if diff < 0 || diff > 1 {
            t.Fatalf("triggered at wrong time, diff = %d", diff)
        }
    case <-time.After(6 * time.Second):
        t.Fatal("wait timer ticker result timeout")
    }
}

// TestTimerTickerExpired 测试已过期任务不会被触发
func TestTimerTickerExpired(t *testing.T) {
    tt := NewTimerTicker(1)
    defer tt.Stop()

    tt.Add("expired", time.Now().Unix()-10)
    select {
    case k := <-tt.C:
        t.Fatalf("expired task should not be triggered, got %s", k)
    case <-time.After(2 * time.Second):
        // 正常，过期任务不会被触发
    }
}

// TestTimerTickerConcurrency 测试并发添加任务
func TestTimerTickerConcurrency(t *testing.T) {
    const taskNum = 100
    tt := NewTimerTicker(taskNum)
    defer tt.Stop()

    now := time.Now().Unix()
    expected := make(map[string]struct{}, taskNum)
    for i := 0; i < taskNum; i++ {
        // 触发时间分布在2~4秒之后
        triggerAt := now + int64(i%3) + 2
        key := fmt.Sprintf("data_of_%d_%d", triggerAt, i)
        expected[key] = struct{}{}
        tt.Add(key, triggerAt)
    }

    timeout := time.After(12 * time.Second)
    got := 0
    for got < taskNum {
        select {
        case k, ok := <-tt.C:
            if !ok {
                t.Fatalf("channel closed unexpectedly, got %d/%d", got, taskNum)
            }
            if _, exists := expected[k]; !exists {
                t.Fatalf("unexpected key: %s", k)
            }
            delete(expected, k)
            got++
        case <-timeout:
            t.Fatalf("timeout, got %d/%d tasks", got, taskNum)
        }
    }
}

// TestTimerTickerStop 测试停止操作的幂等性与安全性
func TestTimerTickerStop(t *testing.T) {
    tt := NewTimerTicker(1)
    tt.Add("k1", time.Now().Unix()+1)
    tt.Stop()
    // 重复调用不应该panic
    tt.Stop()
    // 停止后通道最终应被关闭（关闭前可能还存在已写入的结果）
    deadline := time.After(3 * time.Second)
    for {
        select {
        case _, ok := <-tt.C:
            if !ok {
                return
            }
        case <-deadline:
            t.Fatal("channel not closed after stop")
        }
    }
}
