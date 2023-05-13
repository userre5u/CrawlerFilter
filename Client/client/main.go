package main

import (
	"context"
	"crawlerDetection/Client/internal"
	"crawlerDetection/Client/s3Service"
	"crawlerDetection/Client/utils"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger, err := utils.GetLogger()
	if err != nil {
		log.Fatal(err)
	}
	dbConn, err := internal.InitDB(config)
	if err != nil {
		logger.Fatal(err)
	}
	sess, err := s3Service.CreateSession(&config)
	if err != nil {
		panic(err)
	}
	context := initContext(logger, sess, s3Service.GetS3(sess), s3Service.GetDownloader(sess), dbConn)
	go internal.Start(context, config)
	<-sigs
	logger.Info("Program is closing resources before exit...")
	dbConn.Close()
	utils.CloseLogger()

}
