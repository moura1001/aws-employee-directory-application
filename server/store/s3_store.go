package store

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/moura1001/aws-employee-directory-application/server/utils"
)

type S3Store struct {
	bucket string
}

func NewS3Store() S3Store {
	return S3Store{
		bucket: utils.PHOTOS_BUCKET,
	}
}

func (s S3Store) getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(opts *config.LoadOptions) error {
		opts.Region = utils.AWS_DEFAULT_REGION
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error to get s3 connection. Details: '%s'", err)
	}

	return s3.NewFromConfig(cfg), nil
}

func (s S3Store) GeneratePresignedURL(objectKey string) (string, error) {
	errMsg := "error to get s3 object presigned url%s. Details: '%s'"

	svc, err := s.getS3Client()
	if err != nil {
		return "", fmt.Errorf(errMsg, "", err)
	}

	presignClient := s3.NewPresignClient(svc)

	result, err := presignClient.PresignGetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(time.Minute*2),
	)
	if err != nil {
		return "", fmt.Errorf(errMsg, " PresignGetObject", err)
	}

	return result.URL, nil
}

func (s S3Store) UploadObject(objectKey string, content []byte) error {
	errMsg := "error to upload s3 object%s. Details: '%s'"

	svc, err := s.getS3Client()
	if err != nil {
		return fmt.Errorf(errMsg, "", err)
	}

	contentBuffer := bytes.NewReader(content)

	uploader := manager.NewUploader(svc)

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   contentBuffer,
	})
	if err != nil {
		return fmt.Errorf(errMsg, " Upload", err)
	}

	return nil
}
