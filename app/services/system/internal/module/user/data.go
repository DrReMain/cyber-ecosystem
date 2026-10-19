package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	entsql "entgo.io/ent/dialect/sql"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/credential"
	"cyber-ecosystem/app/services/system/internal/ent/membership"
	"cyber-ecosystem/app/services/system/internal/ent/population"
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
		SetAccount(*u.Account).
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
	// The admin face lists the caller's tenant's members — seats, not
	// birthplace: an identity seated here is visible regardless of where
	// the row originates. A subjectless context (seed, framework paths)
	// lists the registry.
	if subject, ok := security.SubjectFromCtx(ctx); ok {
		query.Where(seatNarrowing(subject.TenantID))
	}
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), user.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), user.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtA), user.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtZ), user.UpdatedAtLTE)
	helper.Where(query, in.Account != nil, func() predicate.User { return user.AccountContainsFold(*in.Account) })
	helper.Where(query, in.Enabled != nil, func() predicate.User { return user.EnabledEQ(*in.Enabled) })
	helper.Where(query, in.DeptID != nil, func() predicate.User { return user.DeptIDEQ(*in.DeptID) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"account":   func(sel helper.SQLSelector) { query.Order(sel(user.FieldAccount)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldUpdatedAt)) },
	})

	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
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

func (rp *userRP) ListByPopulation(ctx context.Context, in *UserListIn) (*UserListOut, error) {
	query := rp.Platform.GetClient(ctx).User.Query()
	if subject, ok := security.SubjectFromCtx(ctx); ok {
		query.Where(populationNarrowing(subject.TenantID, utils.Deref(in.Population, "")))
	}
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"account":   func(sel helper.SQLSelector) { query.Order(sel(user.FieldAccount)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(user.FieldUpdatedAt)) },
	})

	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
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

func (rp *userRP) FindByAccount(ctx context.Context, account string) (*User, error) {
	// Registry-plane lookup: the plane flag must ride the query execution
	// too — the tenant interceptor fires at Only/All with the context it is
	// handed.
	globctx := security.RegistryPlane(ctx)
	d, err := rp.Platform.GetClient(globctx).User.Query().
		Where(user.AccountEQ(account)).
		Only(globctx)
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

func (rp *userRP) FindSeated(ctx context.Context, id string) (*User, error) {
	query := rp.Platform.GetClient(ctx).User.Query().Where(user.IDEQ(id))
	if subject, ok := security.SubjectFromCtx(ctx); ok {
		query.Where(seatNarrowing(subject.TenantID))
	}
	d, err := query.Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapUser(d), nil
}

func (rp *userRP) FindByIDs(ctx context.Context, ids []string) (map[string]*User, error) {
	rows, err := rp.Platform.GetClient(ctx).User.Query().Where(user.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	byID := make(map[string]*User, len(rows))
	for _, d := range rows {
		byID[d.ID] = mapUser(d)
	}
	return byID, nil
}

func (rp *userRP) FindCredential(ctx context.Context, kind, identifier string) (*Credential, error) {
	c, err := rp.Platform.GetClient(ctx).Credential.Query().
		Where(credential.Kind(kind), credential.Identifier(identifier)).
		Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapCredential(c), nil
}

func (rp *userRP) CreateCredential(ctx context.Context, c *Credential) error {
	if _, err := rp.Platform.GetClient(ctx).Credential.Create().
		SetKind(c.Kind).SetIdentifier(c.Identifier).SetUserID(c.UserID).
		Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *userRP) DeleteCredentialsByUser(ctx context.Context, userID string) error {
	if _, err := rp.Platform.GetClient(ctx).Credential.Delete().Where(credential.UserID(userID)).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *userRP) EnsureSeat(ctx context.Context, seat *Seat) error {
	// the unique index is the idempotence; a lost race is a no-op
	if err := rp.Platform.GetClient(ctx).Membership.Create().
		SetUserID(seat.UserID).SetTenantID(seat.TenantID).SetPopulation(seat.Population).
		OnConflictColumns(membership.FieldUserID, membership.FieldTenantID, membership.FieldPopulation).
		DoNothing().
		Exec(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *userRP) DeleteSeatsByUser(ctx context.Context, userID string) error {
	if _, err := rp.Platform.GetClient(ctx).Membership.Delete().Where(membership.UserID(userID)).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *userRP) ListSeatsByUser(ctx context.Context, userID string) ([]*Seat, error) {
	rows, err := rp.Platform.GetClient(ctx).Membership.Query().Where(membership.UserID(userID)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapSeat), nil
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
		if err := client.PrincipalRole.Create().
			SetPrincipalType(shared.PrincipalTypeUser).
			SetPrincipalID(userID).
			SetRoleID(id).
			OnConflictColumns(principalrole.FieldPrincipalType, principalrole.FieldPrincipalID, principalrole.FieldRoleID).
			DoNothing().
			Exec(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return rp.Platform.HandleEntError(err)
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

func seatNarrowing(tenantID string) predicate.User {
	return predicate.User(func(s *entsql.Selector) {
		m := entsql.Table(membership.Table)
		pops := entsql.Table(population.Table)
		s.Where(entsql.Exists(entsql.Select(m.C(membership.FieldUserID)).From(m).Where(entsql.And(
			entsql.EQ(m.C(membership.FieldTenantID), tenantID),
			entsql.EQ(m.C(membership.FieldUserID), entsql.Raw(s.C(user.FieldID))),
			entsql.NotIn(
				m.C(membership.FieldPopulation),
				entsql.Select(pops.C(population.FieldCode)).From(pops).Where(
					entsql.EQ(pops.C(population.FieldPlane), security.PlaneMember),
				),
			),
		))))
	})
}

func populationNarrowing(tenantID, populationCode string) predicate.User {
	return predicate.User(func(s *entsql.Selector) {
		m := entsql.Table(membership.Table)
		s.Where(entsql.Exists(entsql.Select(m.C(membership.FieldUserID)).From(m).Where(entsql.And(
			entsql.EQ(m.C(membership.FieldTenantID), tenantID),
			entsql.EQ(m.C(membership.FieldUserID), entsql.Raw(s.C(user.FieldID))),
			entsql.EQ(m.C(membership.FieldPopulation), populationCode),
		))))
	})
}

func mapUser(d *ent.User) *User {
	return &User{
		ID:           d.ID,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		DeptID:       d.DeptID,
		Avatar:       d.Avatar,
		Account:      &d.Account,
		PasswordHash: &d.PasswordHash,
		Enabled:      d.Enabled,
	}
}

func mapCredential(d *ent.Credential) *Credential {
	return &Credential{
		Kind:       d.Kind,
		Identifier: d.Identifier,
		UserID:     d.UserID,
	}
}

func mapSeat(d *ent.Membership) *Seat {
	return &Seat{
		UserID:     d.UserID,
		TenantID:   d.TenantID,
		Population: d.Population,
	}
}
