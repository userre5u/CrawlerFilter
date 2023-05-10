package internal

import (
	"bufio"
	"context"
	"crawlerDetection/Client/s3Service"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

type FileContent struct {
	Date          string
	Session       string
	IP            string
	Crawler       bool
	IpType        string
	UA            string
	Country       string
	SessionKey    string
	Path          string
	Method        string
	CriticalWords map[string]bool
}

type internal_metadata struct {
	deleteFile *bool
}
type object_metadata struct {
	name         string
	lastmodified time.Time
	tag          string
	size         int64
	internal_metadata
}

type Global_objects struct {
	Object_s3         *s3.S3
	Object_sess       *session.Session
	Object_downloader *s3manager.Downloader
	Logger            *logrus.Logger
	DBobject          *sql.DB
}

type collectionData []object_metadata

func filterOutput(output *s3.ListObjectsOutput) collectionData {
	objects := output.Contents
	retBool := false
	collectMetaData := make(collectionData, 0, len(objects))
	if len(objects) == 0 {

		return collectMetaData
	}
	if len(objects) == 1 {
		collectMetaData = append(collectMetaData, object_metadata{name: *objects[0].Key, size: *objects[0].Size, tag: *objects[0].ETag, lastmodified: *objects[0].LastModified, internal_metadata: internal_metadata{deleteFile: &retBool}})
		return collectMetaData
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.Before(*objects[j].LastModified)
	})
	for _, object := range objects[:len(objects)-1] {
		collectMetaData = append(collectMetaData, object_metadata{name: *object.Key, size: *object.Size, tag: *object.ETag, lastmodified: *object.LastModified, internal_metadata: internal_metadata{deleteFile: &retBool}})
	}

	return collectMetaData

}

func (p Global_objects) prepareMetadataInsert(data collectionData) {
	timeNow := time.Now()
	for _, object := range data {
		fd, err := os.OpenFile("Client/objects_Tests/"+object.name, os.O_RDONLY, 0444)
		if err != nil {
			p.Logger.Errorf("Could not open filename %q for reading: %s\n", object.name, err)
		}
		defer fd.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, fd); err != nil {
			p.Logger.Errorf("Could not calculate hash of file %q: %s\n", object.name, err)

		}
		sum := hex.EncodeToString(hash.Sum(nil))

		err = p.InsertMetadataToDb(object.name, sum, timeNow.String(), object.lastmodified.String(), object.tag, object.size)
		if err != nil {
			p.Logger.Errorf("Could not insert data to DB: %s\n", err)
			continue
		}
		value := object.internal_metadata
		*value.deleteFile = true

	}

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
		fd, err := os.Create("Client/objects_Tests/" + object.name)
		if err != nil {
			p.Logger.Errorf("[-] Coult not create file: %q, %w", object.name, err)
			continue
		}
		defer fd.Close()
		n_bytes, err := p.Object_downloader.Download(fd, &s3.GetObjectInput{
			Bucket: aws.String("bucketbuckettt"),
			Key:    aws.String(object.name),
		})
		if err != nil {
			p.Logger.Errorf("[-] Could not download file: %q, %w", object.name, err)
			continue
		}
		p.Logger.Infof("File name: %q downloaded, %d bytes\n", object.name, n_bytes)
	}
}

func parseLine(content string) (FileContent, error) {
	var fc FileContent
	if err := json.Unmarshal([]byte(content), &fc); err != nil {
		return fc, err
	}
	return fc, nil

}

func (p Global_objects) ParseFileContentInsert(objectMetadata collectionData) {
	for _, object := range objectMetadata {
		fd, err := os.OpenFile("Client/objects_Tests/"+object.name, os.O_RDONLY, 0444)
		if err != nil {
			p.Logger.Error(err)
			continue
		}
		defer fd.Close()
		scanner := bufio.NewScanner(fd)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			fc, err := parseLine(scanner.Text())
			p.Logger.Debugf("data from parser: %#v\n", fc)
			if err != nil {
				p.Logger.Errorf("Error during data parsing: %s\n", err)
				continue
			}
			p.InsertContentToDb(fc)
			break // remove this afterwards

		}

	}

}

func deleteRemoteFile(metadata collectionData) {
	// delete remote files from aws s3

}

func (p Global_objects) deleteLocalFile(metadata collectionData) {
	// delete local files
	for _, object := range metadata {
		if *object.internal_metadata.deleteFile {
			if err := os.Remove("Client/objects_Tests/" + object.name); err != nil {
				p.Logger.Errorf("Could not delete file name: %q, %s", object.name, err)
				continue
			}
			p.Logger.Infof("Deleted file: %q\n", object.name)
		}
	}
}

func GetS3(sess *session.Session) *s3.S3 {
	return s3.New(sess)
}

func GetDownloader(sess *session.Session) *s3manager.Downloader {
	return s3manager.NewDownloader(sess)
}

func Start(ctx context.Context, sessionKey string) {
	globalObject := ctx.Value(Global_objects{}).(Global_objects)
	collectionData, err := globalObject.runList()
	if err != nil {
		globalObject.Logger.Error(err)
	}
	if len(collectionData) > 0 {
		globalObject.downloadObjects(collectionData)
		globalObject.prepareMetadataInsert(collectionData)
		globalObject.ParseFileContentInsert(collectionData)
		//globalObject.deleteRemoteFile(collectionData)
		globalObject.deleteLocalFile(collectionData)
	}
	// for range time.Tick(time.Second * 1) {
	// }
}
