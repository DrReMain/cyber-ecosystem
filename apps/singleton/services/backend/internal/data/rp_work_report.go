package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entworkreport "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/workreport"
)

type workReportRP struct {
	RP
}

func NewWorkReportRP(logger log.Logger, store *Store) biz.WorkReportRP {
	return &workReportRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_work_report")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *workReportRP) Create(ctx context.Context, r *biz.WorkReport) (*biz.WorkReport, error) {
	builder := rp.store.GetClient(ctx).WorkReport.Create()
	if r.Title != nil {
		builder.SetTitle(*r.Title)
	}
	if r.Content != nil {
		builder.SetContent(*r.Content)
	}
	if r.Type != nil {
		builder.SetType(*r.Type)
	}
	if r.DepartmentID != nil {
		builder.SetDepartmentID(*r.DepartmentID)
	}
	if r.AccessLevel != nil {
		builder.SetAccessLevel(*r.AccessLevel)
	}
	if r.Region != nil {
		builder.SetRegion(*r.Region)
	}

	result, err := builder.Save(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapWorkReport(result), nil
}

func (rp *workReportRP) Update(ctx context.Context, r *biz.WorkReport, fieldsMask []string) error {
	updater := rp.store.GetClient(ctx).WorkReport.UpdateOneID(*r.ID)
	utils.Handler{
		"title": {
			Condition: r.Title != nil,
			OnTrue:    func() { updater.SetTitle(*r.Title) },
			OnFalse:   func() {},
		},
		"content": {
			Condition: r.Content != nil,
			OnTrue:    func() { updater.SetContent(*r.Content) },
			OnFalse:   func() { updater.SetContent("") },
		},
		"type": {
			Condition: r.Type != nil,
			OnTrue:    func() { updater.SetType(*r.Type) },
			OnFalse:   func() {},
		},
		"access_level": {
			Condition: r.AccessLevel != nil,
			OnTrue:    func() { updater.SetAccessLevel(*r.AccessLevel) },
			OnFalse:   func() {},
		},
		"region": {
			Condition: r.Region != nil,
			OnTrue:    func() { updater.SetRegion(*r.Region) },
			OnFalse:   func() { updater.ClearRegion() },
		},
		"status": {
			Condition: r.Status != nil,
			OnTrue:    func() { updater.SetStatus(*r.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *workReportRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).WorkReport.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *workReportRP) Get(ctx context.Context, id string) (*biz.WorkReport, error) {
	result, err := rp.store.GetClient(ctx).WorkReport.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapWorkReport(result), nil
}

func (rp *workReportRP) Query(ctx context.Context, in *biz.WorkReportQueryIn) (*biz.WorkReportQueryOut, error) {
	query := rp.store.GetClient(ctx).WorkReport.Query()

	entutil.WherePtr(query, in.Type, entworkreport.TypeEQ)
	entutil.WherePtr(query, in.Status, entworkreport.StatusEQ)

	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entworkreport.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entworkreport.FieldUpdatedAt)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		singletonV1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}

	list, err := query.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return &biz.WorkReportQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(list, mapWorkReport),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapWorkReport(r *ent.WorkReport) *biz.WorkReport {
	return &biz.WorkReport{
		ID:           &r.ID,
		CreatedAt:    &r.CreatedAt,
		UpdatedAt:    &r.UpdatedAt,
		Title:        &r.Title,
		Content:      &r.Content,
		Type:         &r.Type,
		DepartmentID: r.DepartmentID,
		AccessLevel:  &r.AccessLevel,
		Region:       r.Region,
		CreatedBy:    r.CreatedBy,
		Status:       &r.Status,
	}
}
