package log

import (
    "fmt"
    "runtime"
    "strings"
    "sync"

    kratoslog "github.com/go-kratos/kratos/v2/log"
    "github.com/sirupsen/logrus"
    gormlog "gorm.io/gorm/logger"
)

// stdLogger a instance of Logger, this make sure you can directly call log.XXX function
var stdLogger Logger

// a locker
var mu sync.Mutex

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
    stdLogger = New(&Config{})
}

// DefaultLogger 创建一个默认的logger
func DefaultLogger() Logger {
    return stdLogger
}

// New create a logger
// 此方法会绑定默认logger对象，因此重复调用此方法会导致之前创建的logger对象行为发生改变
// 如果需要使用不同的logger，使用下面的Instance方法
func New(c *Config) Logger {
    logger := Instance(c)
    mu.Lock()
    stdLogger = logger
    mu.Unlock()
    return stdLogger
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

// WithStack 增加调用栈信息
func WithStack(l Logger) Logger {
    var pc uintptr
    var codePath, prevCodePath, prevFuncName string
    var codeLine, prevCodeLine int
    var ok bool
    for skip := 1; true; skip++ {
        pc, codePath, codeLine, ok = runtime.Caller(skip)
        if !ok {
            // 不ok，函数栈用尽了
            break
        } else {
            prevCodePath = codePath
            prevCodeLine = codeLine
            prevFuncName = runtime.FuncForPC(pc).Name()
            if !strings.Contains(prevCodePath, "/gotil/log") {
                // 找到包外的函数了
                break
            }
        }
    }
    if prevFuncName != "" {
        pos := strings.LastIndex(prevFuncName, ".")
        prevFuncName = prevFuncName[pos+1:]
    }
    return l.WithField("caller", fmt.Sprintf("%s::%s:%d", prevCodePath, prevFuncName, prevCodeLine))
}

func WithField(key string, value interface{}) Logger {
    return stdLogger.WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
    return stdLogger.WithFields(fields)
}

// Entry Print family functions
func Debug(args ...interface{}) {
    WithStack(stdLogger).Debug(args...)
}

func Print(args ...interface{}) {
    WithStack(stdLogger).Print(args...)
}

func Info(args ...interface{}) {
    WithStack(stdLogger).Info(args...)
}

func Warn(args ...interface{}) {
    WithStack(stdLogger).Warn(args...)
}

func Error(args ...interface{}) {
    WithStack(stdLogger).Error(args...)
}

func Fatal(args ...interface{}) {
    WithStack(stdLogger).Fatal(args...)
}

func Panic(args ...interface{}) {
    WithStack(stdLogger).Panic(args...)
}

// Entry Printf family functions
func Debugf(format string, args ...interface{}) {
    WithStack(stdLogger).Debugf(format, args...)
}

func Printf(format string, args ...interface{}) {
    WithStack(stdLogger).Printf(format, args...)
}

func Infof(format string, args ...interface{}) {
    WithStack(stdLogger).Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
    WithStack(stdLogger).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
    WithStack(stdLogger).Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
    WithStack(stdLogger).Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
    WithStack(stdLogger).Panicf(format, args...)
}

// Entry Println family functions
func Debugln(args ...interface{}) {
    WithStack(stdLogger).Panic(args...)
}

func Println(args ...interface{}) {
    WithStack(stdLogger).Println(args...)
}

func Infoln(args ...interface{}) {
    WithStack(stdLogger).Infoln(args...)
}

func Warnln(args ...interface{}) {
    WithStack(stdLogger).Warnln(args...)
}

func Errorln(args ...interface{}) {
    WithStack(stdLogger).Errorln(args...)
}

func Fatalln(args ...interface{}) {
    WithStack(stdLogger).Fatalln(args...)
}

func Panicln(args ...interface{}) {
    WithStack(stdLogger).Panicln(args...)
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
    return gormlog.New(stdLogger, conf)
}

// Log 实现gokratos日志接口
func (l *defaultLogger) Log(level kratoslog.Level, keyvals ...interface{}) error {
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
    case kratoslog.LevelDebug:
        logger.Debug("")
    case kratoslog.LevelInfo:
        logger.Info("")
    case kratoslog.LevelWarn:
        logger.Warn("")
    case kratoslog.LevelError:
        logger.Error("")
    case kratoslog.LevelFatal:
        logger.Fatal("")
    }
    return nil
}

// KratosLogger 获取一个实现kratos日志接口的日志对象
func KratosLogger() kratoslog.Logger {
    return stdLogger.(*defaultLogger)
}
