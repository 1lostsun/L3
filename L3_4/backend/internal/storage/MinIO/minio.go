package MinIO

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
)

type Storage struct {
	Client *minio.Client
	Bucket string
}

func New() (*Storage, error) {
	var endpoint, accessKey, secretKey, bucket string
	endpoint = os.Getenv("ENDPOINT")
	accessKey = os.Getenv("ACCESS_KEY")
	secretKey = os.Getenv("SECRET_KEY")
	bucket = os.Getenv("BUCKET")

	fmt.Println(endpoint, accessKey, secretKey, bucket)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("MinIO constructor Error: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("MinIO Bucket Exists Error: %v", err)
	}

	if !exists {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("MinIO Create Bucket Error: %v", err)
		}
	}

	return &Storage{Client: client, Bucket: bucket}, nil
}

func (s *Storage) Upload(ctx context.Context, fileData []byte, outputPath, contentType string) error {
	_, err := s.Client.PutObject(ctx,
		s.Bucket,
		outputPath,
		bytes.NewReader(fileData),
		int64(len(fileData)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("MinIO Upload Error: %v", err)
	}

	return nil
}

func (s *Storage) Download(ctx context.Context, outputPath string) ([]byte, error) {
	obj, err := s.Client.GetObject(ctx, s.Bucket, outputPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("MinIO Download Error: %v", err)
	}

	return io.ReadAll(obj)
}

func (s *Storage) Delete(ctx context.Context, outputPath string) error {
	err := s.Client.RemoveObject(ctx, s.Bucket, outputPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("MinIO Delete Error: %v", err)
	}

	return nil
}
