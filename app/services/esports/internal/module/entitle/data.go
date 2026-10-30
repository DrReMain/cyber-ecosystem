package entitle

import (
	"context"
	dbsql "database/sql"
	"errors"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"

	"cyber-ecosystem/shared-go/helper"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/activationcode"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/ent/reportpolicy"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type codeRP struct {
	shared.RP
}

func NewActivationCodeRP(logger *slog.Logger, p *platform.Platform) ActivationCodeRP {
	return &codeRP{RP: shared.NewRP(logger.With("module", "module/entitle_rp"), p)}
}

type entitlementRP struct {
	shared.RP
}

func NewEntitlementRP(logger *slog.Logger, p *platform.Platform) EntitlementRP {
	return &entitlementRP{RP: shared.NewRP(logger.With("module", "module/entitle_rp"), p)}
}

type policyRP struct {
	shared.RP
}

func NewReportPolicyRP(logger *slog.Logger, p *platform.Platform) ReportPolicyRP {
	return &policyRP{RP: shared.NewRP(logger.With("module", "module/entitle_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *codeRP) CreateBatch(ctx context.Context, rows []*ActivationCode) error {
	builders := make([]*ent.ActivationCodeCreate, 0, len(rows))
	create := rp.Platform.GetClient(ctx).ActivationCode
	for _, r := range rows {
		builders = append(builders, create.Create().
			SetCode(r.Code).
			SetReportType(activationcode.ReportType(r.ReportType)).
			SetQuantity(r.Quantity).
			SetUsedCount(r.UsedCount).
			SetExpiresAt(r.ExpiresAt).
			SetCreatedBy(r.CreatedBy))
	}
	if _, err := create.CreateBulk(builders...).Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *codeRP) List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error) {
	query := rp.Platform.GetClient(ctx).ActivationCode.Query()
	if in.ReportType != nil {
		query = query.Where(activationcode.ReportTypeEQ(activationcode.ReportType(*in.ReportType)))
	}
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(activationcode.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListCodesOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Codes:        mapActivationCodes(rows),
	}, nil
}

func (rp *codeRP) ByCode(ctx context.Context, code string) (*ActivationCode, error) {
	row, err := rp.Platform.GetClient(ctx).ActivationCode.Query().
		Where(activationcode.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapActivationCode(row), nil
}

func (rp *codeRP) TakeUse(ctx context.Context, code string) (bool, error) {
	rows, err := rp.Platform.GetClient(ctx).ActivationCode.Update().
		Where(
			activationcode.CodeEQ(code),
			func(s *sql.Selector) {
				s.Where(sql.ColumnsLT(s.C(activationcode.FieldUsedCount), s.C(activationcode.FieldQuantity)))
			},
		).
		AddUsedCount(1).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *entitlementRP) Grant(ctx context.Context, g *EntitlementGrant) error {
	create := rp.Platform.GetClient(ctx).Entitlement.Create().
		SetUserID(g.UserID).
		SetScopeKind(entitlement.ScopeKind(g.ScopeKind)).
		SetSource(g.Source)
	if g.ScopeRef != nil {
		create = create.SetScopeRef(*g.ScopeRef)
	}
	if g.ExpiresAt != nil {
		create = create.SetExpiresAt(*g.ExpiresAt)
	}
	// ON CONFLICT DO NOTHING keeps the caller's transaction unpoisoned: a
	// rival grant for the same scope is a no-op statement, never an error
	// the transaction would have to survive. ent surfaces the swallowed
	// insert as ErrNoRows — the RETURNING clause finds no row — and that
	// is the no-op answer itself, not a failure to report.
	if err := create.OnConflict().DoNothing().Exec(ctx); err != nil && !errors.Is(err, dbsql.ErrNoRows) {
		return rp.Platform.HandleEntError(err)
	}
	if g.ScopeKind != shared.ScopeReportType || g.ExpiresAt == nil {
		return nil
	}
	return rp.ratchetGrant(ctx, g)
}

func (rp *entitlementRP) ByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*EntitlementRow, error) {
	query := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(scopeKind)),
		)
	if scopeRef != nil {
		query = query.Where(entitlement.ScopeRefEQ(*scopeRef))
	} else {
		query = query.Where(entitlement.ScopeRefIsNil())
	}
	row, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapEntitlementRow(row), nil
}

func (rp *entitlementRP) HasActiveMembership(ctx context.Context, userID string) (bool, error) {
	ok, err := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(shared.ScopeMembership)),
			entitlement.ExpiresAtGT(time.Now()),
		).
		Exist(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return ok, nil
}

