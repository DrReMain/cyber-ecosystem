package storage

import (
	"context"
	"errors"
	"io"
)

type FileInfo struct {
	ID          string
	Name        string
	ContentType string
	Size        int64
}

type Storage interface {
	Upload(ctx context.Context, id string, name string, reader io.Reader, size int64, contentType string) (*FileInfo, error)
	Download(ctx context.Context, id string) (io.ReadCloser, *FileInfo, error)
	Delete(ctx context.Context, id string) error
}

var (
	ErrNotFound    = errors.New("storage: object not found")
	ErrForbidden   = errors.New("storage: access denied")
	ErrLimitExceed = errors.New("storage: file size exceeds limit")
)
