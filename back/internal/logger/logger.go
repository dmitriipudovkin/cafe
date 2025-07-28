package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Error(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Fatal(args ...interface{})
}

var logger = logrus.New()

func GetLogger() *logrus.Logger {
	return logger
}
