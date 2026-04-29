package data

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"roci.dev/fracdex"

	"github.com/go-kratos/kratos/v2/log"

	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/template/services/base/internal/biz"
	"cyber-ecosystem/apps/template/services/base/internal/ent"
	"cyber-ecosystem/apps/template/services/base/internal/ent/message"
	"cyber-ecosystem/apps/template/services/base/internal/ent/predicate"
	"cyber-ecosystem/apps/template/services/base/internal/platform"
)

type messageRP struct {
	RP
}

func NewMessageRP(logger log.Logger, p *platform.Platform) biz.MessageRP {
	return &messageRP{
		RP: RP{
			log:      log.NewHelper(log.With(logger, "module", "data/message_rp")),
			platform: p,
		},
	}
}

// region[rgba(0,188,212,0.12)] 🩵 Repo --------------------------------------------------------------------------------

func (rp *messageRP) Create(ctx context.Context, m *biz.Message) (*biz.Message, error) {
	created, err := rp.platform.GetClient(ctx).Message.Create().
		SetTitle(*m.Title).
		SetNillableContent(m.Content).
		Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMessage(created), nil
}

func (rp *messageRP) Update(ctx context.Context, fieldsMask []string, m *biz.Message) (*biz.Message, error) {
	updater := rp.platform.GetClient(ctx).Message.UpdateOneID(*m.ID)
	utils.Handler{
		"title": {
			Condition: m.Title != nil,
			OnTrue:    func() { updater.SetTitle(*m.Title) },
			OnFalse:   func() {},
		},
		"content": {
			Condition: m.Content != nil,
			OnTrue:    func() { updater.SetContent(*m.Content) },
			OnFalse:   func() { updater.SetContent("") },
		},
		"status": {
			Condition: m.Status != nil,
			OnTrue:    func() { updater.SetStatus(*m.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)

	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMessage(updated), nil
}

func (rp *messageRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.platform.GetClient(ctx).Message.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *messageRP) Get(ctx context.Context, id string) (*biz.Message, error) {
	d, err := rp.platform.GetClient(ctx).Message.Get(ctx, id)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMessage(d), nil
}

func (rp *messageRP) Query(ctx context.Context, in *biz.MessageQueryIn) (*biz.MessageQueryOut, error) {
	query := rp.platform.GetClient(ctx).Message.Query()
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtA), message.CreatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtZ), message.CreatedAtLTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtA), message.UpdatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtZ), message.UpdatedAtLTE)
	entutil.Where(query, in.ID != nil, func() predicate.Message { return message.IDEQ(*in.ID) })
	entutil.Where(query, in.Title != nil, func() predicate.Message { return message.TitleContainsFold(*in.Title) })
	entutil.Where(query, in.Status != nil, func() predicate.Message { return message.StatusEQ(*in.Status) })
	// Apply user-specified ordering rules; supported fields are createdAt, updatedAt and sort.
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"createdAt": func(sel entutil.SQLSelector) { query.Order(sel(message.FieldCreatedAt)) },
		"updatedAt": func(sel entutil.SQLSelector) { query.Order(sel(message.FieldUpdatedAt)) },
		"sort":      func(sel entutil.SQLSelector) { query.Order(sel(message.FieldSort)) },
	})
	// Append a fixed default sort as a tie-breaker so rows without an explicit order remain stable.
	query.Order(func(s *sql.Selector) { s.OrderBy(s.C(message.FieldSort)) })

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	pos, err := query.All(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return &biz.MessageQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapMessage),
	}, nil
}

func (rp *messageRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.Message, error) {
	var prevSort, nextSort string
	client := rp.platform.GetClient(ctx)

	if prevID != nil {
		d, err := client.Message.Get(ctx, *prevID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		prevSort = d.Sort
	}

	if nextID != nil {
		d, err := client.Message.Get(ctx, *nextID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		nextSort = d.Sort
	}

	newSort, err := fracdex.KeyBetween(prevSort, nextSort)
	if err != nil {
		return nil, err
	}

	updated, err := client.Message.UpdateOneID(id).SetSort(newSort).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMessage(updated), nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func mapMessage(d *ent.Message) *biz.Message {
	return &biz.Message{
		ID:        &d.ID,
		CreatedAt: &d.CreatedAt,
		UpdatedAt: &d.UpdatedAt,
		Title:     &d.Title,
		Content:   &d.Content,
		Status:    &d.Status,
	}
}
