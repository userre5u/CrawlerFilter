package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

func GetLogger() *logrus.Logger {
	// init logger
	logger := &logrus.Logger{Out: os.Stdout, Level: logrus.DebugLevel, Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
	}}
	return logger

}
