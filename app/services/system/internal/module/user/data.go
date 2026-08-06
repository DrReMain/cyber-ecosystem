package user

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

type userRP struct {
	shared.RP
}

func NewUserRP(logger *slog.Logger, p *platform.Platform) UserRP {
	return &userRP{
		RP: shared.NewRP(logger.With("module", "module/user_rp"), p),
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *userRP) Create(ctx context.Context, u *User) (*User, error) {
	created, err := rp.Platform.GetClient(ctx).User.Create().
		SetTenantID(u.TenantID).
		SetNillableDeptID(u.DeptID).
		SetEmail(*u.Email).
		SetPasswordHash(*u.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(created), nil
}

func (rp *userRP) Update(ctx context.Context, fieldsMask []string, u *User) (*User, error) {
	updater := rp.Platform.GetClient(ctx).User.UpdateOneID(u.ID)
	helper.Handler{
		"email": {
			Condition: u.Email != nil,
			OnTrue:    func() { updater.SetEmail(*u.Email) },
			OnFalse:   func() {},
		},
		"password": {
			Condition: u.PasswordHash != nil,
			OnTrue:    func() { updater.SetPasswordHash(*u.PasswordHash) },
			OnFalse:   func() {},
		},
		"dept_id": {
			Condition: u.DeptID != nil,
			OnTrue:    func() { updater.SetNillableDeptID(u.DeptID) },
			OnFalse:   func() { updater.ClearDeptID() },
		},
	}.Emit(fieldsMask)
	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(updated), nil
}

func (rp *userRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.Platform.GetClient(ctx).User.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *userRP) List(ctx context.Context, in *UserListIn) (*UserListOut, error) {
	query := rp.Platform.GetClient(ctx).User.Query()
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), user.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), user.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtA), user.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtZ), user.UpdatedAtLTE)
	helper.Where(query, in.Email != nil, func() predicate.User { return user.EmailContainsFold(*in.Email) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"email":     func(sel helper.SQLSelector) { query.Order(sel(user.FieldEmail)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldUpdatedAt)) },
	})

	total, offset, limit, err := helper.ApplyPagination(ctx, query, in.PageRequest,
		helper.NewPageConfig(helper.DefaultPageSize, helper.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	us, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &UserListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(us, mapUser),
	}, nil
}

func (rp *userRP) FindByEmail(ctx context.Context, email string) (*User, error) {
	d, err := rp.Platform.GetClient(ctx).User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

func (rp *userRP) FindByID(ctx context.Context, id string) (*User, error) {
	d, err := rp.Platform.GetClient(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapUser(d *ent.User) *User {
	return &User{
		ID:           d.ID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		TenantID:     d.TenantID,
		DeptID:       d.DeptID,
		Email:        &d.Email,
		PasswordHash: &d.PasswordHash,
	}
}
