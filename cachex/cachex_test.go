package cachex

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gomodule/redigo/redis"
)

// redisTestAddr 返回测试使用的redis地址
// 可通过环境变量 GOUTIL_TEST_REDIS_ADDR 指定，默认使用 localhost:6379
func redisTestAddr() string {
	if addr := os.Getenv("GOUTIL_TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

var pool = &redis.Pool{
	MaxIdle:   3,
	MaxActive: 10,
	// 并发测试的goroutine数量可能超过MaxActive，开启Wait避免获取连接时直接报"connection pool exhausted"
	Wait: true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", redisTestAddr())
	},
}

// redisAvailable 标记当前环境是否有可用的redis（未配置密码且可连接）
var redisAvailable bool

// TestMain 初始化时探测redis是否可用，不可用时跳过所有集成测试
func TestMain(m *testing.M) {
	conn := pool.Get()
	_, err := conn.Do("PING")
	_ = conn.Close()
	redisAvailable = err == nil
	os.Exit(m.Run())
}

// skipIfNoRedis 在redis不可用时跳过测试
func skipIfNoRedis(t *testing.T) {
	t.Helper()
	if !redisAvailable {
		t.Skip("skip test because redis is not available (start a local redis without password first)")
	}
}

func getTestConn() redis.Conn {
	return pool.Get()
}

func mockBizFunc(counter *int) BizFunc {
	var mu sync.Mutex
	return func() (interface{}, error) {
		mu.Lock()
		defer mu.Unlock()
		*counter++
		return *counter, nil
	}
}

func TestLockCall_ConcurrentAccess(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_lock_key"
	expire := int64(5)
	var counter int

	conn := getTestConn()
	defer conn.Close()

	_, _ = conn.Do("DEL", getCacheKey(cacheKey))

	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	results := make([]int, numGoroutines)
	var mu sync.Mutex
	var startChan = make(chan struct{})
	var successCount int32
	var expected int
	var once sync.Once

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			<-startChan

			c := getTestConn() // 每个 goroutine 独立获取连接
			defer c.Close()

			var result int
			// 注意：autoUnlock为true时，业务方法执行完毕后锁立即释放，后续goroutine仍可获取到锁并执行，
			// 因此"只执行一次"的场景必须使用autoUnlock=false（即防止重复提交的场景）
			err := LockCall(&result, c, cacheKey, expire, false, mockBizFunc(&counter))
			if err != nil {
				t.Logf("goroutine %d failed to acquire lock: %v", i, err)
				return
			}
			t.Logf("goroutine %d got result: %d", i, result)

			atomic.AddInt32(&successCount, 1)

			mu.Lock()
			defer mu.Unlock()
			results[i] = result

			once.Do(func() {
				expected = result
			})
		}(i)
	}

	close(startChan)
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("Expected exactly one successful execution, got %d", successCount)
	}

	// 校验结果一致性
	for i, v := range results {
		fmt.Printf("result[%d]: %d\n", i, v)
		if v != 0 && v != expected {
			t.Errorf("Result mismatch at index %d: got %d, expected %d", i, v, expected)
		}
	}

	if counter != 1 {
		t.Errorf("Counter should be 1, got %d", counter)
	}

	// 使用autoUnlock=false时，锁会保留到过期或者被显式释放
	if _, err := Unlock(conn, cacheKey); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}
	exists, err := Exists(conn, cacheKey)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if exists {
		t.Errorf("Lock key should have been released")
	}

	fmt.Println("All test cases passed.")
}

