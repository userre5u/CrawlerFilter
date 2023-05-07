package main

import (
	"crawlerDetection/Client/internal"
	"crawlerDetection/Client/s3Service"
	"crawlerDetection/Client/utils"
)

func main() {
	config, err := utils.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger := utils.GetLogger()
	sess, err := s3Service.CreateSession(&config)
	if err != nil {
		panic(err)
	}
	internal.Start(logger, sess, config.SessionKey)

}
