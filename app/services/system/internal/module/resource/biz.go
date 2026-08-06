package resource

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/system/internal/shared"
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

type ServiceMeta struct {
	Name       string
	FullName   string
	Package    string
	SourceFile string
	Comment    string
	Methods    []*ResourceMethod
}

// Port ----------------------------------------------------------------------------------------------------------------

type ResourceRP interface {
	ListResource(ctx context.Context) ([]*ServiceMeta, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ResourceUC struct {
	shared.UC
	resourceRP ResourceRP
}

func NewResourceUC(logger *slog.Logger, tm shared.Transaction, resourceRP ResourceRP) *ResourceUC {
	return &ResourceUC{
		UC:         shared.NewUC(logger.With("module", "module/resource"), tm),
		resourceRP: resourceRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ResourceUC) ListResource(ctx context.Context) ([]*ServiceMeta, error) {
	return uc.resourceRP.ListResource(ctx)
}
