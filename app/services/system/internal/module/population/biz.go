package population

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Population struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Code      string
	Name      string
	Archetype string
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
	Create(ctx context.Context, p *Population) (*Population, error)
	List(ctx context.Context, in *PopulationListIn) (*PopulationListOut, error)
	FindByCode(ctx context.Context, code string) (*Population, error)
	NotifyChanged(ctx context.Context)
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

func (uc *PopulationUC) Create(ctx context.Context, code, name, archetype string) (*Population, error) {
	// The default population is seed-born; an operator-borne row (a patron
	// one, say) would silently turn every default seat birth into a
	// grantless identity — reserved rejects even where the seed row is
	// missing.
	if code == security.DefaultPopulation {
		return nil, systempb.ErrorSystemPopulationProtected("")
	}
	if !security.ValidArchetype(archetype) {
		return nil, systempb.ErrorSystemPopulationInvalidArchetype("")
	}
	_, err := uc.populationRP.FindByCode(ctx, code)
	if err == nil {
		return nil, systempb.ErrorSystemPopulationExists("")
	}
	if !errorspb.IsInfraErrorDbNotFound(err) {
		return nil, err
	}
	p, err := uc.populationRP.Create(ctx, &Population{Code: code, Name: name, Archetype: archetype})
	if err != nil {
		// The unique index is the race backstop — same reason as the
		// pre-check, never an infra leak.
		if errorspb.IsInfraErrorDbConstraint(err) {
			return nil, systempb.ErrorSystemPopulationExists("")
		}
		return nil, err
	}
	uc.populationRP.NotifyChanged(ctx)
	return p, nil
}

func (uc *PopulationUC) List(ctx context.Context, in *PopulationListIn) (*PopulationListOut, error) {
	return uc.populationRP.List(ctx, in)
}
