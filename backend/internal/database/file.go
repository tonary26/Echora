package database

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStore struct {
	internal *minio.Client
	public   *minio.Client
	bucket   string
}

func NewFileStore(ctx context.Context, endPoint string, publicEndPoint string, accessKey string, secretKey string, bucket string) (*FileStore, error) {
	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	internal, err := minio.New(endPoint, &minio.Options{Creds: creds})
	if err != nil {
		return nil, err
	}

	public, err := minio.New(publicEndPoint, &minio.Options{Creds: creds, Region: "us-east-1"})
	if err != nil {
		return nil, err
	}

	for i := 0; i < 15; i++ {
		exists, err := internal.BucketExists(ctx, bucket)
		if err == nil {
			if !exists {
				if err := internal.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
					return nil, err
				}
			}
			return &FileStore{internal: internal, public: public, bucket: bucket}, nil
		}
		time.Sleep(time.Second)
	}

	return nil, errors.New("Minio недоступен")
}

func (s *FileStore) PutFile(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.internal.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *FileStore) DeleteFile(ctx context.Context, key string) error {
	return s.internal.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *FileStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.public.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}