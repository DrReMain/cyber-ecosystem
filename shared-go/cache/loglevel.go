package cache

import "github.com/go-kratos/kratos/v2/log"

func ParseLogLevel(s string) log.Level {
	switch s {
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		return log.LevelInfo
	}
}
