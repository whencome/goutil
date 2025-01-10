package cachex

import (
    "fmt"
    "strings"

    "github.com/gomodule/redigo/redis"
    "github.com/whencome/goutil"
    "github.com/whencome/goutil/jsonkit"
)

const (
    // NoPrefixKey 无前缀key，在缓存key前面加上此前缀，则会忽略设置的全局前缀key
    NoPrefixKey = "!<:"
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
    // 使用“!<:”表示不需要前缀，一般用于多服务共享缓存的情形
    if strings.HasPrefix(k, NoPrefixKey) {
        return k[len(NoPrefixKey):]
    }
    // 添加前缀
    if cacheKeyPrefix != "" {
        k = cacheKeyPrefix + ":" + k
    }
    return k
}

// Exists 判断指定key是否存在
func Exists(rds redis.Conn, cacheKey string) (bool, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Bool(rds.Do("EXISTS", cacheKey))
}

func PExists(provider Provider, cacheKey string) (bool, error) {
    rds := provider.Redis()
    defer rds.Close()
    return Exists(rds, cacheKey)
}

// Call 带有缓存的调用，会将业务方法结果进行缓存
func Call(ret interface{}, rds redis.Conn, cacheKey string, expire int64, bf BizFunc) error {
    cacheKey = getCacheKey(cacheKey)
    // get data from cache
    cacheData, err := redis.Bytes(rds.Do("GET", cacheKey))
    if err == nil {
        err = jsonkit.Unmarshal(cacheData, ret)
        if err == nil {
            return nil
        }
    }

    // get data by call business func
    data, err := bf()
    if err != nil {
        return err
    }
    if goutil.IsNil(data) {
        return nil
    }
    bytesData, err := jsonkit.Marshal(data)
    if err != nil {
        return err
    }
    // 赋值
    if ret != nil {
        err = jsonkit.Unmarshal(bytesData, ret)
        if err != nil {
            return err
        }
    }

    // 缓存数据(暂时忽略错误)
    if expire > 0 {
        _, _ = rds.Do("SETEX", cacheKey, expire, string(bytesData))
    } else {
        _, _ = rds.Do("SETEX", cacheKey, string(bytesData))
    }
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
    cacheKey = getCacheKey(cacheKey)
    // 1. 先检查锁是否存在，如果不存在再尝试获取锁
    // 1.1 检查锁是否存在
    isExists, err := redis.Bool(rds.Do("EXISTS", cacheKey))
    if err != nil {
        return err
    }
    if isExists {
        return fmt.Errorf("你的手速有点快，喝口水再试试吧~")
    }
    // 1.2 尝试获取锁
    lockRs, err := rds.Do("SET", cacheKey, "1", "EX", expire, "NX")
    if err != nil {
        return err
    }
    if lockRs == nil {
        // 抢占失败
        return fmt.Errorf("请求失败，请稍后再试")
    }

    // 2. 执行业务调用
    resp, err := bf()
    if err != nil {
        return err
    }
    if goutil.IsNil(resp) {
        ret = nil
        return nil
    }
    data, err := jsonkit.Marshal(resp)
    if err != nil {
        return err
    }
    // 赋值
    if ret != nil {
        err = jsonkit.Unmarshal(data, ret)
        if err != nil {
            return err
        }
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
    bytesData, err := jsonkit.Marshal(data)
    if err != nil {
        return err
    }
    if expire > 0 {
        _, err = rds.Do("SETEX", cacheKey, expire, string(bytesData))
    } else {
        _, err = rds.Do("SET", cacheKey, string(bytesData))
    }
    return err
}

// StoreMany 缓存多个值
func StoreMany(rds redis.Conn, expire int64, data map[string]interface{}) error {
    if len(data) == 0 {
        return nil
    }
    for cacheKey, cacheData := range data {
        err := Store(rds, cacheKey, expire, cacheData)
        if err != nil {
            return err
        }
    }
    return nil
}

// PStore 直接缓存结果
func PStore(provider Provider, cacheKey string, expire int64, data interface{}) error {
    rds := provider.Redis()
    defer rds.Close()
    return Store(rds, cacheKey, expire, data)
}

// PStoreMany 缓存多个值
func PStoreMany(provider Provider, expire int64, data map[string]interface{}) error {
    rds := provider.Redis()
    defer rds.Close()
    return StoreMany(rds, expire, data)
}

// Fetch 从缓存中取值
func Fetch(ret interface{}, rds redis.Conn, cacheKey string) error {
    cacheKey = getCacheKey(cacheKey)
    bytesData, err := redis.Bytes(rds.Do("GET", cacheKey))
    if err != nil {
        if err == redis.ErrNil {
            return nil
        }
        return err
    }
    return jsonkit.Unmarshal(bytesData, ret)
}

// PFetch 从缓存中取值
func PFetch(ret interface{}, provider Provider, cacheKey string) error {
    rds := provider.Redis()
    defer rds.Close()
    return Fetch(ret, rds, cacheKey)
}

// Incr 对指定的key的数值加1
func Incr(rds redis.Conn, cacheKey string) (int64, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Int64(rds.Do("INCR", cacheKey))
}

func PIncr(provider Provider, cacheKey string) (int64, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Int64(provider.Redis().Do("INCR", cacheKey))
}

// Decr 对指定的key的数值减1
func Decr(rds redis.Conn, cacheKey string) (int64, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Int64(rds.Do("INCR", cacheKey))
}

func PDecr(provider Provider, cacheKey string) (int64, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Int64(provider.Redis().Do("INCR", cacheKey))
}

// SetBit 设置或清除指定偏移量上的位(bit)。位的设置或清除取决于 value，可以是 0 或者是 1 。
// 当expire为大于0的时候，表示需要定时过期
func SetBit(rds redis.Conn, cacheKey string, expire int64, offset int64, v int) error {
    cacheKey = getCacheKey(cacheKey)
    _, err := rds.Do("SETBIT", cacheKey, offset, v)
    if err != nil {
        return err
    }
    // 如果指定了过期时间，则需要进行设置
    // 过期时间单位为秒
    if expire > 0 {
        _, _ = rds.Do("EXPIRE", cacheKey, expire)
    }
    return err
}

func PSetBit(provider Provider, cacheKey string, expire int64, offset int64, v int) error {
    rds := provider.Redis()
    defer rds.Close()
    return SetBit(rds, cacheKey, expire, offset, v)
}

// GetBit 获取指定位的值
func GetBit(rds redis.Conn, cacheKey string, offset int64) (int, error) {
    cacheKey = getCacheKey(cacheKey)
    return redis.Int(rds.Do("GETBIT", cacheKey, offset))
}

func PGetBit(provider Provider, cacheKey string, offset int64) (int, error) {
    rds := provider.Redis()
    defer rds.Close()
    return GetBit(rds, cacheKey, offset)
}
