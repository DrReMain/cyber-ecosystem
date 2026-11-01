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
	"cyber-ecosystem/app/services/esports/internal/ent/activationbatch"
	"cyber-ecosystem/app/services/esports/internal/ent/activationcode"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/ent/reportpolicy"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

type batchAgg struct {
	BatchNo string `json:"no"`
	Count   int    `json:"count"`
}

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

func (rp *codeRP) List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error) {
	query, err := rp.applyCodeFilters(ctx, in)
	if err != nil {
		return nil, err
	}
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(activationcode.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	codes := mapActivationCodes(rows)
	if err := rp.hydrateBatchNos(ctx, codes); err != nil {
		return nil, err
	}
	return &ListCodesOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Codes:        codes,
	}, nil
}

func (rp *codeRP) ListAll(ctx context.Context, in *ListCodesIn) ([]*ActivationCode, error) {
	query, err := rp.applyCodeFilters(ctx, in)
	if err != nil {
		return nil, err
	}
	rows, err := query.Order(ent.Desc(activationcode.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	codes := mapActivationCodes(rows)
	if err := rp.hydrateBatchNos(ctx, codes); err != nil {
		return nil, err
	}
	return codes, nil
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
	out := mapActivationCode(row)
	if row.BatchID != nil {
		batch, err := rp.Platform.GetClient(ctx).ActivationBatch.Query().
			Where(activationbatch.IDEQ(*row.BatchID)).
			Only(ctx)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
		no := batch.No
		out.BatchNo = &no
		out.BatchDisabled = batch.DisabledAt != nil
	}
	return out, nil
}

func (rp *codeRP) TakeUse(ctx context.Context, code, redeemedBy string) (bool, error) {
	// The use and its trail land in one statement — who took it and when
	// are facts of the code row itself, because the family grant the
	// redeem produces can later be ratcheted by a newer code and lose the
	// source trail this column preserves.
	rows, err := rp.Platform.GetClient(ctx).ActivationCode.Update().
		Where(
			activationcode.CodeEQ(code),
			func(s *sql.Selector) {
				s.Where(sql.ColumnsLT(s.C(activationcode.FieldUsedCount), s.C(activationcode.FieldQuantity)))
				// The batch must still be live — a disabled batch voids its
				// unredeemed codes on the spot, and legacy rows without a
				// batch predate the model and stay spendable.
				live := sql.Select(activationbatch.FieldID).
					From(sql.Table(activationbatch.Table)).
					Where(sql.IsNull(activationbatch.FieldDisabledAt))
				s.Where(sql.Or(
					sql.IsNull(s.C(activationcode.FieldBatchID)),
					sql.In(s.C(activationcode.FieldBatchID), live),
				))
			},
		).
		AddUsedCount(1).
		SetRedeemedBy(redeemedBy).
		SetRedeemedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *codeRP) MintBatch(ctx context.Context, batch *ActivationBatch, rows []*ActivationCode) error {
	client := rp.Platform.GetClient(ctx)
	batchRow, err := client.ActivationBatch.Create().
		SetNo(batch.BatchNo).
		SetReportType(activationbatch.ReportType(batch.ReportType)).
		SetExpiresAt(batch.ExpiresAt).
		SetCreatedBy(batch.CreatedBy).
		Save(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	builders := make([]*ent.ActivationCodeCreate, 0, len(rows))
	for _, r := range rows {
		builders = append(builders, client.ActivationCode.Create().
			SetCode(r.Code).
			SetReportType(activationcode.ReportType(r.ReportType)).
			SetQuantity(r.Quantity).
			SetUsedCount(r.UsedCount).
			SetExpiresAt(r.ExpiresAt).
			SetCreatedBy(r.CreatedBy).
			SetBatchID(batchRow.ID))
	}
	if _, err := client.ActivationCode.CreateBulk(builders...).Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *codeRP) SetBatchDisabled(ctx context.Context, batchNo string, disabled bool) (*ActivationBatch, error) {
	client := rp.Platform.GetClient(ctx)
	update := client.ActivationBatch.Update().Where(activationbatch.NoEQ(batchNo))
	if disabled {
		// Only a live batch flips off — a repeated disable is a no-op, not
		// a moved timestamp.
		update = update.Where(activationbatch.DisabledAtIsNil()).SetDisabledAt(time.Now())
	} else {
		update = update.Where(activationbatch.DisabledAtNotNil()).ClearDisabledAt()
	}
	if _, err := update.Save(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	row, err := client.ActivationBatch.Query().Where(activationbatch.NoEQ(batchNo)).Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	batch := mapActivationBatch(row)
	if err := rp.foldBatchAggregates(ctx, []*ActivationBatch{batch}); err != nil {
		return nil, err
	}
	return batch, nil
}

func (rp *codeRP) ListBatches(ctx context.Context, in *ListBatchesIn) (*ListBatchesOut, error) {
	query := rp.Platform.GetClient(ctx).ActivationBatch.Query()
	if in.ReportType != nil {
		query = query.Where(activationbatch.ReportTypeEQ(activationbatch.ReportType(*in.ReportType)))
	}
	if in.Disabled != nil {
		if *in.Disabled {
			query = query.Where(activationbatch.DisabledAtNotNil())
		} else {
			query = query.Where(activationbatch.DisabledAtIsNil())
		}
	}
	if in.BatchNo != nil {
		query = query.Where(activationbatch.NoEQ(*in.BatchNo))
	}
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(activationbatch.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := &ListBatchesOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Batches:      mapActivationBatches(rows),
	}
	if err := rp.foldBatchAggregates(ctx, out.Batches); err != nil {
		return nil, err
	}
	// The footer's fold ignores every filter — one answer over the whole
	// corpus instead of one query per number.
	if out.TotalBatches, err = rp.Platform.GetClient(ctx).ActivationBatch.Query().Count(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if out.TotalCodes, err = rp.Platform.GetClient(ctx).ActivationCode.Query().Count(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if out.UsedCodes, err = rp.Platform.GetClient(ctx).ActivationCode.Query().Where(activationcode.UsedCountGT(0)).Count(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return out, nil
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
			entitlement.StatusEQ(entitlement.StatusActive),
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
			entitlement.StatusEQ(entitlement.StatusActive),
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

func (rp *entitlementRP) ListMembershipRows(ctx context.Context, userIDs []string) ([]*EntitlementRow, error) {
	query := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(entitlement.ScopeKindEQ(entitlement.ScopeKind(shared.ScopeMembership)))
	if userIDs != nil {
		if len(userIDs) == 0 {
			return nil, nil
		}
		query = query.Where(entitlement.UserIDIn(userIDs...))
	}
	// Oldest first: the fold walks a member's rows in grant order, so the
	// last row is the live one and the count of rows is the period count.
	rows, err := query.Order(ent.Asc(entitlement.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make([]*EntitlementRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapEntitlementRow(row))
	}
	return out, nil
}

func (rp *entitlementRP) TerminateLiveMembership(ctx context.Context, userID string) (bool, error) {
	// The TakeUse shape: one conditional UPDATE is the whole arbitration —
	// every still-active, still-open row flips, and a replay matches
	// nothing. All of them, not just the newest: renewal tiles windows, so
	// a member mid-renewal holds several unexpired rows, and leaving the
	// older ones active would keep access alive behind the terminate
	// action's back. Save's row count carries the verdict; zero rows is the
	// no-live-period answer the caller reports, not an error to surface.
	rows, err := rp.Platform.GetClient(ctx).Entitlement.Update().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(shared.ScopeMembership)),
			entitlement.StatusEQ(entitlement.StatusActive),
			entitlement.ExpiresAtGT(time.Now()),
		).
		SetStatus(entitlement.StatusTerminated).
		SetTerminatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *entitlementRP) ListMembershipUnlocks(ctx context.Context, userIDs []string) ([]*WindowUnlockRow, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	rows, err := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDIn(userIDs...),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(shared.ScopeTrial)),
			entitlement.SourceEQ(shared.SourceMembership),
		).
		Select(entitlement.FieldUserID, entitlement.FieldCreatedAt).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make([]*WindowUnlockRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, &WindowUnlockRow{UserID: row.UserID, GrantedAt: row.CreatedAt})
	}
	return out, nil
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

func (rp *codeRP) foldBatchAggregates(ctx context.Context, batches []*ActivationBatch) error {
	// Fills the count/used pair on each batch row with two grouped
	// queries — the batch row carries no counts of its own, so the numbers
	// cannot drift from the code rows they describe.
	ids := make([]any, 0, len(batches))
	byID := make(map[string]*ActivationBatch, len(batches))
	for _, b := range batches {
		ids = append(ids, b.BatchNo)
		byID[b.BatchNo] = b
	}
	if len(ids) == 0 {
		return nil
	}
	counts, err := rp.groupedCount(ctx, false, ids)
	if err != nil {
		return err
	}
	used, err := rp.groupedCount(ctx, true, ids)
	if err != nil {
		return err
	}
	for _, agg := range counts {
		if b := byID[agg.BatchNo]; b != nil {
			b.Count = agg.Count
		}
	}
	for _, agg := range used {
		if b := byID[agg.BatchNo]; b != nil {
			b.UsedCount = agg.Count
		}
	}
	return nil
}

func (rp *codeRP) groupedCount(ctx context.Context, redeemedOnly bool, batchNos []any) ([]batchAgg, error) {
	// The aggregate joins the code rows back to their batch by batch_id;
	// the returned key is the human batch number the caller folds by.
	query := rp.Platform.GetClient(ctx).ActivationCode.Query()
	if redeemedOnly {
		query = query.Where(activationcode.UsedCountGT(0))
	}
	var out []batchAgg
	err := query.Modify(func(s *sql.Selector) {
		batch := sql.Table(activationbatch.Table)
		s.Join(batch).On(s.C(activationcode.FieldBatchID), batch.C(activationbatch.FieldID))
		if redeemedOnly {
			s.Where(sql.NotNull(s.C(activationcode.FieldBatchID)))
		}
		s.Where(sql.In(batch.C(activationbatch.FieldNo), batchNos...))
		s.Select(batch.C(activationbatch.FieldNo), sql.Count(s.C(activationcode.FieldID)))
		s.GroupBy(batch.C(activationbatch.FieldNo))
	}).Scan(ctx, &out)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return out, nil
}

func (rp *codeRP) applyCodeFilters(ctx context.Context, in *ListCodesIn) (*ent.ActivationCodeQuery, error) {
	// The one home of the list/export predicate set so the paged read and
	// the export can never disagree on what matches.
	query := rp.Platform.GetClient(ctx).ActivationCode.Query()
	if in.ReportType != nil {
		query = query.Where(activationcode.ReportTypeEQ(activationcode.ReportType(*in.ReportType)))
	}
	if in.BatchNo != nil {
		batch, err := rp.Platform.GetClient(ctx).ActivationBatch.Query().
			Where(activationbatch.NoEQ(*in.BatchNo)).
			Only(ctx)
		switch {
		case err == nil:
			query = query.Where(activationcode.BatchIDEQ(batch.ID))
		case ent.IsNotFound(err):
			// A batch number that matches nothing matches no code either —
			// an empty answer, not an error, the lazy-shape ruling.
			query = query.Where(activationcode.IDIn())
		default:
			return nil, rp.Platform.HandleEntError(err)
		}
	}
	if in.Status != nil {
		switch *in.Status {
		case CodeStatusUnused:
			query = query.Where(activationcode.UsedCountEQ(0))
		case CodeStatusRedeemed:
			query = query.Where(activationcode.UsedCountGT(0))
		}
	}
	if in.Code != nil {
		// Exact match — the search box chases one known code, batch_no's
		// own lookup shape.
		query = query.Where(activationcode.CodeEQ(*in.Code))
	}
	if in.User != nil {
		// The member's redeemed view — the user detail reads its
		// report-unlock facts from these rows.
		query = query.Where(activationcode.RedeemedByEQ(*in.User))
	}
	return query, nil
}

func (rp *codeRP) hydrateBatchNos(ctx context.Context, codes []*ActivationCode) error {
	// Fills each code's batch number from the batch table — the rows
	// carry the id link, the faces read the human shape.
	ids := make([]string, 0, len(codes))
	for _, c := range codes {
		if c.BatchID != nil {
			ids = append(ids, *c.BatchID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	batches, err := rp.Platform.GetClient(ctx).ActivationBatch.Query().
		Where(activationbatch.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	byID := make(map[string]string, len(batches))
	for _, b := range batches {
		byID[b.ID] = b.No
	}
	for _, c := range codes {
		if c.BatchID != nil {
			if no, ok := byID[*c.BatchID]; ok {
				c.BatchNo = &no
			}
		}
	}
	return nil
}

func mapActivationCode(row *ent.ActivationCode) *ActivationCode {
	c := &ActivationCode{
		Code:       row.Code,
		ReportType: string(row.ReportType),
		Quantity:   row.Quantity,
		UsedCount:  row.UsedCount,
		ExpiresAt:  row.ExpiresAt,
		CreatedBy:  row.CreatedBy,
		CreatedAt:  row.CreatedAt,
		BatchID:    row.BatchID,
		RedeemedBy: row.RedeemedBy,
		RedeemedAt: row.RedeemedAt,
	}
	return c
}

func mapActivationBatch(row *ent.ActivationBatch) *ActivationBatch {
	return &ActivationBatch{
		BatchNo:    row.No,
		ReportType: string(row.ReportType),
		ExpiresAt:  row.ExpiresAt,
		DisabledAt: row.DisabledAt,
		CreatedAt:  row.CreatedAt,
	}
}

func mapActivationBatches(rows []*ent.ActivationBatch) []*ActivationBatch {
	out := make([]*ActivationBatch, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapActivationBatch(r))
	}
	return out
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
		UserID:       row.UserID,
		ScopeKind:    string(row.ScopeKind),
		ScopeRef:     row.ScopeRef,
		ExpiresAt:    row.ExpiresAt,
		Source:       row.Source,
		CreatedAt:    row.CreatedAt,
		Status:       string(row.Status),
		TerminatedAt: row.TerminatedAt,
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
