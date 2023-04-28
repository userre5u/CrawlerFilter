package utils

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type Logging struct{}

func HasLogger(ctx context.Context) (*logrus.Logger, bool) {
	// check if underlying type of value is *logrusLogger
	ut := ctx.Value(Logging{})
	logger, ok := ut.(*logrus.Logger)
	if !ok {
		return nil, false
	}
	return logger, true

}

func InitLogger() *logrus.Logger {
	// init the logger
	logger := &logrus.Logger{Out: os.Stdout, Level: logrus.DebugLevel, Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
		PadLevelText:    true,
	}}
	return logger

}
