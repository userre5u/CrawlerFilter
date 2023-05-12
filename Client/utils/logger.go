package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var fd *os.File

func GetLogger() (*logrus.Logger, error) {
	// init logger
	fd, err := os.OpenFile("Client/logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	logger := &logrus.Logger{Out: fd, Level: logrus.DebugLevel, Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
		DisableColors:   true,
	}}
	return logger, nil
}

func CloseLogger() {
	fd.Close()
}
