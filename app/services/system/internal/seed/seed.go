package seed

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/permission"
	"cyber-ecosystem/app/services/system/internal/ent/principalrole"
	"cyber-ecosystem/app/services/system/internal/ent/role"
	"cyber-ecosystem/app/services/system/internal/ent/user"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	superAdminRoleCode = "superadmin"
	superAdminPattern  = "/*" // catch-all grant; no superuser code path exists
	superAdminScope    = "all"
)

type Seed struct {
	log  *slog.Logger
	plat *platform.Platform
	cfg  *conf.Authz
}

func NewSeed(cfg *conf.Authz, p *platform.Platform, logger *slog.Logger) (*Seed, error) {
	return &Seed{log: logger, plat: p, cfg: cfg}, nil
}

// Run --------------------------------------------------------------------------------------------------------------

func (s *Seed) Run(ctx context.Context) error {
	return s.seedSuperAdmins(ctx)
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *Seed) seedSuperAdmins(ctx context.Context) error {
	// Without this floor a fresh deployment denies every request — including
	// the RPC that would create the first user — with no way out.
	if s.cfg == nil || len(s.cfg.GetSuperAdmins()) == 0 {
		return nil
	}
	if s.cfg.GetInitialPassword() == "" {
		return fmt.Errorf("seed: super_admins configured but initial_password is empty")
	}
	client := s.plat.GetClient(ctx)
	tenant := security.DefaultTenant

	superRole, err := s.ensureRole(ctx, client, tenant)
	if err != nil {
		return err
	}
	if err := s.ensureSuperGrant(ctx, client, tenant, superRole.ID); err != nil {
		return err
	}
	hash, err := utils.Hash(s.cfg.GetInitialPassword())
	if err != nil {
		return fmt.Errorf("seed: hash initial password: %w", err)
	}
	for _, email := range s.cfg.GetSuperAdmins() {
		u, err := s.ensureUser(ctx, client, tenant, email, hash)
		if err != nil {
			return err
		}
		if err := s.ensureBinding(ctx, client, tenant, u.ID, superRole.ID); err != nil {
			return err
		}
	}
	// Bump + notify so every replica — including this one — recompiles from
	// the seeded tables.
	shared.NotifyPolicyChanged(ctx, s.plat.GetCache(), s.log)
	return nil
}

func (s *Seed) ensureRole(ctx context.Context, client *ent.Client, tenant string) (*ent.Role, error) {
	r, err := client.Role.Query().Where(role.TenantID(tenant), role.Code(superAdminRoleCode)).Only(ctx)
	if r != nil {
		return r, nil
	}
	if !ent.IsNotFound(err) {
		return nil, s.plat.HandleEntError(fmt.Errorf("seed: query superadmin role: %w", err))
	}
	if r, err = client.Role.Create().SetTenantID(tenant).SetCode(superAdminRoleCode).SetName("Super Admin").Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			if r, err = client.Role.Query().Where(role.TenantID(tenant), role.Code(superAdminRoleCode)).Only(ctx); err != nil {
				return nil, s.plat.HandleEntError(fmt.Errorf("seed: requery superadmin role: %w", err))
			}
			return r, nil // racing replica won the create
		}
		return nil, s.plat.HandleEntError(fmt.Errorf("seed: create superadmin role: %w", err))
	}
	return r, nil
}

func (s *Seed) ensureSuperGrant(ctx context.Context, client *ent.Client, tenant, roleID string) error {
	p, err := client.Permission.Query().Where(permission.RoleID(roleID), permission.Operation(superAdminPattern)).Only(ctx)
	if p != nil {
		return nil
	}
	if !ent.IsNotFound(err) {
		return s.plat.HandleEntError(fmt.Errorf("seed: query superadmin grant: %w", err))
	}
	if _, err = client.Permission.Create().SetTenantID(tenant).SetRoleID(roleID).SetOperation(superAdminPattern).SetScopeKind(superAdminScope).Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil // racing replica won the create
		}
		return s.plat.HandleEntError(fmt.Errorf("seed: create superadmin grant: %w", err))
	}
	return nil
}

func (s *Seed) ensureUser(ctx context.Context, client *ent.Client, tenant, email, hash string) (*ent.User, error) {
	u, err := client.User.Query().Where(user.TenantID(tenant), user.Email(email)).Only(ctx)
	if u != nil {
		return u, nil // password is owned by admin flows, never reset here
	}
	if !ent.IsNotFound(err) {
		return nil, s.plat.HandleEntError(fmt.Errorf("seed: query user %q: %w", email, err))
	}
	if u, err = client.User.Create().SetTenantID(tenant).SetEmail(email).SetPasswordHash(hash).Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			if u, err = client.User.Query().Where(user.TenantID(tenant), user.Email(email)).Only(ctx); err != nil {
				return nil, s.plat.HandleEntError(fmt.Errorf("seed: requery user %q: %w", email, err))
			}
			return u, nil // racing replica won the create
		}
		return nil, s.plat.HandleEntError(fmt.Errorf("seed: create user %q: %w", email, err))
	}
	s.log.Info("seed: super admin created", "email", email)
	return u, nil
}

func (s *Seed) ensureBinding(ctx context.Context, client *ent.Client, tenant, userID, roleID string) error {
	b, err := client.PrincipalRole.Query().Where(
		principalrole.PrincipalType(shared.PrincipalTypeUser),
		principalrole.PrincipalID(userID),
		principalrole.RoleID(roleID),
	).Only(ctx)
	if b != nil {
		return nil
	}
	if !ent.IsNotFound(err) {
		return s.plat.HandleEntError(fmt.Errorf("seed: query binding: %w", err))
	}
	if _, err = client.PrincipalRole.Create().SetTenantID(tenant).SetPrincipalType(shared.PrincipalTypeUser).SetPrincipalID(userID).SetRoleID(roleID).Save(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil // racing replica won the create
		}
		return s.plat.HandleEntError(fmt.Errorf("seed: create binding: %w", err))
	}
	return nil
}
