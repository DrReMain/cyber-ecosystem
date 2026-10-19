package population

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/population"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type populationRP struct {
	shared.RP
}

func NewPopulationRP(logger *slog.Logger, p *platform.Platform) PopulationRP {
	return &populationRP{
		RP: shared.NewRP(logger.With("module", "module/population_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *populationRP) List(ctx context.Context, in *PopulationListIn) (*PopulationListOut, error) {
	query := rp.Platform.GetClient(ctx).Population.Query()
	helper.Where(query, in.Code != nil, func() predicate.Population { return population.CodeContainsFold(*in.Code) })
	helper.Where(query, in.Name != nil, func() predicate.Population { return population.NameContainsFold(*in.Name) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"code":      func(sel helper.SQLSelector) { query.Order(sel(population.FieldCode)) },
		"name":      func(sel helper.SQLSelector) { query.Order(sel(population.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(population.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(population.FieldUpdatedAt)) },
	})
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &PopulationListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(rs, mapPopulation),
	}, nil
}

func (rp *populationRP) FindByCode(ctx context.Context, code string) (*Population, error) {
	r, err := rp.Platform.GetClient(ctx).Population.Query().Where(population.Code(code)).Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapPopulation(r), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapPopulation(d *ent.Population) *Population {
	return &Population{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Code:      d.Code,
		Name:      d.Name,
		Plane:     d.Plane,
	}
}
