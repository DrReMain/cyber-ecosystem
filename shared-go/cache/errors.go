package cache

import "errors"

var (
	ErrCacheMiss       = errors.New("cache miss")
	ErrKeyNotFound     = errors.New("key not found")
	ErrSessionNotFound = errors.New("session not found")
	ErrQuotaExceeded   = errors.New("quota exceeded")
	ErrInvalidArgument = errors.New("invalid argument")
)
