package data

import (
	"context"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	"roci.dev/fracdex"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/core/internal/biz"
	"cyber-ecosystem/app/services/core/internal/ent"
	"cyber-ecosystem/app/services/core/internal/ent/predicate"
	"cyber-ecosystem/app/services/core/internal/ent/user"
	"cyber-ecosystem/app/services/core/internal/platform"
)

type userRP struct {
	RP
}

func NewUserRP(logger *slog.Logger, p *platform.Platform) biz.UserRP {
	return &userRP{
		RP: RP{
			log:      logger.With("module", "data/user_rp"),
			platform: p,
		},
	}
}

// Repo --------------------------------------------------------------------------------------------------------

func (rp *userRP) Create(ctx context.Context, u *biz.User) (*biz.User, error) {
	created, err := rp.platform.GetClient(ctx).User.Create().
		SetNillableNickname(u.Nickname).
		SetNillableAvatar(u.Avatar).
		SetPhone(*u.Phone).
		SetNillableStatus(u.Status).
		SetPasswordHash(*u.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(created), nil
}

func (rp *userRP) Update(ctx context.Context, fieldsMask []string, u *biz.User) (*biz.User, error) {
	updater := rp.platform.GetClient(ctx).User.UpdateOneID(u.ID)
	helper.Handler{
		"nickname": {Condition: u.Nickname != nil, OnTrue: func() { updater.SetNillableNickname(u.Nickname) }, OnFalse: func() { updater.SetNickname("") }},
		"avatar":   {Condition: u.Avatar != nil, OnTrue: func() { updater.SetNillableAvatar(u.Avatar) }, OnFalse: func() { updater.SetAvatar("") }},
		"phone":    {Condition: u.Phone != nil, OnTrue: func() { updater.SetPhone(*u.Phone) }, OnFalse: func() {}},
		"password": {Condition: u.PasswordHash != nil, OnTrue: func() { updater.SetPasswordHash(*u.PasswordHash) }, OnFalse: func() {}},
	}.Emit(fieldsMask)

	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(updated), nil
}

func (rp *userRP) UpdateStatus(ctx context.Context, id, status string) (*biz.User, error) {
	updated, err := rp.platform.GetClient(ctx).User.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(updated), nil
}

func (rp *userRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.platform.GetClient(ctx).User.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *userRP) List(ctx context.Context, in *biz.UserListIn) (*biz.UserListOut, error) {
	query := rp.platform.GetClient(ctx).User.Query()
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), user.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), user.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtA), user.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtZ), user.UpdatedAtLTE)
	helper.Where(query, in.Phone != nil, func() predicate.User { return user.PhoneEQ(*in.Phone) })
	helper.Where(query, in.Status != nil, func() predicate.User { return user.StatusEQ(*in.Status) })
	helper.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, helper.FOMapping{
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldUpdatedAt)) },
		"sort":      func(sel helper.SQLSelector) { query.Order(sel(user.FieldSort)) },
	})
	query.Order(func(s *sql.Selector) { s.OrderBy(s.C(user.FieldSort)) })

	total, offset, limit, err := helper.ApplyPagination(ctx, query, in.PageRequest,
		helper.NewPageConfig(helper.DefaultPageSize, helper.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return &biz.UserListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(items, mapUser),
	}, nil
}

func (rp *userRP) Get(ctx context.Context, id string) (*biz.User, error) {
	d, err := rp.platform.GetClient(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

func (rp *userRP) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	count, err := rp.platform.GetClient(ctx).User.Query().
		Where(user.PhoneEQ(phone)).
		Count(ctx)
	if err != nil {
		return false, rp.platform.HandleEntError(err)
	}
	return count > 0, nil
}

func (rp *userRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.User, error) {
	var prevSort, nextSort string
	client := rp.platform.GetClient(ctx)

	if prevID != nil {
		d, err := client.User.Get(ctx, *prevID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		prevSort = d.Sort
	}
	if nextID != nil {
		d, err := client.User.Get(ctx, *nextID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		nextSort = d.Sort
	}

	newSort, err := fracdex.KeyBetween(prevSort, nextSort)
	if err != nil {
		return nil, err
	}

	updated, err := client.User.UpdateOneID(id).SetSort(newSort).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(updated), nil
}

// Private --------------------------------------------------------------------------------------------------------

func mapUser(d *ent.User) *biz.User {
	return &biz.User{
		ID:           d.ID,
		Nickname:     &d.Nickname,
		Avatar:       &d.Avatar,
		Phone:        &d.Phone,
		Status:       &d.Status,
		PasswordHash: &d.PasswordHash,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
