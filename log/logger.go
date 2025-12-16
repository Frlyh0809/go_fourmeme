// log/logger.go
package log

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger 初始化日志系统：控制台单份 + 按级别分文件
func InitLogger() {
	Logger = logrus.New()

	// 创建日志目录（默认 logs，或通过环境变量 LOG_DIR 指定）
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 打开各级别日志文件（仅文件，不含 Stdout）
	infoFile, _ := os.OpenFile(filepath.Join(logDir, "info.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	warnFile, _ := os.OpenFile(filepath.Join(logDir, "warn.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	errorFile, _ := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	debugFile, _ := os.OpenFile(filepath.Join(logDir, "debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	// 文件专用 formatter（纯文本，无颜色）
	fileFormatter := &logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	// 开发环境：控制台彩色/纯文本
	if os.Getenv("ENV") != "production" {
		Logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		// 生产环境：控制台 JSON
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	}

	// 控制台输出（唯一途径）
	Logger.SetOutput(os.Stdout)

	// 日志级别
	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// Hook 只写文件（不包含 Stdout，避免重复）
	Logger.AddHook(&LevelFileHook{
		InfoWriter:  infoFile,
		WarnWriter:  warnFile,
		ErrorWriter: errorFile,
		DebugWriter: debugFile,
		Formatter:   fileFormatter,
	})

	Logger.Infof("日志系统初始化完成 [Level: %s] [Dir: %s]", Logger.Level.String(), logDir)
}

// LevelFileHook 只负责写文件
type LevelFileHook struct {
	InfoWriter  io.Writer
	WarnWriter  io.Writer
	ErrorWriter io.Writer
	DebugWriter io.Writer
	Formatter   logrus.Formatter
}

func (hook *LevelFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *LevelFileHook) Fire(entry *logrus.Entry) error {
	if hook.Formatter == nil {
		return nil
	}

	line, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}

	var writer io.Writer
	switch entry.Level {
	case logrus.InfoLevel:
		writer = hook.InfoWriter
	case logrus.WarnLevel:
		writer = hook.WarnWriter
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		writer = hook.ErrorWriter
	case logrus.DebugLevel:
		writer = hook.DebugWriter
	}

	if writer != nil {
		_, _ = writer.Write(line)
	}
	return nil
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
