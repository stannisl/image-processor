package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type BucketName string

func (b BucketName) String() string {
	return string(b)
}

const (
	OriginalBucketName  BucketName = "originals"
	ProcessedBucketName BucketName = "processed"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketOrig BucketName
	bucketProc BucketName
}

func NewMinIO(endpoint, accessKey, secretKey string) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	s := &MinIOStorage{
		client:     client,
		bucketOrig: OriginalBucketName,
		bucketProc: ProcessedBucketName,
	}

	for _, bucket := range []string{s.bucketOrig.String(), s.bucketProc.String()} {
		exists, err := client.BucketExists(context.Background(), bucket)
		if err != nil {
			return nil, fmt.Errorf("bucket exists: %w", err)
		}

		if !exists {
			if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
				return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
	}

	return s, nil
}

func (s *MinIOStorage) SaveOriginal(ctx context.Context, id, contentType string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucketOrig.String(), id,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (s *MinIOStorage) SaveProcessed(ctx context.Context, id, contentType string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucketProc.String(), id,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (s *MinIOStorage) GetOriginal(ctx context.Context, id string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucketOrig.String(), id, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	return data, nil
}

func (s *MinIOStorage) GetProcessed(ctx context.Context, id string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucketProc.String(), id, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	return data, nil
}
