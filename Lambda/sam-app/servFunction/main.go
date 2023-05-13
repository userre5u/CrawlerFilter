package main

import (
	"bytes"
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

type Global_objects struct {
	logger   *logrus.Logger
	s3Object *s3.S3
}

func displayContent(Crawler bool) (string, int) {
	if Crawler {
		return "Forbidden", 403
	}
	return "Good", 200

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
		`{"DateTime": %q, "Session": %q, "IP": %q, "IpType": %q, "UA": %q, "Method": %q, "Country": %q, "SessionKey": %q, "Path": %q, "CriticalWords": %s, "Crawler": %t}`,
		reqinfo.DateTime, reqinfo.Session, reqinfo.IP, reqinfo.IpType, reqinfo.UA, reqinfo.Method, reqinfo.Country,
		reqinfo.SessionKey, reqinfo.Path, string(pCriticalW), reqinfo.Crawler,
	)

}

func checkNewFile(input string) {
	if len(input)+len(data) > utils.MaxDataLen {
		data = data[:0]
		object = bucket.CreateNewObject()
	}
}

func parseReqInfo(input string) ([]byte, error) {
	var formatMsg internal.ReqInfo
	err := json.Unmarshal([]byte(input), &formatMsg)
	if err != nil {
		return nil, err
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(formatMsg)

	return reqBodyBytes.Bytes(), nil

}

func saveData(input string, s3object *s3.S3) error {
	checkNewFile(input)
	formatMsg, err := parseReqInfo(input)
	if err != nil {
		return err
	}
	data = append(data, formatMsg...)
	bucket.PutS3(s3object, data, object.String())
	return nil

}

func Handler(globalObject Global_objects, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

	reqInfo := internal.ReqInfo{
		DateTime:      "",
		Session:       "",
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
	store := runChecker(globalObject.logger, &reqInfo, e)
	err := saveData(store, globalObject.s3Object)
	if err != nil {
		globalObject.logger.Error(err)
		reqInfo.Crawler = true
	}
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

func initContext(log *logrus.Logger, s3 *s3.S3) context.Context {
	prime_object := Global_objects{logger: log, s3Object: s3}
	return context.WithValue(context.Background(), Global_objects{}, prime_object)
}

func main() {
	err := utils.LoadEnv()
	if err != nil {
		panic(err)
	}
	data = make([]byte, 0, 2048)
	ctx := initContext(utils.InitLogger(), initS3())
	globalObject := ctx.Value(Global_objects{}).(Global_objects)
	lambda.Start(func(ctx context.Context, e events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		return Handler(globalObject, e)
	})

}
