package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logging struct{}

func DisplayMsg(l *logrus.Logger, msg string, typeCode ...interface{}) {
	if typeCode != nil {
		switch underValue := typeCode[0].(type) {
		case bool:
			if underValue {
				l.Info(msg)
			} else {
				l.Warn(msg)
			}
		case error:
			if underValue == nil {
				l.Info(msg)
			} else {
				l.Error(underValue)
			}
		}
		return
	}
	l.Info(msg)

}

func InitLogger() *logrus.Logger {
	// init the logger
	logger := &logrus.Logger{Out: os.Stdout, Level: logrus.DebugLevel, Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
		DisableColors:   true,
	}}
	return logger

}
