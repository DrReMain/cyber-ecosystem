package biz

import (
	"time"
)

// region[rgba(239,83,80,0.15)] 🔴 Model -------------------------------------------------------------------------------

type File struct {
	ID          string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Name        string
	ContentType string
	Size        int64
	Status      string
}
