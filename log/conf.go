package log

import (
    "io"
    "os"
    "strings"
    "time"

    rotatelogs "github.com/lestrrat-go/file-rotatelogs"
    "github.com/sirupsen/logrus"
)

// 默认日志配置
var defaultConfig = &Config{
    Level:        "info",
    Output:       "stdout",
    Path:         "",
    RotationTime: "",
    MaxKeepTime:  "",
}

// Config 定义简单的日志配置
type Config struct {
    Level        string `yaml:"level" json:"level" toml:"level"`                         // 日志级别
    Output       string `yaml:"output" json:"output" toml:"output"`                      // 设置输出目标，支持：file,stdout,stderr
    Path         string `yaml:"path" json:"path" toml:"path"`                            // 日志文件路径，包含文件名
    Format       string `yaml:"format" json:"format" toml:"format"`                      // 设置日志格式，支持text和json，默认为：text
    RotationTime string `yaml:"rotation_time" json:"rotation_time" toml:"rotation_time"` // 设置多久切割一次
    MaxKeepTime  string `yaml:"max_keep_time" json:"max_keep_time" toml:"max_keep_time"` // 最大保存时间，超过此时间将被清理
}

// GetRotationTime 获取切割时间，默认24小时切割一次
func (c *Config) GetRotationTime() time.Duration {
    d, e := time.ParseDuration(c.RotationTime)
    if e != nil {
        return time.Duration(24) * time.Hour
    }
    return d
}

// GetMaxKeepTime 获取日志文件最大保存时间，默认30天
func (c *Config) GetMaxKeepTime() time.Duration {
    d, e := time.ParseDuration(c.MaxKeepTime)
    if e != nil {
        return time.Duration(30*24) * time.Hour
    }
    return d
}

// GetLogLevel 获取日志等级
func (c *Config) GetLogLevel() uint32 {
    logLevel, err := logrus.ParseLevel(c.Level)
    if err != nil {
        logLevel = logrus.InfoLevel
    }
    return uint32(logLevel)
}

// GetOutput 获取日志输出目标
func (c *Config) GetOutput() io.Writer {
    output := strings.TrimSpace(strings.ToLower(c.Output))
    switch output {
    case "stdout":
        return os.Stdout
    case "stderr":
        return os.Stderr
    case "file":
        /*
        	日志轮转相关函数
        	`WithLinkName` 为最新的日志建立软连接
        	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
        	 WithMaxAge 和 WithRotationCount二者只能设置一个
        	`WithMaxAge` 设置文件清理前的最长保存时间
        	`WithRotationCount` 设置文件清理前最多保存的个数
        */
        // 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
        writer, err := rotatelogs.New(
            c.Path+".%Y%m%d%H%M",
            rotatelogs.WithLinkName(c.Path),
            rotatelogs.WithMaxAge(c.GetMaxKeepTime()),
            rotatelogs.WithRotationTime(c.GetRotationTime()),
        )
        if err != nil {
            // if fail, return os.Stdout as default
            return os.Stdout
        }
        return writer
    default:
        return os.Stdout
    }
}
