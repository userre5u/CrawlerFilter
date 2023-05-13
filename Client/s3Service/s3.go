package s3Service

import (
	"crawlerDetection/Client/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func CreateSession(config *utils.Config) (*session.Session, error) {
	creds := credentials.NewStaticCredentials(config.Aws_access_key_id, config.Aws_secret_access_key, "")
	session, err := session.NewSession(&aws.Config{Region: aws.String(config.Region), Credentials: creds})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func ListObjects(s3object *s3.S3) (*s3.ListObjectsOutput, error) {
	input := s3.ListObjectsInput{Bucket: aws.String(utils.Bucketname)}
	output, err := s3object.ListObjects(&input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func DeleteObject(s3object *s3.S3, object string) error {
	input := s3.DeleteObjectInput{Bucket: aws.String(utils.Bucketname), Key: &object}
	_, err := s3object.DeleteObject(&input)
	if err != nil {
		return err
	}
	return nil
}

func GetS3(sess *session.Session) *s3.S3 {
	return s3.New(sess)
}

func GetDownloader(sess *session.Session) *s3manager.Downloader {
	return s3manager.NewDownloader(sess)
}
