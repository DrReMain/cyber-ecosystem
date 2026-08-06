package shared

import "log/slog"

type UC struct {
	Log *slog.Logger
}

func NewUC(log *slog.Logger) UC {
	return UC{Log: log}
}
