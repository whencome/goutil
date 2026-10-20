package fileutil

import (
    "os"
    "sync/atomic"
    "testing"
    "time"
)

// go test -v -run ^TestFileMonitor_Watch$
func TestFileMonitor_Watch(t *testing.T) {
    // create file
    f := "/tmp/test.txt"
    // 退出时删除临时文件
    defer func() {
        if Exists(f) {
            os.Remove(f)
        }
    }()
    // 创建文件监控器
    monitor := NewFileMonitor()
    defer monitor.Close()
    // 设置去抖时间，避免同一次写入产生多个事件时回调被重复执行
    monitor.SetDebounce(300 * time.Millisecond)
    // 用于计数（回调在监控协程中执行，需使用原子操作）
    var count int32
    chExit := make(chan struct{})
    if err := monitor.Watch(f, func(f string) error {
        atomic.AddInt32(&count, 1)
        t.Logf("-- count = %d\n", atomic.LoadInt32(&count))
        return nil
    }, false); err != nil {
        t.Fatalf("watch file failed: %v", err)
    }
    monitor.Run()
    // 更改文件
    go func() {
        time.Sleep(time.Second)
        // 每次写入之间间隔超过去抖时间，保证每次写入都能触发回调
        for i := 0; i < 3; i++ {
            err := os.WriteFile(f, []byte("content\n"), os.ModePerm)
            if err != nil {
                t.Logf("write test file fail: %v\n", err)
                t.Fail()
            }
            time.Sleep(500 * time.Millisecond)
        }
        chExit <- struct{}{}
    }()
    <-chExit
    // 等待最后一次回调执行完成
    time.Sleep(500 * time.Millisecond)
    if got := atomic.LoadInt32(&count); got != 3 {
        t.Logf("count not match, expect 3, got %d\n", got)
        t.Fail()
    }
}

// TestFileMonitorCloseTwice 测试重复关闭不会panic
func TestFileMonitorCloseTwice(t *testing.T) {
    monitor := NewFileMonitor()
    if err := monitor.Close(); err != nil {
        t.Fatalf("close an unstarted monitor failed: %v", err)
    }
    monitor.Run()
    if err := monitor.Close(); err != nil {
        t.Fatalf("close monitor failed: %v", err)
    }
    // 重复调用Run不应该启动多个协程/不panic
    monitor.Run()
    _ = monitor.Close()
}
