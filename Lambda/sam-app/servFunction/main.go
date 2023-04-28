package main

import (
	"context"
	"encoding/json"
	"servFunction/internal"
	"servFunction/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

func displayContent(bot bool) string {
	if bot {
		return "Something went wrong during your request :("
	} else {
		return "WOW"
	}
}

func displayMsg(l *logrus.Logger, msg string, typeCode ...interface{}) {
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

func displayCriticalWords(pCriticalW *[]byte, r *internal.ReqInfo) (err error) {
	(*pCriticalW), err = json.Marshal(r.CriticalWords)
	if err != nil {
		(*pCriticalW) = []byte(" ")
	}
	return

}

// use info events.LambdaFunctionURLRequest instead of EventInfo struct
func Handler(ctx context.Context, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	pCriticalW := make([]byte, 80)
	reqInfo := internal.ReqInfo{
		IP:            "",
		Bot:           false,
		IpType:        "",
		UA:            "",
		Country:       "",
		SessionKey:    "",
		Path:          "",
		Method:        "",
		CriticalWords: map[string]bool{"malware": false, "botnet": false, "bot": false, "zombie": false, "zeus": false},
	}
	logger, ok := utils.HasLogger(ctx)
	if !ok {
		panic("Could not extract logger")
	} else {
		logger.Info("http request sanity [STARTED]")
		// Check #1
		displayMsg(logger, reqInfo.GetIP(e))
		// Check #2
		country, err := reqInfo.Getcountry(e)
		displayMsg(logger, country, err)
		// Check #3
		msg, retCode := reqInfo.Getmethod(e)
		displayMsg(logger, msg, retCode)
		// Check #4
		msg, retCode = reqInfo.GetPath(e)
		displayMsg(logger, msg, retCode)
		// check #5
		msg, retCode = reqInfo.GetAgent(e)
		displayMsg(logger, msg, retCode)
		// Check #6
		msg, retCode = reqInfo.GetSessionKey(e)
		displayMsg(logger, msg, retCode)
		// check #7
		msg, retCode = reqInfo.GetBody(e)
		displayMsg(logger, msg, retCode)
		logger.Info("http request sanity [COMPLETED]")
		err = displayCriticalWords(&pCriticalW, &reqInfo)
		if err != nil {
			logger.Error(err)
		}
		logger.Infof(
			"IP: %s, IpType: %s, user-agent: %s, Method: %s, Country: %s, SessionKey: %s, Path: %s, CriticalWords: %s, Bot: %t\n",
			reqInfo.IP, reqInfo.IpType, reqInfo.UA, reqInfo.Method, reqInfo.Country, reqInfo.SessionKey, reqInfo.Path, string(pCriticalW), reqInfo.Bot,
		)
	}
	return events.LambdaFunctionURLResponse{StatusCode: 200, Body: displayContent(reqInfo.Bot)}, nil
}

func main() {
	sLogging := utils.Logging{}
	logger := utils.InitLogger()
	child := context.WithValue(context.Background(), sLogging, logger)
	child.Value(sLogging)
	lambda.Start(func(ctx context.Context, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		return Handler(child, e)
	})

}
