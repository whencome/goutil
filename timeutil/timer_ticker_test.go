package timeutil

import (
    "fmt"
    "math/rand"
    "strconv"
    "strings"
    "testing"
    "time"
)

func TestTimerTicker(t *testing.T) {
    finish := make(chan bool)

    // create timer ticker
    tt := NewTimerTicker(10)
    go tt.start()

    // add consumer
    go func() {
        for {
            select {
            case k := <-tt.C:
                now := time.Now().Unix()
                fmt.Printf("get %s at %d\n", k, now)
                parts := strings.Split(k, "_")
                t1, e := strconv.Atoi(parts[2])
                if e != nil {
                    fmt.Printf("convert time failed: %s\n", e)
                    finish <- false
                    break
                }
                expectedTime := int64(t1)
                if expectedTime != now {
                    fmt.Printf("time not match")
                    finish <- false
                    break
                }
                finish <- true
            }
            break
        }
        fmt.Printf("consumer exit...\n")
    }()

    now := time.Now().Unix()
    expectedTime := now + 5
    k := fmt.Sprintf("data_at_%d", expectedTime)
    tt.Add(k, expectedTime)
    fmt.Printf("expect to get %s at %d\n", k, expectedTime)

    // wait result
    v := <-finish
    fmt.Printf("test rs: %+v\n", v)
    if !v {
        t.Fail()
    }
}

func TestTimerTickerConcurrency(t *testing.T) {
    finish := make(chan bool)

    // create timer ticker
    tt := NewTimerTicker(10)
    go tt.start()

    // add consumer
    go func() {
        for {
            select {
            case k := <-tt.C:
                now := time.Now().Unix()
                fmt.Printf("get %s at %d\n", k, now)
                parts := strings.Split(k, "_")
                t1, e := strconv.Atoi(parts[2])
                if e != nil {
                    fmt.Printf("convert time failed: %s\n", e)
                    finish <- false
                    break
                }
                expectedTime := int64(t1) / 1e9
                if expectedTime != now {
                    fmt.Printf("time not match[got %d, expect %d]\n", expectedTime, now)
                    finish <- false
                    break
                }
                finish <- true
            }
        }
        fmt.Printf("consumer exit...\n")
    }()

    maxNum := 1000
    failCnt := 0
    succCnt := 0
    startTime := time.Now()
    go func() {
        for i := 0; i < maxNum; i++ {
            now := time.Now()
            expectedTime := now.Add(time.Second * time.Duration(int64(rand.Intn(300)) + 300))
            k := fmt.Sprintf("data_of_%d", expectedTime.UnixNano())
            tt.Add(k, expectedTime.Unix())
            fmt.Printf("expect to get %s at %d\n", k, expectedTime.Unix())
        }
    }()

    // wait result
    getNum := 0
    for {
        v := <-finish
        getNum++
        if v {
            succCnt++
        } else {
            failCnt++
        }
        if getNum >= maxNum {
            break
        }
    }
    endTime := time.Now()
    fmt.Printf("\n--------------------\n")
    fmt.Printf("Total: %d\n", maxNum)
    fmt.Printf("Succ: %d\n", succCnt)
    fmt.Printf("Fail: %d\n", failCnt)
    fmt.Printf("Time Cost: %.4f s\n", float64(endTime.Sub(startTime)) / 1e9)
    if failCnt > 0 {
        t.Fail()
    }
}