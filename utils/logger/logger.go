package utils

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var logging *logrus.Logger

type Formatter struct {
	logrus.TextFormatter
}

func Debug(text ...interface{}) {
	logging.Debug(text)

}

func Info(text ...interface{}) {
	logging.Info(text)

}

func Warning(text ...interface{}) {
	logging.Warn(text)

}

func Error(text ...interface{}) {
	logging.Error(text)

}

func Fatal(text ...interface{}) {
	logging.Fatal(text)

}

func GetLogger() *logrus.Logger {
	fd, err := os.OpenFile("logs/log", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("ERROR opening file: %v", err)
	}
	logging = logrus.New()

	//logging.SetReportCaller(true)
	mw := io.MultiWriter(os.Stdout, fd)
	logging.SetOutput(mw)

	return logging

}
