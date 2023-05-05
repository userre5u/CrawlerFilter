package main

import (
	"context"
	"encoding/json"
	"fmt"
	"servFunction/bucket"
	"servFunction/internal"
	"servFunction/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var data []byte
var object uuid.UUID

func displayContent(Crawler bool) (string, int) {
	if Crawler {
		return "Forbidden", 403
	} else {
		return "Good", 200
	}
}

func displayCriticalWords(pCriticalW *[]byte, r *internal.ReqInfo) (err error) {
	(*pCriticalW), err = json.Marshal(r.CriticalWords)
	if err != nil {
		(*pCriticalW) = []byte(" ")
	}
	return

}

func runChecker(logger *logrus.Logger, reqinfo *internal.ReqInfo, e events.LambdaFunctionURLRequest) string {
	pCriticalW := make([]byte, 120)
	logger.Info("http request sanity [STARTED]")
	// Check #1
	utils.DisplayMsg(logger, reqinfo.GetIP(e))
	// Check #2
	country, err := reqinfo.Getcountry(e)
	utils.DisplayMsg(logger, country, err)
	// Check #3
	msg, retCode := reqinfo.Getmethod(e)
	utils.DisplayMsg(logger, msg, retCode)
	// Check #4
	msg, retCode = reqinfo.GetPath(e)
	utils.DisplayMsg(logger, msg, retCode)
	// check #5
	msg, retCode = reqinfo.GetAgent(e)
	utils.DisplayMsg(logger, msg, retCode)
	// Check #6
	msg, retCode = reqinfo.GetSessionKey(e)
	utils.DisplayMsg(logger, msg, retCode)
	// check #7
	msg, retCode = reqinfo.GetBody(e)
	utils.DisplayMsg(logger, msg, retCode)
	logger.Info("http request sanity [COMPLETED]")
	err = displayCriticalWords(&pCriticalW, reqinfo)
	if err != nil {
		logger.Error(err)
	}
	reqinfo.SetSession(e)
	reqinfo.SetDateTime(e)

	return fmt.Sprintf(
		"Date: %s, Session: %s, IP: %s, IpType: %s, user-agent: %s, Method: %s, Country: %s, SessionKey: %s, Path: %s, CriticalWords: %s, Crawler: %t\n",
		reqinfo.DateTime, reqinfo.Session, reqinfo.IP, reqinfo.IpType, reqinfo.UA, reqinfo.Method, reqinfo.Country,
		reqinfo.SessionKey, reqinfo.Path, string(pCriticalW), reqinfo.Crawler,
	)

}

func checkNewFile(input string) bool {
	if len(input)+len(data) > utils.MaxDataLen {
		data = data[:0]
		object = bucket.CreateNewObject()
		return true
	}
	return false
}

func saveData(input string, s3object *s3.S3) {
	bNewFile := checkNewFile(input)
	data = append(data, input...)
	if bNewFile {
		bucket.PutS3(s3object, data, object.String())
		return
	}
	bucket.PutS3(s3object, data, object.String())

}

func Handler(ctx context.Context, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	logger, ok := utils.HasLogger(ctx)
	if !ok {
		panic("Could not extract logger")
	}

	reqInfo := internal.ReqInfo{
		IP:            "",
		Crawler:       false,
		IpType:        "",
		UA:            "",
		Country:       "",
		SessionKey:    "",
		Path:          "",
		Method:        "",
		CriticalWords: map[string]bool{"malware": false, "botnet": false, "Crawler": false, "zombie": false, "zeus": false},
	}
	store := runChecker(logger, &reqInfo, e)
	s3object, ok := ctx.Value(bucket.S3Bucket{}).(*s3.S3)
	if !ok {
		panic("Could not extract s3 object")
	}

	saveData(store, s3object)
	msg, status_code := displayContent(reqInfo.Crawler)
	return events.LambdaFunctionURLResponse{StatusCode: status_code, Body: msg}, nil
}

func initS3() *s3.S3 {
	s3, err := bucket.CreateSession()
	if err != nil {
		panic(err)
	}
	if !bucket.BucketExists(s3) {
		if err := bucket.CreateS3(s3); err != nil {
			panic(err)
		}
	}
	object = bucket.CreateNewObject()
	return s3
}

func initContext(logger *logrus.Logger, s3 *s3.S3) context.Context {
	// store logger into context
	sLogging := utils.Logging{}
	subContext := context.WithValue(context.Background(), sLogging, logger)
	subContext.Value(sLogging)
	// store s3 object into context
	s3object := bucket.S3Bucket{}
	subContext = context.WithValue(subContext, s3object, s3)

	return subContext
}

func main() {
	err := utils.LoadEnv()
	if err != nil {
		panic(err)
	}
	data = make([]byte, 0, 2048)
	subcontext := initContext(utils.InitLogger(), initS3())
	lambda.Start(func(ctx context.Context, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		return Handler(subcontext, e)
	})

}
