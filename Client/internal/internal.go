package internal

import (
	"crawlerDetection/Client/s3Service"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
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
	objects = objects[:len(objects)-1]
	fmt.Println(objects)
	for _, object := range objects {
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

func Start(logger *logrus.Logger, s3 *s3.S3, sessionKey string) {
	// _ = returns result after filter
	_, err := runList(s3)
	if err != nil {
		logger.Error(err)
		return
	}
	// for range time.Tick(time.Second * 1) {

	// }
}
