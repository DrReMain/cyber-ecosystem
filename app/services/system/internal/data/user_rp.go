package data

import (
	"context"
	"log/slog"

	"cyber-ecosystem/app/services/system/internal/biz"
	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
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

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *userRP) Create(ctx context.Context, u *biz.User) (*biz.User, error) {
	created, err := rp.platform.GetClient(ctx).User.Create().
		SetTenantID(u.TenantID).
		SetNillableDeptID(u.DeptID).
		SetEmail(u.Email).
		SetPasswordHash(u.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(created), nil
}

func (rp *userRP) FindByEmail(ctx context.Context, email string) (*biz.User, error) {
	d, err := rp.platform.GetClient(ctx).User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

func (rp *userRP) FindByID(ctx context.Context, id string) (*biz.User, error) {
	d, err := rp.platform.GetClient(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapUser(d *ent.User) *biz.User {
	return &biz.User{
		ID:           d.ID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		TenantID:     d.TenantID,
		DeptID:       d.DeptID,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
	}
}
