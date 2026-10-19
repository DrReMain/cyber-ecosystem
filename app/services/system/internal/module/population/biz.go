package population

import (
	"context"
	"log/slog"
	"time"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Population struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Code      string
	Name      string
	Plane     string
}

type PopulationListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Code    *string
	Name    *string
}

type PopulationListOut struct {
	*commonpb.PageResponse
	List []*Population
}

// Port ----------------------------------------------------------------------------------------------------------------

type PopulationRP interface {
	List(ctx context.Context, in *PopulationListIn) (*PopulationListOut, error)
	FindByCode(ctx context.Context, code string) (*Population, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type PopulationUC struct {
	shared.UC
	populationRP PopulationRP
}

func NewPopulationUC(logger *slog.Logger, tm shared.Transaction, populationRP PopulationRP) *PopulationUC {
	return &PopulationUC{
		UC:           shared.NewUC(logger.With("module", "module/population"), tm),
		populationRP: populationRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *PopulationUC) List(ctx context.Context, in *PopulationListIn) (*PopulationListOut, error) {
	return uc.populationRP.List(ctx, in)
}
