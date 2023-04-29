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

func displayContent(Crawler bool) string {
	if Crawler {
		return "Something went wrong during your request :("
	} else {
		return "WOW"
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
	pCriticalW := make([]byte, 80)
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
	fmt.Printf("Data len: %d + Input len: %d = %d (object=%#v)\n", len(data), len(input), len(input)+len(data), object)
	if len(input)+len(data) > bucket.MaxDataLen {
		fmt.Println("Creating new file...")
		data = data[:0]
		object = bucket.CreateNewObject()
		fmt.Printf("New file: %#v\n", object)
		return true
	}
	fmt.Printf("Not creating new file - using: %#v\n", object)
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

func Handler(ctx context.Context, e events.LambdaFunctionURLRequest) (string, error) {
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

	return displayContent(reqInfo.Crawler), nil
	//return events.LambdaFunctionURLResponse{StatusCode: 200, Body: displayContent(reqInfo.Bot)}, nil
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
	data = make([]byte, 0, 2048)
	logger := utils.InitLogger()
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

	subcontext := initContext(logger, s3)
	lambda.Start(func(ctx context.Context, e events.LambdaFunctionURLRequest) (string, error) {
		return Handler(subcontext, e)
	})

}
