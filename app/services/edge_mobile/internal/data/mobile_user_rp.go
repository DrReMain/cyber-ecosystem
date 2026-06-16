package data

import (
	"context"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	"roci.dev/fracdex"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/edge_mobile/internal/biz"
	"cyber-ecosystem/app/services/edge_mobile/internal/ent"
	"cyber-ecosystem/app/services/edge_mobile/internal/ent/mobileuser"
	"cyber-ecosystem/app/services/edge_mobile/internal/ent/predicate"
	"cyber-ecosystem/app/services/edge_mobile/internal/platform"
)

type mobileUserRP struct {
	RP
}

func NewMobileUserRP(logger *slog.Logger, p *platform.Platform) biz.MobileUserRP {
	return &mobileUserRP{
		RP: RP{
			log:      logger.With("module", "data/mobile_user_rp"),
			platform: p,
		},
	}
}

// Repo --------------------------------------------------------------------------------------------------------

func (rp *mobileUserRP) Create(ctx context.Context, u *biz.MobileUser) (*biz.MobileUser, error) {
	created, err := rp.platform.GetClient(ctx).MobileUser.Create().
		SetNillableNickname(u.Nickname).
		SetNillableAvatar(u.Avatar).
		SetPhone(*u.Phone).
		SetNillableStatus(u.Status).
		SetPasswordHash(*u.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMobileUser(created), nil
}

func (rp *mobileUserRP) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	count, err := rp.platform.GetClient(ctx).MobileUser.Query().
		Where(mobileuser.PhoneEQ(phone)).
		Count(ctx)
	if err != nil {
		return false, rp.platform.HandleEntError(err)
	}
	return count > 0, nil
}

func (rp *mobileUserRP) Get(ctx context.Context, id string) (*biz.MobileUser, error) {
	d, err := rp.platform.GetClient(ctx).MobileUser.Get(ctx, id)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMobileUser(d), nil
}

func (rp *mobileUserRP) List(ctx context.Context, in *biz.MobileUserListIn) (*biz.MobileUserListOut, error) {
	query := rp.platform.GetClient(ctx).MobileUser.Query()
	helper.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtA), mobileuser.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtZ), mobileuser.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtA), mobileuser.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtZ), mobileuser.UpdatedAtLTE)
	helper.Where(query, in.Phone != nil, func() predicate.MobileUser { return mobileuser.PhoneEQ(*in.Phone) })
	helper.Where(query, in.Status != nil, func() predicate.MobileUser { return mobileuser.StatusEQ(*in.Status) })
	helper.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, helper.FOMapping{
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(mobileuser.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(mobileuser.FieldUpdatedAt)) },
		"sort":      func(sel helper.SQLSelector) { query.Order(sel(mobileuser.FieldSort)) },
	})
	query.Order(func(s *sql.Selector) { s.OrderBy(s.C(mobileuser.FieldSort)) })

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
	return &biz.MobileUserListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(items, mapMobileUser),
	}, nil
}

func (rp *mobileUserRP) Update(ctx context.Context, fieldsMask []string, u *biz.MobileUser) (*biz.MobileUser, error) {
	updater := rp.platform.GetClient(ctx).MobileUser.UpdateOneID(u.ID)
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
	return mapMobileUser(updated), nil
}

func (rp *mobileUserRP) UpdateStatus(ctx context.Context, id, status string) (*biz.MobileUser, error) {
	updated, err := rp.platform.GetClient(ctx).MobileUser.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMobileUser(updated), nil
}

func (rp *mobileUserRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.platform.GetClient(ctx).MobileUser.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *mobileUserRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.MobileUser, error) {
	var prevSort, nextSort string
	client := rp.platform.GetClient(ctx)

	if prevID != nil {
		d, err := client.MobileUser.Get(ctx, *prevID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		prevSort = d.Sort
	}
	if nextID != nil {
		d, err := client.MobileUser.Get(ctx, *nextID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		nextSort = d.Sort
	}

	newSort, err := fracdex.KeyBetween(prevSort, nextSort)
	if err != nil {
		return nil, err
	}

	updated, err := client.MobileUser.UpdateOneID(id).SetSort(newSort).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapMobileUser(updated), nil
}

// Private --------------------------------------------------------------------------------------------------------

func mapMobileUser(d *ent.MobileUser) *biz.MobileUser {
	return &biz.MobileUser{
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
