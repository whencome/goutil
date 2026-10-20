package log

import (
	"sync"
	"testing"
)

// TestDebugln 测试Debugln不会触发panic
// 历史实现中Debugln误调用了Panic方法，会在调用时直接panic
func TestDebugln(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Debugln should not panic, got: %v", r)
		}
	}()
	Debugln("debug message")
	Debug("debug message")
	Println("print message")
}

// TestWithStackNil 测试WithStack传入nil时不会panic
func TestWithStackNil(t *testing.T) {
	l := WithStack(nil)
	if l == nil {
		t.Fatal("WithStack should return a logger")
	}
}

// TestNewAndLogConcurrently 测试并发初始化与输出日志时的数据竞争
// 请配合 go test -race 运行
func TestNewAndLogConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			New(&Config{Level: "error", Output: "stdout"})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			Info("concurrent log")
			WithField("k", "v").Infof("concurrent log %d", i)
		}
	}()
	wg.Wait()
}

// TestInstanceWithNilConfig 测试使用nil配置创建logger
func TestInstanceWithNilConfig(t *testing.T) {
	l := Instance(nil)
	if l == nil {
		t.Fatal("Instance(nil) should return a logger")
	}
	l.Info("hello")
}

// TestParseLevel 测试日志等级解析
func TestParseLevel(t *testing.T) {
	cases := map[string]Level{
		"debug":   LevelDebug,
		"INFO":    LevelInfo,
		"warn":    LevelWarn,
		"error":   LevelError,
		"fatal":   LevelFatal,
		"unknown": LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}
