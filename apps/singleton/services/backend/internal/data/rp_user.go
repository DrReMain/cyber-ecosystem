package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/predicate"
	entuser "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/user"
)

type userRP struct {
	RP
}

func NewUserRP(logger log.Logger, store *Store) biz.UserRP {
	return &userRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_user")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *userRP) Create(ctx context.Context, user *biz.User) error {
	// Application-level check for friendly 409 error (DB partial unique index is the final safety net)
	exists, err := rp.store.GetClient(ctx).User.Query().
		Where(entuser.EmailEQ(*user.Email)).
		Exist(ctx)
	if err != nil {
		return HandleError(err)
	}
	if exists {
		return singletonV1.ErrorErrorReasonEmailAlreadyExists("")
	}
	if err := rp.store.GetClient(ctx).User.Create().
		SetPassword(*user.PasswordCipher).
		SetEmail(*user.Email).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *userRP) Update(ctx context.Context, fieldsMask []string, user *biz.User) error {
	updater := rp.store.GetClient(ctx).User.UpdateOneID(*user.ID)
	utils.Handler{
		"password": {
			Condition: user.PasswordCipher != nil,
			OnTrue:    func() { updater.SetPassword(*user.PasswordCipher) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)
	if err := updater.Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *userRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).User.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *userRP) Get(ctx context.Context, id string) (*biz.User, error) {
	d, err := rp.store.GetClient(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapUser(d), nil
}

func (rp *userRP) Query(ctx context.Context, in *biz.UserQueryIn) (*biz.UserQueryOut, error) {
	query := rp.store.GetClient(ctx).User.Query()
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtA), entuser.CreatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtZ), entuser.CreatedAtLTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtA), entuser.UpdatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtZ), entuser.UpdatedAtLTE)
	entutil.Where(query, in.ID != nil, func() predicate.User { return entuser.IDEQ(*in.ID) })
	entutil.Where(query, in.Email != nil, func() predicate.User { return entuser.EmailEQ(*in.Email) })
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entuser.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(entuser.FieldUpdatedAt)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		singletonV1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}
	pos, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return &biz.UserQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapUser),
	}, nil
}

func (rp *userRP) FindByEmail(ctx context.Context, email string) (*biz.User, error) {
	d, err := rp.store.GetClient(ctx).User.Query().
		Where(entuser.EmailEQ(email)).
		First(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return mapUser(d), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapUser(d *ent.User) *biz.User {
	return &biz.User{
		ID:             &d.ID,
		CreatedAt:      &d.CreatedAt,
		UpdatedAt:      &d.UpdatedAt,
		PasswordCipher: &d.Password,
		Email:          &d.Email,
	}
}
