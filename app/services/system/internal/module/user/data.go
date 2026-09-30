package user

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/principalrole"
	"cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type userRP struct {
	shared.RP
}

func NewUserRP(logger *slog.Logger, p *platform.Platform) UserRP {
	return &userRP{
		RP: shared.NewRP(logger.With("module", "module/user_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *userRP) Create(ctx context.Context, u *User) (*User, error) {
	created, err := rp.Platform.GetClient(ctx).User.Create().
		SetNillableDeptID(u.DeptID).
		SetNillableAvatar(u.Avatar).
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
		"avatar": {
			Condition: u.Avatar != nil,
			OnTrue:    func() { updater.SetNillableAvatar(u.Avatar) },
			OnFalse:   func() { updater.ClearAvatar() },
		},
	}.Emit(fieldsMask)
	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(updated), nil
}

func (rp *userRP) UpdateStatus(ctx context.Context, id string, enabled bool) (*User, error) {
	updated, err := rp.Platform.GetClient(ctx).User.UpdateOneID(id).SetEnabled(enabled).Save(ctx)
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
	helper.Where(query, in.Enabled != nil, func() predicate.User { return user.EnabledEQ(*in.Enabled) })
	helper.Where(query, in.DeptID != nil, func() predicate.User { return user.DeptIDEQ(*in.DeptID) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"email":     func(sel helper.SQLSelector) { query.Order(sel(user.FieldEmail)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldUpdatedAt)) },
	})

	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
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

func (rp *userRP) FindRoleIDsByCodes(ctx context.Context, codes []string) (map[string]string, error) {
	rows, err := rp.Platform.GetClient(ctx).Role.Query().Where(role.CodeIn(codes...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	byCode := make(map[string]string, len(rows))
	for _, r := range rows {
		byCode[r.Code] = r.ID
	}
	return byCode, nil
}

func (rp *userRP) SetUserRoles(ctx context.Context, userID string, wanted []string) error {
	client := rp.Platform.GetClient(ctx)
	stored, err := client.PrincipalRole.Query().Where(
		principalrole.PrincipalType(shared.PrincipalTypeUser),
		principalrole.PrincipalID(userID),
	).All(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	storedByID := make(map[string]struct{}, len(stored))
	for _, b := range stored {
		storedByID[b.RoleID] = struct{}{}
	}
	wantedSet := make(map[string]struct{}, len(wanted))
	for _, id := range wanted {
		wantedSet[id] = struct{}{}
	}
	for _, id := range wanted {
		if _, ok := storedByID[id]; ok {
			continue
		}
		if _, err := client.PrincipalRole.Create().
			SetPrincipalType(shared.PrincipalTypeUser).
			SetPrincipalID(userID).
			SetRoleID(id).
			Save(ctx); err != nil {
			// A racing writer won the unique; the binding exists either way.
			if !ent.IsConstraintError(err) {
				return rp.Platform.HandleEntError(err)
			}
		}
	}
	for _, b := range stored {
		if _, ok := wantedSet[b.RoleID]; ok {
			continue
		}
		if err := client.PrincipalRole.DeleteOneID(b.ID).Exec(ctx); err != nil {
			return rp.Platform.HandleEntError(err)
		}
	}
	return nil
}

func (rp *userRP) DeleteUserRoles(ctx context.Context, userID string) (bool, error) {
	n, err := rp.Platform.GetClient(ctx).PrincipalRole.Delete().Where(
		principalrole.PrincipalType(shared.PrincipalTypeUser),
		principalrole.PrincipalID(userID),
	).Exec(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return n > 0, nil
}

func (rp *userRP) ListUserRoles(ctx context.Context, userIDs []string) (map[string][]string, error) {
	client := rp.Platform.GetClient(ctx)
	bindings, err := client.PrincipalRole.Query().Where(
		principalrole.PrincipalType(shared.PrincipalTypeUser),
		principalrole.PrincipalIDIn(userIDs...),
	).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	roleIDs := make([]string, 0, len(bindings))
	for _, b := range bindings {
		roleIDs = append(roleIDs, b.RoleID)
	}
	codesByID := make(map[string]string)
	if len(roleIDs) > 0 {
		rows, err := client.Role.Query().Where(role.IDIn(roleIDs...)).All(ctx)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
		for _, r := range rows {
			codesByID[r.ID] = r.Code
		}
	}
	byUser := make(map[string][]string, len(userIDs))
	for _, b := range bindings {
		if code, ok := codesByID[b.RoleID]; ok {
			byUser[b.PrincipalID] = append(byUser[b.PrincipalID], code)
		}
	}
	return byUser, nil
}

func (rp *userRP) NotifyChanged(ctx context.Context) {
	shared.NotifyPolicyChanged(ctx, rp.Platform.GetCache(), rp.Log)
}

func (rp *userRP) RevokeUserSessions(ctx context.Context, userID string) error {
	if err := shared.RevokeSessions(ctx, rp.Platform.GetCache(), userID); err != nil {
		return rp.Platform.HandleCacheError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapUser(d *ent.User) *User {
	return &User{
		ID:           d.ID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		TenantID:     d.TenantID,
		DeptID:       d.DeptID,
		Avatar:       d.Avatar,
		Email:        &d.Email,
		PasswordHash: &d.PasswordHash,
		Enabled:      d.Enabled,
	}
}
