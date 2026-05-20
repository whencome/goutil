package cachex

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gomodule/redigo/redis"
)

var pool = &redis.Pool{
	MaxIdle:   3,
	MaxActive: 10,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", "localhost:6379")
	},
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
			err := LockCall(&result, c, cacheKey, expire, true, mockBizFunc(&counter))
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
	cacheKey := "test_lock_key"
	conn := getTestConn()
	defer conn.Close()
	var count = 1000
	var lockTime int64 = 120
	var succCount = 0
	var failCount = 0
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			ok, err := Lock(conn, cacheKey, lockTime)
			if err != nil {
				t.Errorf("Lock check failed: %v", err)
				failCount++
				return
			}
			if ok {
				succCount++
			} else {
				failCount++
			}
		}()
	}
	wg.Wait()
	// 打印结果
	fmt.Println("All test cases passed.")
	fmt.Printf("Total: %d\n", count)
	fmt.Printf("SUCC: %d\n", succCount)
	fmt.Printf("FAIL: %d\n", failCount)
}

// RUN: go test -v -run TestAddLock
func TestAddLock(t *testing.T) {
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
	result, err := HGet(conn, cacheKey, "nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent field, got nil")
	}
	if result != "" {
		t.Errorf("Expected empty string for non-existent field, got '%s'", result)
	}

	fmt.Println("TestHSetHGet_Integration passed.")
}
