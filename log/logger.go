package log

/**
 * 使用logrus，支持文件输出
 */

// log/logger.go
import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger() {
	Logger.SetOutput(os.Stdout) // 或文件
	Logger.SetLevel(logrus.InfoLevel)
}

func LogInfo(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func LogFailure(err error, msg string) {
	Logger.Errorf("%s: %v", msg, err)
}

//func LogFailure(err error, format string, args ...interface{}) {
//	Logger.Errorf(format, args, err)
//}

func LogFatal(msg string, err error) {
	Logger.Errorf("%s: %v", msg, err)
}
func LogError(msg string, err error) {
	Logger.Errorf("%s: %v", msg, err)
}
func LogErrorMsg(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func LogWarn(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}
