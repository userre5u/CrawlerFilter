package s3Service

import (
	"crawlerDetection/Client/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	bucketname = "bucketbuckettt"
)

func CreateSession(config *utils.Config) (*s3.S3, error) {
	creds := credentials.NewStaticCredentials(config.Aws_access_key_id, config.Aws_secret_access_key, "")
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region), Credentials: creds})
	if err != nil {
		return nil, err
	}
	s3 := s3.New(session)
	return s3, nil
}

func ListObjects(s3object *s3.S3) (*s3.ListObjectsOutput, error) {
	input := s3.ListObjectsInput{Bucket: aws.String(bucketname)}
	output, err := s3object.ListObjects(&input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
