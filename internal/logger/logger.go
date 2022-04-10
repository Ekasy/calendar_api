package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = logrus.InfoLevel
	logger.Formatter = &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	}
	return logger
}
