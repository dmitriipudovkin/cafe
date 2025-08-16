package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger = logrus.Logger

var logger = logrus.New()

func GetLogger() *logrus.Logger {
	return logger
}
