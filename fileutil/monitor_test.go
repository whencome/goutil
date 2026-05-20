package fileutil

import (
    "os"
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
    // 用于计数
    count := 0
    chExit := make(chan struct{})
    monitor.Watch(f, func(f string) error {
        count++
        t.Logf("-- count = %d\n", count)
        return nil
    }, false)
    monitor.Run()
    // 更改文件
    go func() {
        time.Sleep(time.Second)
        err := os.WriteFile(f, []byte("11111\n"), os.ModePerm)
        if err != nil {
            t.Logf("create test file fail: %v\n", err)
            t.Fail()
        }
        time.Sleep(time.Second)
        err = os.WriteFile(f, []byte("22222\n"), os.ModeAppend)
        if err != nil {
            t.Logf("write test file fail: %v\n", err)
            t.Fail()
        }
        time.Sleep(time.Second)
        err = os.WriteFile(f, []byte("33333\n"), os.ModeAppend)
        if err != nil {
            t.Logf("write test file fail: %v\n", err)
            t.Fail()
        }
        time.Sleep(time.Second)
        chExit <- struct{}{}
    }()
    <-chExit
    if count != 3 {
        t.Logf("count not match, expect 3, got %d\n", count)
        t.Fail()
    }
}
