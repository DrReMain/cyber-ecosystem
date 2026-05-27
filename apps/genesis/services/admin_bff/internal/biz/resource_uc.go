package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port --------------------------------------------------------------------------------

type ResourceRP interface {
	ListResource(ctx context.Context) ([]*ResourceService, error)
}

// region[rgba(102,187,106,0.15)] 🟢 UC ----------------------------------------------------------------------------------

type ResourceUC struct {
	UC
	resourceRP ResourceRP
}

func NewResourceUC(logger log.Logger, tm Transaction, resourceRP ResourceRP) *ResourceUC {
	return &ResourceUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/resource_uc")),
			tm:  tm,
		},
		resourceRP: resourceRP,
	}
}

// region[rgba(186,104,200,0.15)] 🟣 Method ------------------------------------------------------------------------------

func (uc *ResourceUC) ListResource(ctx context.Context) ([]*ResourceService, error) {
	return uc.resourceRP.ListResource(ctx)
}
