package utils

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var fd *os.File

func GetLogger() (*logrus.Logger, error) {
	// init logger
	fd, err := os.OpenFile("Client/logs/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	logger := &logrus.Logger{
		Out:   io.MultiWriter(fd, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%\n",
		}}
	return logger, nil
}

func CloseLogger() {
	// Close fd before program exit
	fd.Close()
}
