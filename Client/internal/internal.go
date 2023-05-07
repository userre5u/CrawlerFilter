package internal

import (
	"crawlerDetection/Client/s3Service"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

type object_metadata struct {
	name         string
	lastmodified time.Time
	tag          string
	size         int64
}

type collectionData []object_metadata

func filterOutput(output *s3.ListObjectsOutput) collectionData {
	objects := output.Contents
	collectMetaData := make(collectionData, 0, len(objects))
	if len(objects) == 0 {

		return collectMetaData
	}
	if len(objects) == 1 {

		collectMetaData = append(collectMetaData, object_metadata{name: *objects[0].Key, size: *objects[0].Size, tag: *objects[0].ETag, lastmodified: *objects[0].LastModified})
		return collectMetaData
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.Before(*objects[j].LastModified)
	})
	for _, object := range objects[:len(objects)-1] {
		collectMetaData = append(collectMetaData, object_metadata{name: *object.Key, size: *object.Size, tag: *object.ETag, lastmodified: *object.LastModified})
	}

	return collectMetaData

}

func runList(s3 *s3.S3) (collectionData, error) {
	output, err := s3Service.ListObjects(s3)
	if err != nil {
		return nil, err
	}
	metaData_List := filterOutput(output)
	return metaData_List, nil
}

func downloadObjects(downloader *s3manager.Downloader, objectsMetadata collectionData) {
	for _, object := range objectsMetadata {
		fd, err := os.Create("objects_Tests/" + object.name)
		if err != nil {
			fmt.Println(fmt.Errorf("[-] Coult not create file: %q, %w", object.name, err))
			continue
		}
		defer fd.Close()
		n_bytes, err := downloader.Download(fd, &s3.GetObjectInput{
			Bucket: aws.String("bucketbuckettt"),
			Key:    aws.String(object.name),
		})
		if err != nil {
			fmt.Println(fmt.Errorf("[-] Could not download file: %q, %w", object.name, err))
			continue
		}
		fmt.Printf("File name: %q downloaded, %d bytes\n", object.name, n_bytes)
	}

}

func getS3(sess *session.Session) *s3.S3 {
	return s3.New(sess)
}

func getDownloader(sess *session.Session) *s3manager.Downloader {
	return s3manager.NewDownloader(sess)
}

func Start(logger *logrus.Logger, sess *session.Session, sessionKey string) {
	s3Object := getS3(sess)
	collectionData, err := runList(s3Object)
	if err != nil {
		logger.Error(err)
		return
	}
	downloadObjects(getDownloader(sess), collectionData)
	// for range time.Tick(time.Second * 1) {

	// }
}