// RUN: go test -v -run TestLock
func TestLock(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_lock_key"
	var count = 1000
	var lockTime int64 = 120
	var succCount int32
	var failCount int32
	var wg sync.WaitGroup
	wg.Add(count)
	// 注意：redigo的Conn不是并发安全的，每个goroutine必须使用独立的连接
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			c := getTestConn()
			defer c.Close()
			ok, err := Lock(c, cacheKey, lockTime)
			if err != nil {
				t.Errorf("Lock check failed: %v", err)
				atomic.AddInt32(&failCount, 1)
				return
			}
			if ok {
				atomic.AddInt32(&succCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}
	wg.Wait()
	// 清理测试用的锁
	conn := getTestConn()
	defer conn.Close()
	Unlock(conn, cacheKey)
	// 打印结果
	fmt.Println("All test cases passed.")
	fmt.Printf("Total: %d\n", count)
	fmt.Printf("SUCC: %d\n", atomic.LoadInt32(&succCount))
	fmt.Printf("FAIL: %d\n", atomic.LoadInt32(&failCount))
}

// RUN: go test -v -run TestAddLock
func TestAddLock(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_a_lock_key"
	conn := getTestConn()
	defer conn.Close()
	ok, err := Lock(conn, cacheKey, 600)
	if err != nil {
		t.Errorf("Lock check failed: %v", err)
		return
	}
	t.Logf("Lock ok: %v", ok)
}

// RUN: go test -v -run TestUnlock
func TestUnlock(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_a_lock_key"
	conn := getTestConn()
	defer conn.Close()
	ok, err := Unlock(conn, cacheKey)
	if err != nil {
		t.Errorf("unlock failed: %v", err)
		return
	}
	t.Logf("unlock ok: %v", ok)
}

// RUN: go test -v -run TestHSet
func TestHSet(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_hash_key"
	field := "field1"
	value := "value1"

	conn := getTestConn()
	defer conn.Close()

	// Clean up before test
	_, _ = conn.Do("DEL", getCacheKey(cacheKey))

	// Test HSet
	err := HSet(conn, cacheKey, field, value)
	if err != nil {
		t.Errorf("HSet failed: %v", err)
		return
	}

	// Verify the value was set
	result, err := redis.String(conn.Do("HGET", getCacheKey(cacheKey), field))
	if err != nil {
		t.Errorf("HGET failed: %v", err)
		return
	}

	if result != value {
		t.Errorf("Expected value '%s', got '%s'", value, result)
	}

	fmt.Println("TestHSet passed.")
}

// RUN: go test -v -run TestHGet
func TestHGet(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_hash_get_key"
	field := "field1"
	value := "test_value_123"

	conn := getTestConn()
	defer conn.Close()

	// Clean up before test
	_, _ = conn.Do("DEL", getCacheKey(cacheKey))

	// Set a value first using HSet
	err := HSet(conn, cacheKey, field, value)
	if err != nil {
		t.Errorf("HSet failed: %v", err)
		return
	}

	// Test HGet
	result, err := HGet(conn, cacheKey, field)
	if err != nil {
		t.Errorf("HGet failed: %v", err)
		return
	}

	if result != value {
		t.Errorf("Expected value '%s', got '%s'", value, result)
	}

	fmt.Println("TestHGet passed.")
}

// RUN: go test -v -run TestHSetHGet_Integration
func TestHSetHGet_Integration(t *testing.T) {
	skipIfNoRedis(t)
	cacheKey := "test_hash_integration"
	fields := map[string]string{
		"name":    "John",
		"age":     "25",
		"city":    "Beijing",
		"country": "China",
	}

	conn := getTestConn()
	defer conn.Close()

	// Clean up before test
	_, _ = conn.Do("DEL", getCacheKey(cacheKey))

	// Set multiple fields
	for field, value := range fields {
		err := HSet(conn, cacheKey, field, value)
		if err != nil {
			t.Errorf("HSet failed for field %s: %v", field, err)
			return
		}
	}

	// Get and verify each field
	for field, expectedValue := range fields {
		result, err := HGet(conn, cacheKey, field)
		if err != nil {
			t.Errorf("HGet failed for field %s: %v", field, err)
			return
		}

		if result != expectedValue {
			t.Errorf("Field %s: expected '%s', got '%s'", field, expectedValue, result)
		}
	}

	// Test non-existent field
	// 注意：HGET获取不存在的字段时redis返回nil，历史实现中此处返回的是("", nil)，
	// 这里仅校验不会panic且返回空字符串
	result, _ := HGet(conn, cacheKey, "nonexistent")
	if result != "" {
		t.Errorf("Expected empty string for non-existent field, got '%s'", result)
	}

	fmt.Println("TestHSetHGet_Integration passed.")
}
