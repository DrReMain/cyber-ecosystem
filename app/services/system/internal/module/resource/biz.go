package resource

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type ResourceMethod struct {
	Name             string `json:"name"`
	FullName         string `json:"full_name"`
	RequestName      string `json:"request_name"`
	RequestFullName  string `json:"request_full_name"`
	ResponseName     string `json:"response_name"`
	ResponseFullName string `json:"response_full_name"`
	HttpMethod       string `json:"http_method"`
	HttpPath         string `json:"http_path"`
	Comment          string `json:"comment"`
	Builtin          bool   `json:"builtin"`   // builtin baseline op: not grantable
	Access           string `json:"access"`    // audience annotation as declared
	Datascope        bool   `json:"datascope"` // row-level narrowing applies: scope selection is offered
}

type ServiceMeta struct {
	Name       string            `json:"name"`
	FullName   string            `json:"full_name"`
	Package    string            `json:"package"`
	SourceFile string            `json:"source_file"`
	Comment    string            `json:"comment"`
	Methods    []*ResourceMethod `json:"methods"`
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
