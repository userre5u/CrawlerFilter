package bucket

import (
	"fmt"
	"servFunction/utils"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func CreateSession() (*s3.S3, error) {
	creds := credentials.NewStaticCredentials(utils.DataConfig.Aws_access_key_id, utils.DataConfig.Aws_secret_access_key, utils.DataConfig.Token)
	session, err := session.NewSession(&aws.Config{Region: aws.String(utils.DataConfig.Region), Credentials: creds})
	if err != nil {
		return nil, err
	}
	s3 := s3.New(session)
	return s3, nil
}

func CreateNewObject() uuid.UUID {
	filename := uuid.New()
	return filename
}

func CreateS3(s3Object *s3.S3) error {
	params := &s3.CreateBucketInput{
		Bucket: aws.String(utils.Bucketname),
	}
	_, err := s3Object.CreateBucket(params)
	if err != nil {
		return fmt.Errorf("failed to create bucket '%s': %w", utils.Bucketname, err)
	}
	return nil
}

func PutS3(s3Object *s3.S3, data []byte, filename string) error {
	input := &s3.PutObjectInput{Body: strings.NewReader(string(data)), Key: aws.String(filename), Bucket: aws.String(utils.Bucketname)}
	_, err := s3Object.PutObject(input)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	return nil

}

func BucketExists(s3Object *s3.S3) bool {
	input := &s3.HeadBucketInput{Bucket: aws.String(utils.Bucketname)}
	_, err := s3Object.HeadBucket(input)
	if err != nil {
		if _, ok := err.(awserr.Error); ok {
			return false
		}
	}
	return true

}
