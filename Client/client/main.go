package main

import (
	"context"
	"crawlerDetection/Client/internal"
	"crawlerDetection/Client/s3Service"
	"crawlerDetection/Client/utils"

	"database/sql"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func initContext(logger *logrus.Logger, sess *session.Session, s3 *s3.S3, downloader *s3manager.Downloader, db *sql.DB) context.Context {
	prime_object := internal.Global_objects{
		Logger:            logger,
		Object_sess:       sess,
		Object_s3:         s3,
		Object_downloader: downloader,
		DBobject:          db,
	}
	ctx := context.WithValue(context.Background(), internal.Global_objects{}, prime_object)
	return ctx

}

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger := utils.GetLogger()
	dbConn, err := internal.InitDB()
	if err != nil {
		logger.Fatal(err)
	}
	sess, err := s3Service.CreateSession(&config)
	if err != nil {
		panic(err)
	}
	context := initContext(logger, sess, internal.GetS3(sess), internal.GetDownloader(sess), dbConn)
	internal.Start(context, config.SessionKey)
	dbConn.Close()
}
