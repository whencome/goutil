package cachex

import (
    "encoding/json"
    "fmt"
    "github.com/gomodule/redigo/redis"
)

// cacheKeyPrefix 设置缓存key的前缀
var cacheKeyPrefix = "cachex"

// Provider 定义redis数据提供对象接口
type Provider interface {
    Redis() redis.Conn
}

// BizFunc 定义业务方法，用于在无缓存状态下获取数据
type BizFunc func() (interface{}, error)

// SetCacheKeyPrefix 设置缓存key的前缀
func SetCacheKeyPrefix(prefix string) {
    cacheKeyPrefix = prefix
}

// getCacheKey 获取一个完整的cache key
func getCacheKey(k string) string {
    if cacheKeyPrefix != "" {
        k = cacheKeyPrefix + ":" + k
    }
    return k
}

// Call 带有缓存的调用，会将业务方法结果进行缓存
func Call(ret interface{}, rds redis.Conn, cacheKey string, expire int64, bf BizFunc) error {
    cacheKey = getCacheKey(cacheKey)
    // get data from cache
    cacheData, err := redis.Bytes(rds.Do("GET", cacheKey))
    if err == nil {
        err = json.Unmarshal(cacheData, ret)
        if err == nil {
            return nil
        }
    }

    // get data by call business func
    data, err := bf()
    if err != nil {
        return err
    }
    bytesData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    // 赋值
    err = json.Unmarshal(bytesData, ret)
    if err != nil {
        return err
    }

    // 缓存数据(暂时忽略错误)
    _, _ = rds.Do("SETEX", cacheKey, expire, string(bytesData))
    return nil
}

// PCall 使用provider作为redis连接提供方，并自行关闭连接，内部仍旧调用Call，其它Pxxx的方法同理
func PCall(ret interface{}, provider Provider, cacheKey string, expire int64, bf BizFunc) error {
    rds := provider.Redis()
    defer rds.Close()
    return Call(ret, rds, cacheKey, expire, bf)
}

// ResetCall 重置调用，会忽略之前的缓存（先清除之前的缓存），然后在调用业务接口并缓存结果
func ResetCall(ret interface{}, rds redis.Conn, cacheKey string, expire int64, bf BizFunc) error {
    Remove(rds, cacheKey)
    return Call(ret, rds, cacheKey, expire, bf)
}

// PResetCall 重置调用，会忽略之前的缓存（先清除之前的缓存），然后在调用业务接口并缓存结果
func PResetCall(ret interface{}, provider Provider, cacheKey string, expire int64, bf BizFunc) error {
    rds := provider.Redis()
    defer rds.Close()
    Remove(rds, cacheKey)
    return Call(ret, rds, cacheKey, expire, bf)
}

// LockCall 带有锁的调用，当存在并发调用可能时，先寻求获得cacheKey对应的锁，如果获取失败，则无法执行
// autoUnlock - 标记是否自动释放锁（仅限请求成功时），如果不自动释放，则需要等待锁自动过期释放，此参数可以用于防止重复提交等场景
func LockCall(ret interface{}, rds redis.Conn, cacheKey string, expire int64, autoUnlock bool, bf BizFunc) error {
    // 1. 先检查锁是否存在，如果不存在再尝试获取锁
    // 1.1 检查锁是否存在
    isExists, err := redis.Bool(rds.Do("EXISTS", cacheKey))
    if err != nil {
        return err
    }
    if isExists {
        return fmt.Errorf("锁[%s]已经存在", cacheKey)
    }
    // 1.2 尝试获取锁
    lockRs, err := rds.Do("SET", cacheKey, "1", "EX", expire, "NX")
    if err != nil {
        return err
    }
    if lockRs == nil {
        // 抢占失败
        return fmt.Errorf("获取锁[%s]失败", cacheKey)
    }

    // 2. 执行业务调用
    resp, err := bf()
    if err != nil {
        return err
    }
    data, err := json.Marshal(resp)
    if err != nil {
        return err
    }
    // 赋值
    err = json.Unmarshal(data, ret)
    if err != nil {
        return err
    }

    // 3. 业务结束,释放锁
    // 如果是失败不做处理，等待自行过期释放
    if autoUnlock {
        _, _ = rds.Do("DEL", cacheKey)
    }

    // 4. 返回结果
    return nil
}

// PLockCall 带有锁的调用
func PLockCall(ret interface{}, provider Provider, cacheKey string, expire int64, autoUnlock bool, bf BizFunc) error {
    rds := provider.Redis()
    defer rds.Close()
    return LockCall(ret, rds, cacheKey, expire, autoUnlock, bf)
}

// Remove 移除cacheKey对应的缓存
func Remove(rds redis.Conn, cacheKey string) {
    cacheKey = getCacheKey(cacheKey)
    _, _ = rds.Do("DEL", cacheKey)
}

// PRemove 移除cacheKey对应的缓存
func PRemove(provider Provider, cacheKey string) {
    rds := provider.Redis()
    defer rds.Close()
    Remove(rds, cacheKey)
}

// RemoveBatch 批量数据清除
func RemoveBatch(rds redis.Conn, cacheKeys []string) {
    if len(cacheKeys) == 0 {
        return
    }
    for _, cacheKey := range cacheKeys {
        Remove(rds, cacheKey)
    }
}

// PRemoveBatch 批量数据清除
func PRemoveBatch(provider Provider, cacheKeys []string) {
    rds := provider.Redis()
    defer rds.Close()
    RemoveBatch(rds, cacheKeys)
}

// Store 直接缓存结果
func Store(rds redis.Conn, cacheKey string, expire int64, data interface{}) error {
    cacheKey = getCacheKey(cacheKey)
    bytesData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    _, err = rds.Do("SETEX", cacheKey, expire, string(bytesData))
    return err
}

// PStore 直接缓存结果
func PStore(provider Provider, cacheKey string, expire int64, data interface{}) error {
    rds := provider.Redis()
    defer rds.Close()
    return Store(rds, cacheKey, expire, data)
}

// Fetch 从缓存中取值
func Fetch(ret interface{}, rds redis.Conn, cacheKey string) error {
    cacheKey = getCacheKey(cacheKey)
    bytesData, err := redis.Bytes(rds.Do("GET", cacheKey))
    if err != nil {
        return err
    }
    return json.Unmarshal(bytesData, ret)
}

// PFetch 从缓存中取值
func PFetch(ret interface{}, provider Provider, cacheKey string) error {
    rds := provider.Redis()
    defer rds.Close()
    return Fetch(ret, rds, cacheKey)
}
