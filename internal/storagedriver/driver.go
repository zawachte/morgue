package storagedriver

import (
	"bytes"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

type StorageDriver interface {
	GetLocalStorageLocation() string
	UploadTar(string) error
}

type localStorageDriver struct {
	localStorageLocation string
	logger               zap.Logger
}

func (l *localStorageDriver) GetLocalStorageLocation() string {
	return l.localStorageLocation
}

func (l *localStorageDriver) UploadTar(directoryName string) error {
	// noop for local
	return nil
}

type s3StorageDriver struct {
	localStorageLocation string
	region               string
	bucket               string
	logger               zap.Logger
}

func (l *s3StorageDriver) GetLocalStorageLocation() string {
	return l.localStorageLocation
}

func (l *s3StorageDriver) UploadTar(directoryName string) error {

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := session.NewSession(&aws.Config{Region: aws.String(l.region)})
	if err != nil {
		return err
	}

	latestBackup := path.Join(l.localStorageLocation, directoryName)

	// Open the file for use
	file, err := os.Open(latestBackup)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = s3.New(s).ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(l.bucket),
		Key:                aws.String(directoryName),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})

	return err
}

type S3StorageDriverParams struct {
	Region string
	Bucket string
}

type StorageDriverParams struct {
	LocalStorageLocation  string
	S3StorageDriverParams *S3StorageDriverParams
	Logger                zap.Logger
}

func NewStorageDriver(params StorageDriverParams) (StorageDriver, error) {
	if params.S3StorageDriverParams != nil {
		return &s3StorageDriver{
			region:               params.S3StorageDriverParams.Region,
			bucket:               params.S3StorageDriverParams.Bucket,
			localStorageLocation: params.LocalStorageLocation,
		}, nil
	}

	return &localStorageDriver{
		localStorageLocation: params.LocalStorageLocation,
		logger:               params.Logger,
	}, nil
}
