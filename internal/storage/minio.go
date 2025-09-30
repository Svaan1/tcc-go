package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, useSSL bool) *MinioStorage {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(err)
	}
	return &MinioStorage{client: client}
}

func (m *MinioStorage) Upload(ctx context.Context, bucket, object string, r io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucket, object, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinioStorage) Download(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *MinioStorage) List(ctx context.Context, bucket string) ([]string, error) {
	var objects []string
	for objInfo := range m.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if objInfo.Err != nil {
			return nil, objInfo.Err
		}
		objects = append(objects, objInfo.Key)
	}
	return objects, nil
}

func (m *MinioStorage) Exists(ctx context.Context, bucket, object string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucket, object, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
