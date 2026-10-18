package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/dept"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/module/population"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// errBirthLost marks the one retryable failure in Create: a concurrent
// birth won the account's unique. Internal control flow, never returned.
var errBirthLost = fmt.Errorf("identity birth lost the account unique")

// DO ------------------------------------------------------------------------------------------------------------------

type User struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeptID       *string
	Avatar       *string // referenced file id; nil on Update-mask = clear
	Account      *string
	PasswordHash *string  // biz-only (never in proto); nil on Update = leave unchanged
	Roles        []string // role codes; list enrichment, filled by List/Get
	Enabled      bool
}

// Credential is a third-party identity binding: one external identifier
// resolves to exactly one global identity. Identity-plane data — the kind
// vocabulary is code (provider_<kind>.go), instances live in conf rows.
type Credential struct {
	Kind       string
	Identifier string
	UserID     string
}

// Seat is one tenancy of an identity — a membership row. The identity is
// global (account); every tenancy it holds is a seat.
type Seat struct {
	UserID     string
	TenantID   string
	Population string
}

type UserListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Account *string
	Enabled *bool
	DeptID  *string
}

type UserListOut struct {
	*commonpb.PageResponse
	List []*User
}

// Port ----------------------------------------------------------------------------------------------------------------

type UserRP interface {
	Create(ctx context.Context, u *User) (*User, error)
	Update(ctx context.Context, fieldsMask []string, u *User) (*User, error)
	UpdateStatus(ctx context.Context, id string, enabled bool) (*User, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *UserListIn) (*UserListOut, error)
	FindByAccount(ctx context.Context, account string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindSeated(ctx context.Context, id string) (*User, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*User, error)
	FindCredential(ctx context.Context, kind, identifier string) (*Credential, error)
	CreateCredential(ctx context.Context, c *Credential) error
	DeleteCredentialsByUser(ctx context.Context, userID string) error
	EnsureSeat(ctx context.Context, seat *Seat) error
	DeleteSeatsByUser(ctx context.Context, userID string) error
	ListSeatsByUser(ctx context.Context, userID string) ([]*Seat, error)
	FindRoleIDsByCodes(ctx context.Context, codes []string) (map[string]string, error)
	SetUserRoles(ctx context.Context, userID string, roleIDs []string) error
	DeleteUserRoles(ctx context.Context, userID string) (bool, error)
	ListUserRoles(ctx context.Context, userIDs []string) (map[string][]string, error)
	NotifyChanged(ctx context.Context)
	RevokeUserSessions(ctx context.Context, userID string) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type UserUC struct {
	shared.UC
	userRP       UserRP
	deptRP       dept.DeptRP
	fileRP       file.FileRP
	populationRP population.PopulationRP
}

func NewUserUC(logger *slog.Logger, tm shared.Transaction, userRP UserRP, deptRP dept.DeptRP, fileRP file.FileRP, populationRP population.PopulationRP) *UserUC {
	return &UserUC{
		UC:           shared.NewUC(logger.With("module", "module/user"), tm),
		userRP:       userRP,
		deptRP:       deptRP,
		fileRP:       fileRP,
		populationRP: populationRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *UserUC) Create(ctx context.Context, account, password, deptID, avatar *string, population string, roles []string) (out *User, err error) {
	seatPopulation, err := uc.resolveSeatPopulation(ctx, population)
	if err != nil {
		return nil, err
	}
	roleIDs, err := uc.resolveRoles(ctx, roles)
	if err != nil {
		return nil, err
	}
	// Account is the global identity key: an identity that already exists
	// gains a seat in the creating admin's tenant — a second row is never
	// born. Profile params (password, dept, avatar) belong to identity
	// birth alone and are inert on the seat path.
	existing, ferr := uc.userRP.FindByAccount(ctx, *account)
	if ferr != nil && !errorspb.IsInfraErrorDbNotFound(ferr) {
		return nil, ferr
	}
	birth := ferr != nil
	if birth {
		if err = uc.ensureDept(ctx, deptID); err != nil {
			return nil, err
		}
		if err = uc.ensureAvatar(ctx, avatar); err != nil {
			return nil, err
		}
		hash, herr := utils.Hash(*password)
		if herr != nil {
			return nil, errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("password rejected: %w", herr))
		}
		password = &hash
	} else {
		out = existing
	}
	// The seat lands in the creating admin's tenant; an identity born here
	// is never seatless.
	seat := &Seat{TenantID: security.TenantFromCtx(ctx), Population: seatPopulation}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if birth {
			created, cerr := uc.userRP.Create(ctx, &User{DeptID: deptID, Avatar: avatar, Account: account, PasswordHash: password})
			if cerr != nil {
				if errorspb.IsInfraErrorDbConstraint(cerr) {
					return errBirthLost
				}
				return cerr
			}
			out = created
		}
		seat.UserID = out.ID
		if err = uc.userRP.EnsureSeat(ctx, seat); err != nil {
			return err
		}
		// An empty role ask leaves the identity's bindings untouched — the
		// seat is the ask, not a binding replace.
		if len(roleIDs) > 0 {
			return uc.userRP.SetUserRoles(ctx, out.ID, roleIDs)
		}
		return nil
	})
	if err != nil {
		// The winner exists now, so the same ask lands as its seat.
		if errors.Is(err, errBirthLost) {
			return uc.Create(ctx, account, password, deptID, avatar, population, roles)
		}
		return nil, err
	}
	if len(roleIDs) > 0 {
		uc.userRP.NotifyChanged(ctx)
	}
	return
}

