package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Model ---------------------------------------------------------------------------------------------------------------

type ResourceMethod struct {
	Name             string
	FullName         string
	RequestName      string
	RequestFullName  string
	ResponseName     string
	ResponseFullName string
	HttpMethod       string
	HttpPath         string
	Comment          string
}

type ResourceService struct {
	Name       string
	FullName   string
	Package    string
	SourceFile string
	Comment    string
	Methods    []*ResourceMethod
}

// Port ----------------------------------------------------------------------------------------------------------------

type ResourceRP interface {
	ListServices(ctx context.Context) ([]*ResourceService, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ResourceUC struct {
	UC
	resourceRP ResourceRP
}

func NewResourceUC(logger log.Logger, tm Transaction, resourceRP ResourceRP) *ResourceUC {
	return &ResourceUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_resource")),
			tm:  tm,
		},
		resourceRP: resourceRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ResourceUC) ListResource(ctx context.Context) ([]*ResourceService, error) {
	return uc.resourceRP.ListServices(ctx)
}
