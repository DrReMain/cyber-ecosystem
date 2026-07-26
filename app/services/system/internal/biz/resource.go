package biz

import (
	"context"
	"log/slog"
)

// DO ------------------------------------------------------------------------------------------------------------------

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
	ListResource(ctx context.Context) ([]*ResourceService, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ResourceUC struct {
	UC
	resourceRP ResourceRP
}

func NewResourceUC(logger *slog.Logger, tm Transaction, resourceRP ResourceRP) *ResourceUC {
	return &ResourceUC{
		UC:         UC{log: logger.With("module", "biz/resource"), tm: tm},
		resourceRP: resourceRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ResourceUC) ListResource(ctx context.Context) ([]*ResourceService, error) {
	return uc.resourceRP.ListResource(ctx)
}
