package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	appConfig "s3-saver/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// StorageService описывает методы работы с хранилищем
type StorageService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service(cfg *appConfig.AppConfig) *S3Service {
	creds := credentials.NewStaticCredentialsProvider(cfg.S3Config.AccessKey, cfg.S3Config.SecretKey, "")

	sdkConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.S3Config.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to load SDK config: %v", err))
	}

	client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Config.Endpoint)
		o.UsePathStyle = true
	})

	return &S3Service{
		client: client,
		bucket: cfg.S3Config.BucketName,
	}
}

func (s *S3Service) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(header.Filename)

	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(newFileName),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}

	return newFileName, nil
}