func (rp *entitlementRP) MembershipWindow(ctx context.Context, userID string) (*EntitlementRow, error) {
	// The display read for the merge offer: whichever live membership window
	// runs furthest. Multiple tiers ratchet into separate rows; the confirm
	// page promises the longest standing benefit, and a lapsed-only history
	// reads as no membership at all.
	rows, err := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(shared.ScopeMembership)),
			entitlement.ExpiresAtGT(time.Now()),
		).
		Order(ent.Desc(entitlement.FieldExpiresAt)).
		Limit(1).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapEntitlementRow(rows[0]), nil
}

func (rp *entitlementRP) HasAnyScope(ctx context.Context, userID, scopeKind string) (bool, error) {
	ok, err := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(scopeKind)),
		).
		Exist(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return ok, nil
}

func (rp *entitlementRP) Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error) {
	row, err := rp.ByScope(ctx, userID, scopeKind, scopeRef)
	if err != nil {
		return false, err
	}
	return row != nil, nil
}

func (rp *policyRP) Policy(ctx context.Context, family string) (string, bool, error) {
	row, err := rp.Platform.GetClient(ctx).ReportPolicy.Query().
		Where(reportpolicy.ReportTypeEQ(reportpolicy.ReportType(family))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", false, nil
		}
		return "", false, rp.Platform.HandleEntError(err)
	}
	return string(row.Policy), true, nil
}

func (rp *policyRP) List(ctx context.Context) ([]*PolicyRow, error) {
	rows, err := rp.Platform.GetClient(ctx).ReportPolicy.Query().All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make([]*PolicyRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, &PolicyRow{
			Family:     string(row.ReportType),
			Policy:     string(row.Policy),
			Overridden: true,
			UpdatedBy:  row.UpdatedBy,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return out, nil
}

func (rp *policyRP) SetPolicy(ctx context.Context, family, policy, updatedBy string) error {
	// Upsert in place: the family row is the policy's whole state, so the
	// conflict target is the unique family and the update rewrites it.
	if err := rp.Platform.GetClient(ctx).ReportPolicy.Create().
		SetReportType(reportpolicy.ReportType(family)).
		SetPolicy(reportpolicy.Policy(policy)).
		SetUpdatedBy(updatedBy).
		OnConflict(
			sql.ConflictColumns(reportpolicy.FieldReportType),
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapActivationCode(row *ent.ActivationCode) *ActivationCode {
	return &ActivationCode{
		Code:       row.Code,
		ReportType: string(row.ReportType),
		Quantity:   row.Quantity,
		UsedCount:  row.UsedCount,
		ExpiresAt:  row.ExpiresAt,
		CreatedBy:  row.CreatedBy,
		CreatedAt:  row.CreatedAt,
	}
}

func mapActivationCodes(rows []*ent.ActivationCode) []*ActivationCode {
	out := make([]*ActivationCode, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapActivationCode(r))
	}
	return out
}

func mapEntitlementRow(row *ent.Entitlement) *EntitlementRow {
	return &EntitlementRow{
		ScopeKind: string(row.ScopeKind),
		ScopeRef:  row.ScopeRef,
		ExpiresAt: row.ExpiresAt,
		Source:    row.Source,
		CreatedAt: row.CreatedAt,
	}
}

func (rp *entitlementRP) ratchetGrant(ctx context.Context, g *EntitlementGrant) error {
	// A family's grant is one row, so the insert above may have no-op'd
	// against a lapsed or shorter one — the ratchet extends forward only,
	// the TakeUse shape: one arbiter, the loser a no-op, and a replay of
	// the same grant matches nothing (equal is not less). The columns are
	// immutable on purpose — ledger rows never rewrite history — so the
	// forward extension rides raw SQL, the one sanctioned way to move an
	// immutable column at all; Exec (not Save) because zero rows here is
	// the no-op outcome, not a miss to surface.
	if err := rp.Platform.GetClient(ctx).Entitlement.Update().
		Modify(func(u *sql.UpdateBuilder) {
			u.Set(entitlement.FieldExpiresAt, *g.ExpiresAt).
				Set(entitlement.FieldSource, g.Source).
				Where(sql.And(
					sql.EQ(entitlement.FieldUserID, g.UserID),
					sql.EQ(entitlement.FieldScopeKind, g.ScopeKind),
					sql.EQ(entitlement.FieldScopeRef, *g.ScopeRef),
					sql.LT(entitlement.FieldExpiresAt, *g.ExpiresAt),
				))
		}).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}