func (uc *UserUC) Update(ctx context.Context, fieldsMask []string, u *User, password *string, roles []string) (out *User, err error) {
	if _, err = uc.ensureSeatedUser(ctx, u.ID); err != nil {
		return nil, err
	}
	if slices.Contains(fieldsMask, "dept_id") {
		if err = uc.ensureDept(ctx, u.DeptID); err != nil {
			return nil, err
		}
	}
	if slices.Contains(fieldsMask, "avatar") {
		if err = uc.ensureAvatar(ctx, u.Avatar); err != nil {
			return nil, err
		}
	}
	if password != nil {
		hash, e := utils.Hash(*password)
		if e != nil {
			return nil, errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("password rejected: %w", e))
		}
		u.PasswordHash = &hash
	}
	setRoles := slices.Contains(fieldsMask, "roles")
	var roleIDs []string
	if setRoles {
		if roleIDs, err = uc.resolveRoles(ctx, roles); err != nil {
			return nil, err
		}
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if out, err = uc.userRP.Update(ctx, fieldsMask, u); err != nil {
			return err
		}
		if setRoles {
			return uc.userRP.SetUserRoles(ctx, u.ID, roleIDs)
		}
		return nil
	})
	if err == nil && setRoles {
		uc.userRP.NotifyChanged(ctx)
	}
	return
}

func (uc *UserUC) UpdateStatus(ctx context.Context, id string, enabled bool) (out *User, err error) {
	if subject, ok := security.SubjectFromCtx(ctx); ok && !enabled && subject.UserID == id {
		return nil, systempb.ErrorSystemUserSelfDisable("")
	}
	if _, err = uc.ensureSeatedUser(ctx, id); err != nil {
		return nil, err
	}
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.UpdateStatus(ctx, id, enabled)
		return err
	})
	if err != nil {
		return nil, err
	}
	uc.userRP.NotifyChanged(ctx)
	if !enabled {
		// Best-effort: the authz snapshot already strips a disabled user's
		// grants, so a cache blip here must not block the disable itself.
		if rerr := uc.userRP.RevokeUserSessions(ctx, id); rerr != nil {
			uc.Log.Warn("revoke sessions after disable failed", "user_id", id, "err", rerr)
		}
	}
	return out, nil
}

func (uc *UserUC) ChangePassword(ctx context.Context, oldPassword, newPassword string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("change password requires an authenticated caller"))
	}
	u, err := uc.ensureUser(ctx, subject.UserID)
	if err != nil {
		return err
	}
	if !utils.Verify(oldPassword, *u.PasswordHash) {
		return systempb.ErrorSystemPasswordMismatch("").WithCause(fmt.Errorf("old password mismatch: user %s", subject.UserID))
	}
	hash, err := utils.Hash(newPassword)
	if err != nil {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("password rejected: %w", err))
	}
	if _, err = uc.userRP.Update(ctx, []string{"password"}, &User{ID: subject.UserID, PasswordHash: &hash}); err != nil {
		return err
	}
	return nil
}

