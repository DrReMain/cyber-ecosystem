package data

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/conf"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entpolicyrule "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/policyrule"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/predicate"
	entrole "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/role"
	entuser "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/user"
)

//go:embed casbin_model/model.conf
var modelFS embed.FS

const policyNotifyChannel = "casbin:policy:notify"

type policyRP struct {
	RP
	domain   string
	enforcer *casbin.SyncedEnforcer
}

func NewPolicyRP(cs *conf.Super, logger log.Logger, store *Store) (biz.PolicyRP, func(), error) {
	rp := &policyRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_policy")),
			store: store,
		},
		domain: biz.Domain,
	}

	if cs != nil && cs.GetEnabled() {
		if err := store.InTx(context.Background(), func(ctx context.Context) error {
			if err := rp.ensureRole(ctx, cs.GetRoleName(), cs.GetRoleCode(), 1); err != nil {
				return err
			}
			userID, err := rp.ensureUser(ctx, cs.GetEmail(), cs.GetPassword(), cs.GetForceReset())
			if err != nil {
				return err
			}
			if err := rp.ensurePolicy(ctx, "g", []string{userID, cs.GetRoleCode(), biz.Domain}); err != nil {
				return err
			}
			if err := rp.ensurePolicy(ctx, "p", []string{cs.GetRoleCode(), biz.Domain, "/*", "allow"}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	modelText, err := modelFS.ReadFile("casbin_model/model.conf")
	if err != nil {
		return nil, nil, fmt.Errorf("read casbin model: %w", err)
	}
	mdl, err := model.NewModelFromString(string(modelText))
	if err != nil {
		return nil, nil, fmt.Errorf("parse casbin model: %w", err)
	}

	e, err := casbin.NewSyncedEnforcer(mdl, rp)
	if err != nil {
		return nil, nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	rp.enforcer = e

	e.SetLogger(newPolicyLogger(rp.log))

	if err := e.LoadPolicy(); err != nil {
		return nil, nil, fmt.Errorf("load casbin policy: %w", err)
	}

	// Set up Redis Pub/Sub for multi-instance policy sync.
	// Only active when the cache backend is Redis (multi-instance deployment).
	// Memory cache = single instance = no sync needed.
	cleanup := func() {}
	if c := store.GetCache(); c != nil {
		if rc, ok := c.Client.(*redis.Client); ok {
			subCtx, subCancel := context.WithCancel(context.Background())
			sub := rc.Subscribe(subCtx, policyNotifyChannel)

			go func() {
				ch := sub.Channel()
				for range ch {
					if err := e.LoadPolicy(); err != nil {
						rp.log.Errorf("failed to reload policy from notification: %v", err)
					}
				}
				subCancel()
			}()

			cleanup = func() {
				subCancel()
				sub.Close()
			}
		}
	}
	return rp, cleanup, nil
}

func (rp *policyRP) notifyPolicyChange() {
	if c := rp.store.GetCache(); c != nil {
		if rc, ok := c.Client.(*redis.Client); ok {
			if err := rc.Publish(context.Background(), policyNotifyChannel, "reload").Err(); err != nil {
				rp.log.Errorf("failed to publish policy change notification: %v", err)
			}
		}
	}
}

func (rp *policyRP) syncAndNotify() {
	if err := rp.enforcer.LoadPolicy(); err != nil {
		rp.log.Errorf("failed to reload policy: %v", err)
	}
	rp.notifyPolicyChange()
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *policyRP) Enforce(sub, dom, obj string) (bool, error) {
	return rp.enforcer.Enforce(sub, dom, obj)
}

func (rp *policyRP) AddRoleForUser(ctx context.Context, userID, role string) (bool, func(), error) {
	rule := []string{userID, role, rp.domain}
	has, _ := rp.enforcer.HasGroupingPolicy(toAnySlice(rule)...)
	if has {
		return false, func() {}, nil
	}
	if err := rp.createPolicyRow(ctx, "g", rule); err != nil {
		return false, nil, err
	}
	return true, rp.syncAndNotify, nil
}

func (rp *policyRP) RemoveRoleForUser(ctx context.Context, userID, role string) (bool, func(), error) {
	rule := []string{userID, role, rp.domain}
	has, _ := rp.enforcer.HasGroupingPolicy(toAnySlice(rule)...)
	if !has {
		return false, func() {}, nil
	}
	if err := rp.deletePolicyRow(ctx, "g", rule); err != nil {
		return false, nil, err
	}
	return true, rp.syncAndNotify, nil
}

func (rp *policyRP) GetRolesForUser(userID string) []string {
	return rp.enforcer.GetRolesForUserInDomain(userID, rp.domain)
}

func (rp *policyRP) GetUsersForRole(role string) []string {
	return rp.enforcer.GetUsersForRoleInDomain(role, rp.domain)
}

func (rp *policyRP) AddPermissionForRole(ctx context.Context, role, object, effect string) (bool, func(), error) {
	rule := []string{role, rp.domain, object, effect}
	has, _ := rp.enforcer.HasPolicy(toAnySlice(rule)...)
	if has {
		return false, func() {}, nil
	}
	if err := rp.createPolicyRow(ctx, "p", rule); err != nil {
		return false, nil, err
	}
	return true, rp.syncAndNotify, nil
}

func (rp *policyRP) RemovePermissionForRole(ctx context.Context, role, object, effect string) (bool, func(), error) {
	rule := []string{role, rp.domain, object, effect}
	has, _ := rp.enforcer.HasPolicy(toAnySlice(rule)...)
	if !has {
		return false, func() {}, nil
	}
	if err := rp.deletePolicyRow(ctx, "p", rule); err != nil {
		return false, nil, err
	}
	return true, rp.syncAndNotify, nil
}

func (rp *policyRP) GetPermissionsForRole(role string) ([][]string, error) {
	return rp.enforcer.GetFilteredPolicy(0, role, rp.domain)
}

func (rp *policyRP) RemoveRoleGroupings(ctx context.Context, roleCode string) error {
	if err := rp.deleteFilteredPolicyRows(ctx, "g", 1, roleCode); err != nil {
		return err
	}
	rp.syncAndNotify()
	return nil
}

func (rp *policyRP) RemoveUserGroupings(ctx context.Context, userID string) error {
	if err := rp.deleteFilteredPolicyRows(ctx, "g", 0, userID); err != nil {
		return err
	}
	rp.syncAndNotify()
	return nil
}

func (rp *policyRP) RemoveRolePermissions(ctx context.Context, roleCode string) error {
	if err := rp.deleteFilteredPolicyRows(ctx, "p", 0, roleCode); err != nil {
		return err
	}
	rp.syncAndNotify()
	return nil
}

// Seed -----------------------------------------------------------------------------------------------------------------

func (rp *policyRP) ensureRole(ctx context.Context, name, code string, status uint8) error {
	client := rp.store.GetClient(ctx)
	exists, err := client.Role.Query().
		Where(entrole.CodeEQ(code)).
		Exist(ctx)
	if err != nil {
		return HandleError(err)
	}
	if exists {
		return nil
	}
	_, err = client.Role.Create().
		SetName(name).
		SetCode(code).
		SetStatus(status).
		Save(ctx)
	return HandleError(err)
}

func (rp *policyRP) ensureUser(ctx context.Context, email, passwordPlain string, forceReset bool) (string, error) {
	client := rp.store.GetClient(ctx)
	user, err := client.User.Query().
		Where(entuser.EmailEQ(email)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return "", HandleError(err)
	}
	if ent.IsNotFound(err) {
		cipher := utils.EncryptGenerate(passwordPlain)
		user, err = client.User.Create().
			SetEmail(email).
			SetPassword(cipher).
			Save(ctx)
		if err != nil {
			return "", HandleError(err)
		}
		return user.ID, nil
	}
	if forceReset {
		cipher := utils.EncryptGenerate(passwordPlain)
		_, err = client.User.UpdateOneID(user.ID).
			SetPassword(cipher).
			Save(ctx)
		if err != nil {
			return "", HandleError(err)
		}
	}
	return user.ID, nil
}

func (rp *policyRP) ensurePolicy(ctx context.Context, ptype string, values []string) error {
	client := rp.store.GetClient(ctx)
	exists, err := client.PolicyRule.Query().Where(
		entpolicyrule.PtypeEQ(ptype),
		entpolicyrule.V0EQ(policyFieldOrDefault(values, 0)),
		entpolicyrule.V1EQ(policyFieldOrDefault(values, 1)),
		entpolicyrule.V2EQ(policyFieldOrDefault(values, 2)),
		entpolicyrule.V3EQ(policyFieldOrDefault(values, 3)),
	).Exist(ctx)
	if err != nil {
		return HandleError(err)
	}
	if exists {
		return nil
	}
	_, err = client.PolicyRule.Create().
		SetPtype(ptype).
		SetV0(policyFieldOrDefault(values, 0)).
		SetV1(policyFieldOrDefault(values, 1)).
		SetV2(policyFieldOrDefault(values, 2)).
		SetV3(policyFieldOrDefault(values, 3)).
		Save(ctx)
	return HandleError(err)
}

// Adapter --------------------------------------------------------------------------------------------------------------

func (rp *policyRP) LoadPolicy(mdl model.Model) error {
	ctx := context.Background()
	policies, err := rp.store.GetClient(ctx).PolicyRule.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("load policies from db: %w", err)
	}
	for _, p := range policies {
		line := policyLine(p.Ptype, p.V0, p.V1, p.V2, p.V3, p.V4, p.V5)
		persist.LoadPolicyLine(line, mdl)
	}
	return nil
}

func (rp *policyRP) SavePolicy(mdl model.Model) error {
	ctx := context.Background()
	if _, err := rp.store.GetClient(ctx).PolicyRule.Delete().Exec(ctx); err != nil {
		return fmt.Errorf("clear policies: %w", err)
	}
	for ptype, assertion := range mdl["p"] {
		for _, rule := range assertion.Policy {
			if err := rp.createPolicyRow(ctx, ptype, rule); err != nil {
				return fmt.Errorf("save policy %s: %w", ptype, err)
			}
		}
	}
	for ptype, assertion := range mdl["g"] {
		for _, rule := range assertion.Policy {
			if err := rp.createPolicyRow(ctx, ptype, rule); err != nil {
				return fmt.Errorf("save policy %s: %w", ptype, err)
			}
		}
	}
	return nil
}

func (rp *policyRP) AddPolicy(sec string, ptype string, rule []string) error {
	return rp.createPolicyRow(context.Background(), ptype, rule)
}

func (rp *policyRP) RemovePolicy(sec string, ptype string, rule []string) error {
	return rp.deletePolicyRow(context.Background(), ptype, rule)
}

func (rp *policyRP) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return rp.deleteFilteredPolicyRows(context.Background(), ptype, fieldIndex, fieldValues...)
}

// Helper ---------------------------------------------------------------------------------------------------------------

func (rp *policyRP) createPolicyRow(ctx context.Context, ptype string, values []string) error {
	client := rp.store.GetClient(ctx)
	exists, err := client.PolicyRule.Query().Where(
		entpolicyrule.Ptype(ptype),
		entpolicyrule.V0EQ(policyFieldOrDefault(values, 0)),
		entpolicyrule.V1EQ(policyFieldOrDefault(values, 1)),
		entpolicyrule.V2EQ(policyFieldOrDefault(values, 2)),
		entpolicyrule.V3EQ(policyFieldOrDefault(values, 3)),
	).Exist(ctx)
	if err != nil {
		return HandleError(err)
	}
	if exists {
		return nil
	}
	_, err = client.PolicyRule.Create().
		SetPtype(ptype).
		SetV0(policyFieldOrDefault(values, 0)).
		SetV1(policyFieldOrDefault(values, 1)).
		SetV2(policyFieldOrDefault(values, 2)).
		SetV3(policyFieldOrDefault(values, 3)).
		SetV4(policyFieldOrDefault(values, 4)).
		SetV5(policyFieldOrDefault(values, 5)).
		Save(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *policyRP) deletePolicyRow(ctx context.Context, ptype string, values []string) error {
	_, err := rp.store.GetClient(ctx).PolicyRule.Delete().
		Where(
			entpolicyrule.Ptype(ptype),
			entpolicyrule.V0(policyFieldOrDefault(values, 0)),
			entpolicyrule.V1(policyFieldOrDefault(values, 1)),
			entpolicyrule.V2(policyFieldOrDefault(values, 2)),
			entpolicyrule.V3(policyFieldOrDefault(values, 3)),
			entpolicyrule.V4(policyFieldOrDefault(values, 4)),
			entpolicyrule.V5(policyFieldOrDefault(values, 5)),
		).
		Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *policyRP) deleteFilteredPolicyRows(ctx context.Context, ptype string, fieldIndex int, fieldValues ...string) error {
	type fieldPred func(string) predicate.PolicyRule

	fieldMap := map[int]fieldPred{
		0: entpolicyrule.V0EQ,
		1: entpolicyrule.V1EQ,
		2: entpolicyrule.V2EQ,
		3: entpolicyrule.V3EQ,
		4: entpolicyrule.V4EQ,
		5: entpolicyrule.V5EQ,
	}

	var ps []predicate.PolicyRule
	for i, val := range fieldValues {
		if fn, ok := fieldMap[fieldIndex+i]; ok && val != "" {
			ps = append(ps, fn(val))
		}
	}

	_, err := rp.store.GetClient(ctx).PolicyRule.Delete().
		Where(entpolicyrule.Ptype(ptype)).
		Where(ps...).
		Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

// Logger ---------------------------------------------------------------------------------------------------------------

type policyLogger struct {
	helper  *log.Helper
	enabled bool
}

func newPolicyLogger(helper *log.Helper) *policyLogger {
	return &policyLogger{helper: helper, enabled: true}
}

func (l *policyLogger) EnableLog(b bool)      { l.enabled = b }
func (l *policyLogger) IsEnabled() bool       { return l.enabled }
func (l *policyLogger) LogModel(m [][]string) { l.helper.Debugf("casbin model: %v", m) }
func (l *policyLogger) LogEnforce(matcher string, request []any, result bool, explains [][]string) {
	l.helper.Debugf("casbin enforce: matcher=%s request=%v result=%v explains=%v", matcher, request, result, explains)
}
func (l *policyLogger) LogRole(roles []string) { l.helper.Debugf("casbin roles: %v", roles) }
func (l *policyLogger) LogPolicy(policy map[string][][]string) {
	l.helper.Debugf("casbin policy: %v", policy)
}
func (l *policyLogger) LogError(err error, msg ...string) {
	l.helper.Errorf("casbin error: %v %v", err, msg)
}

// Private ----------------------------------------------------------------------------------------------------------------------------

func toAnySlice(ss []string) []any {
	ifaces := make([]any, len(ss))
	for i, s := range ss {
		ifaces[i] = s
	}
	return ifaces
}

func policyFieldOrDefault(rule []string, idx int) string {
	if idx < len(rule) {
		return rule[idx]
	}
	return ""
}

func policyLine(ptype string, fields ...string) string {
	parts := []string{ptype}
	for _, f := range fields {
		if f == "" {
			break
		}
		parts = append(parts, f)
	}
	return strings.Join(parts, ",")
}
