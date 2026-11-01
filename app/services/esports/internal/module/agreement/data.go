package agreement

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"
	"entgo.io/ent/dialect/sql"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/agreementconsent"
	"cyber-ecosystem/app/services/esports/internal/ent/agreementversion"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type agreementRP struct {
	shared.RP
}

func NewAgreementRP(logger *slog.Logger, p *platform.Platform) AgreementRP {
	return &agreementRP{RP: shared.NewRP(logger.With("module", "module/agreement_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *agreementRP) Create(ctx context.Context, v *AgreementVersion) (*AgreementVersion, error) {
	row, err := rp.Platform.GetClient(ctx).AgreementVersion.Create().
		SetDoc(agreementversion.Doc(v.Doc)).
		SetNo(v.No).
		SetTitle(v.Title).
		SetBody(v.Body).
		SetEffectiveAt(v.EffectiveAt).
		SetIsCurrent(v.IsCurrent).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapVersion(row), nil
}

func (rp *agreementRP) ClearCurrent(ctx context.Context, doc, keepID string) error {
	// Exec, not Save: zero matching rows is the no-op outcome (the
	// document may hold no current flag at all), not a miss to surface.
	query := rp.Platform.GetClient(ctx).AgreementVersion.Update().
		Where(agreementversion.DocEQ(agreementversion.Doc(doc)), agreementversion.IsCurrentEQ(true))
	if keepID != "" {
		query = query.Where(agreementversion.IDNEQ(keepID))
	}
	if err := query.SetIsCurrent(false).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *agreementRP) MarkCurrent(ctx context.Context, id string) error {
	if _, err := rp.Platform.GetClient(ctx).AgreementVersion.UpdateOneID(id).
		SetIsCurrent(true).
		Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *agreementRP) Delete(ctx context.Context, id string) error {
	if err := rp.Platform.GetClient(ctx).AgreementVersion.DeleteOneID(id).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *agreementRP) List(ctx context.Context, doc *string) ([]*AgreementVersion, error) {
	query := rp.Platform.GetClient(ctx).AgreementVersion.Query()
	if doc != nil {
		query = query.Where(agreementversion.DocEQ(agreementversion.Doc(*doc)))
	}
	// Current first per document, then newest-effective — the table's
	// reading order in one clause.
	rows, err := query.
		Order(ent.Asc(agreementversion.FieldDoc), ent.Desc(agreementversion.FieldIsCurrent), ent.Desc(agreementversion.FieldEffectiveAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapVersion), nil
}

func (rp *agreementRP) Current(ctx context.Context) ([]*AgreementVersion, error) {
	rows, err := rp.Platform.GetClient(ctx).AgreementVersion.Query().
		Where(agreementversion.IsCurrentEQ(true)).
		Order(ent.Asc(agreementversion.FieldDoc)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapVersion), nil
}

func (rp *agreementRP) Find(ctx context.Context, id string) (*AgreementVersion, error) {
	row, err := rp.Platform.GetClient(ctx).AgreementVersion.Query().
		Where(agreementversion.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapVersion(row), nil
}

func (rp *agreementRP) FindByDocNo(ctx context.Context, doc, no string) (*AgreementVersion, error) {
	// nil row = absent; the duplicate pre-check reads absence, not an
	// error.
	row, err := rp.Platform.GetClient(ctx).AgreementVersion.Query().
		Where(agreementversion.DocEQ(agreementversion.Doc(doc)), agreementversion.NoEQ(no)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapVersion(row), nil
}

func (rp *agreementRP) CountAgreedUsers(ctx context.Context) (map[VersionKey]int, error) {
	// COUNT(DISTINCT user_id) per (doc, version_no): a member re-logging
	// in on the same version adds rows, not agreeing users.
	var rows []struct {
		Doc    string `json:"doc"`
		No     string `json:"version_no"`
		Agreed int    `json:"count"`
	}
	err := rp.Platform.GetClient(ctx).AgreementConsent.Query().
		Modify(func(s *sql.Selector) {
			s.Select(
				s.C(agreementconsent.FieldDoc),
				s.C(agreementconsent.FieldVersionNo),
				sql.Count(sql.Distinct(s.C(agreementconsent.FieldUserID))),
			).GroupBy(
				s.C(agreementconsent.FieldDoc),
				s.C(agreementconsent.FieldVersionNo),
			)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make(map[VersionKey]int, len(rows))
	for _, row := range rows {
		out[VersionKey{Doc: row.Doc, No: row.No}] = row.Agreed
	}
	return out, nil
}

func (rp *agreementRP) ListConsents(ctx context.Context, in *ListConsentsIn) (*ListConsentsOut, error) {
	query := applyConsentFilters(rp.Platform.GetClient(ctx).AgreementConsent.Query(), in)
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(agreementconsent.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := &ListConsentsOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Rows:         make([]*AdminConsentRow, 0, len(rows)),
	}
	for _, row := range rows {
		out.Rows = append(out.Rows, &AdminConsentRow{AgreementConsent: *mapConsent(row)})
	}
	return out, nil
}

func (rp *agreementRP) ListAllConsents(ctx context.Context, in *ListConsentsIn) ([]*AgreementConsent, error) {
	query := applyConsentFilters(rp.Platform.GetClient(ctx).AgreementConsent.Query(), in)
	rows, err := query.Order(ent.Desc(agreementconsent.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapConsent), nil
}

func (rp *agreementRP) RecordConsents(ctx context.Context, rows []*AgreementConsent) error {
	if len(rows) == 0 {
		return nil
	}
	bulk := make([]*ent.AgreementConsentCreate, 0, len(rows))
	client := rp.Platform.GetClient(ctx)
	for _, r := range rows {
		bulk = append(bulk, client.AgreementConsent.Create().
			SetUserID(r.UserID).
			SetDoc(agreementconsent.Doc(r.Doc)).
			SetVersionNo(r.VersionNo).
			SetLoginMethod(agreementconsent.LoginMethod(r.LoginMethod)))
	}
	if _, err := client.AgreementConsent.CreateBulk(bulk...).Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func applyConsentFilters(query *ent.AgreementConsentQuery, in *ListConsentsIn) *ent.AgreementConsentQuery {
	if in.Doc != nil {
		query = query.Where(agreementconsent.DocEQ(agreementconsent.Doc(*in.Doc)))
	}
	if in.VersionNo != nil {
		query = query.Where(agreementconsent.VersionNoEQ(*in.VersionNo))
	}
	if in.UserIDs != nil {
		query = query.Where(agreementconsent.UserIDIn(in.UserIDs...))
	}
	if in.PageRequest != nil {
		// The time range rides the embedded page request; a caller may
		// carry none at all — reading the promoted fields through nil
		// would panic, so the whole range guards on the page existing.
		helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), agreementconsent.CreatedAtGTE)
		helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), agreementconsent.CreatedAtLTE)
	}
	return query
}

func mapVersion(row *ent.AgreementVersion) *AgreementVersion {
	return &AgreementVersion{
		ID:          row.ID,
		Doc:         string(row.Doc),
		No:          row.No,
		Title:       row.Title,
		Body:        row.Body,
		EffectiveAt: row.EffectiveAt,
		IsCurrent:   row.IsCurrent,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func mapConsent(row *ent.AgreementConsent) *AgreementConsent {
	return &AgreementConsent{
		UserID:      row.UserID,
		Doc:         string(row.Doc),
		VersionNo:   row.VersionNo,
		LoginMethod: string(row.LoginMethod),
		ConsentAt:   row.CreatedAt,
	}
}
