package cachex

import (
    "sync"
    "testing"
)

// TestGetCacheKey 测试缓存key的拼接规则（不依赖redis）
func TestGetCacheKey(t *testing.T) {
    defer SetCacheKeyPrefix("cachex")
    SetCacheKeyPrefix("prefix")
    if got := getCacheKey("k"); got != "prefix:k" {
        t.Errorf("getCacheKey(k) = %s, want prefix:k", got)
    }
    // NoPrefixKey前缀表示不使用全局前缀
    if got := getCacheKey(NoPrefixKey + "k"); got != "k" {
        t.Errorf("getCacheKey(NoPrefixKey+k) = %s, want k", got)
    }
    // 空前缀
    SetCacheKeyPrefix("")
    if got := getCacheKey("k"); got != "k" {
        t.Errorf("getCacheKey(k) with empty prefix = %s, want k", got)
    }
}

// TestCacheKeyPrefixConcurrency 测试前缀设置与读取的并发安全
// 请配合 go test -race 运行
func TestCacheKeyPrefixConcurrency(t *testing.T) {
    defer SetCacheKeyPrefix("cachex")
    var wg sync.WaitGroup
    wg.Add(2)
    go func() {
        defer wg.Done()
        for i := 0; i < 1000; i++ {
            SetCacheKeyPrefix("p")
        }
    }()
    go func() {
        defer wg.Done()
        for i := 0; i < 1000; i++ {
            _ = getCacheKey("k")
        }
    }()
    wg.Wait()
}
