package internal

import (
	"bufio"
	"context"
	"crawlerDetection/Client/s3Service"
	"crawlerDetection/Client/utils"
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
	DateTime      string
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

func (p Global_objects) prepareMetadataInsert(object object_metadata) {
	timeNow := time.Now()
	fd, err := os.OpenFile(utils.DownloadObjectsFolder+object.name, os.O_RDONLY, 0444)
	if err != nil {
		p.Logger.Errorf("Could not open filename %q for reading: %s", object.name, err)
	}
	defer fd.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, fd); err != nil {
		p.Logger.Errorf("Could not calculate hash of file %q: %s", object.name, err)
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	err = p.InsertMetadataToDb(object.name, sum, timeNow.String(), object.lastmodified.String(), object.tag, object.size)
	if err != nil {
		p.Logger.Errorf("Could not insert data to DB: %s", err)
	}
	value := object.internal_metadata
	*value.deleteFile = true
	p.Logger.Infof("[+] Successfully prepared object metadata: %q to be inserted to DB, setting delete attribute to 'true'", object.name)

}

func (p Global_objects) runList() (collectionData, error) {
	output, err := s3Service.ListObjects(p.Object_s3)
	if err != nil {
		return nil, err
	}
	metaData_List := filterOutput(output)
	return metaData_List, nil
}

func (p Global_objects) downloadObjects(object object_metadata) {
	fd, err := os.Create(utils.DownloadObjectsFolder + object.name)
	if err != nil {
		p.Logger.Errorf("[-] Coult not create file: %q, %w", object.name, err)
	}
	defer fd.Close()
	n_bytes, err := p.Object_downloader.Download(fd, &s3.GetObjectInput{
		Bucket: aws.String(utils.Bucketname),
		Key:    aws.String(object.name),
	})
	if err != nil {
		p.Logger.Errorf("[-] Could not download file: %q, %w", object.name, err)
	}
	p.Logger.Infof("[+] File name: %q downloaded - %d bytes", object.name, n_bytes)

}

func parseLine(content string) (FileContent, error) {
	var fc FileContent
	if err := json.Unmarshal([]byte(content), &fc); err != nil {
		return fc, err
	}
	return fc, nil
}

func (p Global_objects) prepareContentInsert(object object_metadata) {
	fd, err := os.OpenFile(utils.DownloadObjectsFolder+object.name, os.O_RDONLY, 0444)
	if err != nil {
		p.Logger.Error(err)
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fc, err := parseLine(scanner.Text())
		if err != nil {
			p.Logger.Errorf("Error during data parsing: %s", err)
		}
		p.InsertContentToDb(fc)
	}
	p.Logger.Infof("[+] Successfully inserted object's content: %q to DB", object.name)

}

func (p Global_objects) deleteRemoteFile(object object_metadata) {
	if *object.internal_metadata.deleteFile {
		err := s3Service.DeleteObject(p.Object_s3, object.name)
		if err != nil {
			p.Logger.Errorf("[-] Unable to delete object: %q - reason: %s", object.name, err)
		}
		p.Logger.Infof("[+] Successfully deleted remote object: %q", object.name)
	}

}

func (p Global_objects) deleteLocalFile(object object_metadata) {
	if *object.internal_metadata.deleteFile {
		if err := os.Remove(utils.DownloadObjectsFolder + object.name); err != nil {
			p.Logger.Errorf("[-] unable to delete local object: %q - reason: %s", object.name, err)
		}
		p.Logger.Infof("[+] Successfully deleted local object: %q", object.name)
	}

}

func Start(ctx context.Context, sessionKey string) {
	globalObject := ctx.Value(Global_objects{}).(Global_objects)
	for range time.Tick(time.Second * 60) {
		globalObject.Logger.Info("[+] Starting new Extraction...")
		collectionData, err := globalObject.runList()
		time.Sleep(time.Second * 2)
		if err != nil {
			globalObject.Logger.Error(err)
			continue
		}
		if len(collectionData) > 0 {
			for _, object := range collectionData {
				globalObject.downloadObjects(object)
				globalObject.prepareMetadataInsert(object)
				globalObject.prepareContentInsert(object)
				globalObject.deleteRemoteFile(object)
				globalObject.deleteLocalFile(object)
				globalObject.Logger.Info("[+] Finished extraction, sleeping for 60 seconds...")
				continue
			}
		}
		globalObject.Logger.Info("[-] No files found on S3...")
	}

}
