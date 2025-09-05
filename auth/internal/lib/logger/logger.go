package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger = logrus.Logger

func NewLogger() *logrus.Logger {
	return logrus.New()
}
