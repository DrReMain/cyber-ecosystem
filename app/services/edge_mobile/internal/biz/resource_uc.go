package biz

import (
	"context"
	"log/slog"
)

// Port ------------------------------------------------------------------------------------------------------------------

type ResourceRP interface {
	ListResource(ctx context.Context) ([]*ResourceService, error)
}

// UC --------------------------------------------------------------------------------------------------------------------

type ResourceUC struct {
	UC
	resourceRP ResourceRP
}

func NewResourceUC(logger *slog.Logger, tm Transaction, resourceRP ResourceRP) *ResourceUC {
	return &ResourceUC{
		UC: UC{
			log: logger.With("module", "biz/resource_uc"),
			tm:  tm,
		},
		resourceRP: resourceRP,
	}
}

// Method ----------------------------------------------------------------------------------------------------------------

func (uc *ResourceUC) ListResource(ctx context.Context) ([]*ResourceService, error) {
	return uc.resourceRP.ListResource(ctx)
}
