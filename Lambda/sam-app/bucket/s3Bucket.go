package bucket

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type S3Bucket struct{}

const (
	bucketname = "bucketbuckettt"
	MaxDataLen = 1024
)

func CreateSession() (*s3.S3, error) {
	creds := credentials.NewStaticCredentials("XXXXX", "YYYYY", "")
	session, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1"), Credentials: creds})
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
		Bucket: aws.String(bucketname),
	}
	_, err := s3Object.CreateBucket(params)
	if err != nil {
		return fmt.Errorf("failed to create bucket '%s': %w", bucketname, err)
	}
	return nil
}

func PutS3(s3Object *s3.S3, data []byte, filename string) error {
	input := &s3.PutObjectInput{Body: strings.NewReader(string(data)), Key: aws.String(filename), Bucket: aws.String(bucketname)}
	_, err := s3Object.PutObject(input)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	return nil

}

func BucketExists(s3Object *s3.S3) bool {
	input := &s3.HeadBucketInput{Bucket: aws.String(bucketname)}
	_, err := s3Object.HeadBucket(input)
	if err != nil {
		if _, ok := err.(awserr.Error); ok {
			return false
		}
	}
	return true

}