func (uc *UserUC) Delete(ctx context.Context, id string) (out string, err error) {
	if subject, ok := security.SubjectFromCtx(ctx); ok && subject.UserID == id {
		return "", systempb.ErrorSystemUserSelfDelete("")
	}
	if _, err = uc.ensureSeatedUser(ctx, id); err != nil {
		return "", err
	}
	hadRoles := false
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if hadRoles, err = uc.userRP.DeleteUserRoles(ctx, id); err != nil {
			return err
		}
		if err = uc.userRP.DeleteSeatsByUser(ctx, id); err != nil {
			return err
		}
		if err = uc.userRP.DeleteCredentialsByUser(ctx, id); err != nil {
			return err
		}
		out, err = uc.userRP.Delete(ctx, id)
		return err
	})
	if err == nil && hadRoles {
		uc.userRP.NotifyChanged(ctx)
	}
	if err == nil {
		// Best-effort, same trade-off as the disable hook above.
		if rerr := uc.userRP.RevokeUserSessions(ctx, id); rerr != nil {
			uc.Log.Warn("revoke sessions after delete failed", "user_id", id, "err", rerr)
		}
	}
	return
}

func (uc *UserUC) List(ctx context.Context, in *UserListIn) (*UserListOut, error) {
	out, err := uc.userRP.List(ctx, in)
	if err != nil {
		return nil, err
	}
	if len(out.List) == 0 {
		return out, nil
	}
	ids := make([]string, len(out.List))
	for i, u := range out.List {
		ids[i] = u.ID
	}
	byUser, err := uc.userRP.ListUserRoles(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, u := range out.List {
		u.Roles = byUser[u.ID]
	}
	return out, nil
}

func (uc *UserUC) Get(ctx context.Context, id string) (*User, error) {
	out, err := uc.ensureSeatedUser(ctx, id)
	if err != nil {
		return nil, err
	}
	byUser, err := uc.userRP.ListUserRoles(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	out.Roles = byUser[id]
	return out, nil
}

func (uc *UserUC) BatchGet(ctx context.Context, ids []string) (map[string]*User, error) {
	return uc.userRP.FindByIDs(ctx, ids)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *UserUC) ensureUser(ctx context.Context, id string) (*User, error) {
	u, err := uc.userRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemUserNotFound("")
		}
		return nil, err
	}
	return u, nil
}

func (uc *UserUC) ensureSeatedUser(ctx context.Context, id string) (*User, error) {
	u, err := uc.userRP.FindSeated(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemUserNotFound("")
		}
		return nil, err
	}
	return u, nil
}

func (uc *UserUC) ensureAvatar(ctx context.Context, avatar *string) error {
	if avatar == nil {
		return nil
	}
	f, err := uc.fileRP.FindRef(ctx, *avatar)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return systempb.ErrorSystemFileNotFound("")
		}
		return err
	}
	if f.Status != file.StatusConfirmed {
		return systempb.ErrorSystemFileInvalidState("")
	}
	return nil
}

func (uc *UserUC) resolveSeatPopulation(ctx context.Context, population string) (string, error) {
	if population == "" {
		return security.DefaultPopulation, nil
	}
	if _, err := uc.populationRP.FindByCode(ctx, population); err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return "", systempb.ErrorSystemPopulationNotFound("")
		}
		return "", err
	}
	return population, nil
}

func (uc *UserUC) resolveRoles(ctx context.Context, codes []string) ([]string, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	byCode, err := uc.userRP.FindRoleIDsByCodes(ctx, codes)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(codes))
	for _, c := range codes {
		id, ok := byCode[c]
		if !ok {
			return nil, systempb.ErrorSystemRoleNotFound("").WithCause(fmt.Errorf("unknown role code %q", c))
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (uc *UserUC) ensureDept(ctx context.Context, deptID *string) error {
	if deptID == nil {
		return nil
	}
	if _, err := uc.deptRP.Get(ctx, *deptID); err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return systempb.ErrorSystemDeptNotFound("")
		}
		return err
	}
	return nil
}
