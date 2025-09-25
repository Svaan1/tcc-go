package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, bucket, object string, r io.Reader, size int64, contentType string) error
	Download(ctx context.Context, bucket, object string) (io.ReadCloser, error)
	List(ctx context.Context, bucket string) ([]string, error)
	Exists(ctx context.Context, bucket, object string) (bool, error)
}
