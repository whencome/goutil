package log

import (
    "fmt"
    "path/filepath"
    "runtime"
    "strings"
    "sync/atomic"

    "github.com/sirupsen/logrus"
    gormlog "gorm.io/gorm/logger"
)

// stdLogger a instance of Logger, this make sure you can directly call log.XXX function
// 使用atomic.Value存储，保证在New/Instance与日志输出并发时的读取安全（避免数据竞争）
var stdLogger atomic.Value

// Fields define a map to store log data
type Fields map[string]interface{}

// DefaultLogger a default logger that implements the Logger interface
type defaultLogger struct {
    *logrus.Entry
}

// Logger define logger interface
type Logger interface {
    // 记录日志信息
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger

    // Entry Print family functions
    Debug(args ...interface{})
    Print(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
    Panic(args ...interface{})

    // Entry Printf family functions
    Debugf(format string, args ...interface{})
    Printf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})
    Panicf(format string, args ...interface{})

    // Entry Println family functions
    Debugln(args ...interface{})
    Println(args ...interface{})
    Infoln(args ...interface{})
    Warnln(args ...interface{})
    Errorln(args ...interface{})
    Fatalln(args ...interface{})
    Panicln(args ...interface{})
}

func init() {
    stdLogger.Store(Instance(&Config{}))
}

// loggerInstance 获取当前的全局logger实例
// 若全局logger尚未初始化（理论上只可能出现在init执行期间），则返回一个默认实例
func loggerInstance() Logger {
    if l, ok := stdLogger.Load().(Logger); ok && l != nil {
        return l
    }
    return Instance(&Config{})
}

// DefaultLogger 创建一个默认的logger
func DefaultLogger() Logger {
    return loggerInstance()
}

// New create a logger
// 此方法会绑定默认logger对象，因此重复调用此方法会导致之前创建的logger对象行为发生改变
// 如果需要使用不同的logger，使用下面的Instance方法
func New(c *Config) Logger {
    logger := Instance(c)
    stdLogger.Store(logger)
    return logger
}

// Instance create a logger instance, this won't bind to the default stdLogger, means that you will get a really new logger
func Instance(c *Config) Logger {
    if c == nil {
        c = defaultConfig
    }
    // create a new logger, do not use the global functions
    logrusLogger := logrus.New()
    logrusLogger.SetOutput(c.GetOutput())
    logrusLogger.SetLevel(logrus.Level(c.GetLogLevel()))
    // 设置日志格式化工具
    if c.Format == "json" {
        formatter := new(logrus.JSONFormatter)
        formatter.TimestampFormat = "2006-01-02 15:04:05"
        logrusLogger.SetFormatter(formatter)
    } else {
        formatter := new(logrus.TextFormatter)
        formatter.TimestampFormat = "2006-01-02 15:04:05"
        formatter.FullTimestamp = true
        logrusLogger.SetFormatter(formatter)
    }
    // create entry
    entry := logrus.NewEntry(logrusLogger)
    // create default logger
    logger := &defaultLogger{
        Entry: entry,
    }
    return logger
}

func (l *defaultLogger) WithField(key string, value interface{}) Logger {
    return &defaultLogger{
        Entry: l.Entry.WithField(key, value),
    }
}

func (l *defaultLogger) WithFields(fields map[string]interface{}) Logger {
    return &defaultLogger{
        Entry: l.Entry.WithFields(fields),
    }
}

// logPackageDir 当前日志包所在的目录，用于在获取调用栈时跳过包内部的调用帧
// 通过运行时信息获取，可兼容不同的模块路径以及Windows风格的路径分隔符
var logPackageDir = func() string {
    _, file, _, ok := runtime.Caller(0)
    if !ok || file == "" {
        return ""
    }
    return filepath.Dir(file)
}()

// inLogPackage 判断给定的文件是否属于日志包
// 说明：统一转换为"/"分隔后再比较，兼容Windows路径以及-trimpath编译后的相对路径
func inLogPackage(file string) bool {
    if file == "" {
        return false
    }
    file = strings.ReplaceAll(file, "\\", "/")
    dir := strings.ReplaceAll(logPackageDir, "\\", "/")
    if dir == "" {
        // 兜底处理
        return strings.Contains(file, "/goutil/log/")
    }
    if file == dir {
        return true
    }
    return strings.HasPrefix(file, dir+"/")
}

