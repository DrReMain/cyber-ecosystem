package redis

import "time"

// Config maps conf.Data.Redis to go-redis options. Backend-agnostic cache code
// never sees this type; only the redis adapter and the service platform layer do.
type Config struct {
	Network         string
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	ConnMaxLifetime time.Duration
	DialTimeout     time.Duration // 0 → go-redis default (5s)
	PoolTimeout     time.Duration // 0 → go-redis default (ReadTimeout+1s)
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}
