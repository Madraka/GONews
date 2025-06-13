package storage

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
}

func NewS3Storage() (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, err
	}

	// Create S3 client with modern endpoint resolver if custom endpoint is provided
	var s3Client *s3.Client
	if customEndpoint := os.Getenv("S3_ENDPOINT"); customEndpoint != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &customEndpoint
			o.UsePathStyle = true // Usually needed for custom endpoints like MinIO
		})
	} else {
		s3Client = s3.NewFromConfig(cfg)
	}

	bucketName := os.Getenv("S3_BUCKET")

	return &S3Storage{
		client:     s3Client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) Upload(file io.Reader, filename string) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (s *S3Storage) Download(filename string) (io.Reader, error) {
	resp, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (s *S3Storage) Delete(filename string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filename),
	})
	return err
}
