package internal

import (
	"context"
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

type Global_objects struct {
	Object_s3         *s3.S3
	Object_sess       *session.Session
	Object_downloader *s3manager.Downloader
	Logger            *logrus.Logger
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

func (p Global_objects) runList() (collectionData, error) {
	output, err := s3Service.ListObjects(p.Object_s3)
	if err != nil {
		return nil, err
	}
	metaData_List := filterOutput(output)
	return metaData_List, nil
}

func (p Global_objects) downloadObjects(objectsMetadata collectionData) {
	for _, object := range objectsMetadata {
		fd, err := os.Create("objects_Tests/" + object.name)
		if err != nil {
			fmt.Println(fmt.Errorf("[-] Coult not create file: %q, %w", object.name, err))
			continue
		}
		defer fd.Close()
		n_bytes, err := p.Object_downloader.Download(fd, &s3.GetObjectInput{
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

func GetS3(sess *session.Session) *s3.S3 {
	return s3.New(sess)
}

func GetDownloader(sess *session.Session) *s3manager.Downloader {
	return s3manager.NewDownloader(sess)
}

func Start(ctx context.Context, sessionKey string) {
	primeType := ctx.Value(Global_objects{}).(Global_objects)
	collectionData, err := primeType.runList()
	if err != nil {
		primeType.Logger.Error(err)
	}
	primeType.downloadObjects(collectionData)
	// for range time.Tick(time.Second * 1) {

	// }
}
