package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type FileSystemStorage struct {
	root string
}

func NewFileSystemStorage(root string) *FileSystemStorage {
	return &FileSystemStorage{root: root}
}

func (fs *FileSystemStorage) bucketPath(bucket string) string {
	return filepath.Join(fs.root, bucket)
}

func (fs *FileSystemStorage) objectPath(bucket, object string) string {
	return filepath.Join(fs.bucketPath(bucket), object)
}

func (fs *FileSystemStorage) Upload(ctx context.Context, bucket, object string, r io.Reader, size int64, contentType string) error {
	bPath := fs.bucketPath(bucket)
	if err := os.MkdirAll(bPath, 0755); err != nil {
		return err
	}

	dest := fs.objectPath(bucket, object)
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	return err
}

func (fs *FileSystemStorage) Download(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
	path := fs.objectPath(bucket, object)
	return os.Open(path)
}

func (fs *FileSystemStorage) List(ctx context.Context, bucket string) ([]string, error) {
	bPath := fs.bucketPath(bucket)
	files, err := os.ReadDir(bPath)
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, f := range files {
		if !f.IsDir() {
			objects = append(objects, f.Name())
		}
	}
	return objects, nil
}

func (fs *FileSystemStorage) Exists(ctx context.Context, bucket, object string) (bool, error) {
	path := fs.objectPath(bucket, object)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