// WithStack 增加调用栈信息
// 说明：会跳过日志包内部的调用帧，将第一个包外的调用位置记录到caller字段中
func WithStack(l Logger) Logger {
    if l == nil {
        l = loggerInstance()
    }
    var pc uintptr
    var codePath, prevCodePath, prevFuncName string
    var codeLine, prevCodeLine int
    var ok bool
    for skip := 1; ; skip++ {
        pc, codePath, codeLine, ok = runtime.Caller(skip)
        if !ok {
            // 不ok，函数栈用尽了
            break
        }
        prevCodePath = codePath
        prevCodeLine = codeLine
        if fn := runtime.FuncForPC(pc); fn != nil {
            prevFuncName = fn.Name()
        }
        if !inLogPackage(codePath) {
            // 找到包外的函数了
            break
        }
    }
    if prevFuncName != "" {
        pos := strings.LastIndex(prevFuncName, ".")
        prevFuncName = prevFuncName[pos+1:]
    }
    return l.WithField("caller", fmt.Sprintf("%s::%s:%d", prevCodePath, prevFuncName, prevCodeLine))
}

func WithField(key string, value interface{}) Logger {
    return loggerInstance().WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
    return loggerInstance().WithFields(fields)
}

// Entry Print family functions
func Debug(args ...interface{}) {
    WithStack(loggerInstance()).Debug(args...)
}

func Print(args ...interface{}) {
    WithStack(loggerInstance()).Print(args...)
}

func Info(args ...interface{}) {
    WithStack(loggerInstance()).Info(args...)
}

func Warn(args ...interface{}) {
    WithStack(loggerInstance()).Warn(args...)
}

func Error(args ...interface{}) {
    WithStack(loggerInstance()).Error(args...)
}

func Fatal(args ...interface{}) {
    WithStack(loggerInstance()).Fatal(args...)
}

func Panic(args ...interface{}) {
    WithStack(loggerInstance()).Panic(args...)
}

// Entry Printf family functions
func Debugf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Debugf(format, args...)
}

func Printf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Printf(format, args...)
}

func Infof(format string, args ...interface{}) {
    WithStack(loggerInstance()).Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
    WithStack(loggerInstance()).Panicf(format, args...)
}

// Entry Println family functions
func Debugln(args ...interface{}) {
    WithStack(loggerInstance()).Debugln(args...)
}

func Println(args ...interface{}) {
    WithStack(loggerInstance()).Println(args...)
}

func Infoln(args ...interface{}) {
    WithStack(loggerInstance()).Infoln(args...)
}

func Warnln(args ...interface{}) {
    WithStack(loggerInstance()).Warnln(args...)
}

func Errorln(args ...interface{}) {
    WithStack(loggerInstance()).Errorln(args...)
}

func Fatalln(args ...interface{}) {
    WithStack(loggerInstance()).Fatalln(args...)
}

func Panicln(args ...interface{}) {
    WithStack(loggerInstance()).Panicln(args...)
}

// WithGormLogger 实现gorm日志接口
func WithGormLogger(level gormlog.LogLevel) gormlog.Interface {
    conf := gormlog.Config{
        SlowThreshold:             0,
        Colorful:                  false,
        IgnoreRecordNotFoundError: true,
        ParameterizedQueries:      false,
        LogLevel:                  level,
    }
    return gormlog.New(loggerInstance(), conf)
}

// Log 实现gokratos日志接口
func (l *defaultLogger) Log(level Level, keyvals ...interface{}) error {
    if len(keyvals) == 0 || len(keyvals)%2 != 0 {
        WithStack(l).Warnf("log keyvalues must appear in pairs: %v", keyvals)
        return nil
    }
    fields := Fields{}
    for i := 0; i < len(keyvals); i += 2 {
        fields[fmt.Sprint(keyvals[i])] = keyvals[i+1]
    }
    logger := WithStack(l).WithFields(fields)
    switch level {
    case LevelDebug:
        logger.Debug("")
    case LevelInfo:
        logger.Info("")
    case LevelWarn:
        logger.Warn("")
    case LevelError:
        logger.Error("")
    case LevelFatal:
        logger.Fatal("")
    }
    return nil
}
