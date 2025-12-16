// log/logger.go
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

// InitLogger 初始化日志系统
func InitLogger() {
	// 默认输出到 stdout
	Logger.Out = os.Stdout

	// 开发环境：彩色文本格式
	if os.Getenv("ENV") != "production" {
		Logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		// 生产环境：JSON 格式，便于日志系统采集
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	}

	// 日志级别（可通过环境变量控制）
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}

	Logger.Infof("日志系统初始化完成 [Level: %s]", level)
	// 可选：输出到文件
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Logger.Out = file
		} else {
			Logger.Warnf("打开日志文件失败: %v，使用 stdout", err)
		}
	}

	Logger.Infof("日志系统初始化完成 [Level: %s]", Logger.Level.String())
}

// 便捷函数
func LogInfo(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func LogWarn(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func LogError(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func LogFatal(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func LogDebug(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}
