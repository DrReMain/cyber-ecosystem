package cache

import "errors"

var ErrCacheMiss = errors.New("cache miss")

var ErrKeyNotFound = errors.New("key not found")

var ErrSessionNotFound = errors.New("session not found")

var ErrQuotaExceeded = errors.New("quota exceeded")

var ErrInvalidArgument = errors.New("invalid argument")
