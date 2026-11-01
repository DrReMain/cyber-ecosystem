package agreement

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	docUser    = "user"
	docPrivacy = "privacy"

	loginSMSCode  = "sms_code"
	loginPassword = "password"
	loginWechat   = "wechat"
)

var docTitles = map[string]string{
	docUser:    "《用户服务协议》",
	docPrivacy: "《隐私政策》",
}

var rosterHydrate = utils.Ptr(int32(helper.DefaultPageSizeMax)) // one roster page per round-trip; the export's fold loop pages until the roster answers everyone

// DO ------------------------------------------------------------------------------------------------------------------

type AgreementVersion struct {
	ID string
	// "user" / "privacy" — the doc enum folded to its column value.
	Doc         string
	No          string
	Title       string
	Body        string
	EffectiveAt time.Time
	IsCurrent   bool
	// Distinct users who ever consented to this (doc, no) — a display
	// fold the operator list fills in; zero on a bare row.
	AgreedCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AgreementConsent struct {
	// One evidence row; ConsentAt is the row's created_at.
	UserID      string
	Doc         string
	VersionNo   string
	LoginMethod string
	ConsentAt   time.Time
}

type AdminConsentRow struct {
	// The evidence row with its two display folds — nickname from the
	// local profile, masked phone from the system roster — so the
	// operator reads one table instead of joining three systems.
	AgreementConsent
	Nickname    *string
	PhoneMasked *string
}

type VersionKey struct {
	Doc string
	No  string
}

type ListConsentsIn struct {
	*commonv1.PageRequest
	Doc       *string
	VersionNo *string
	// Nickname contains — resolved to user ids through the profile plane
	// before the consent query runs.
	Keyword *string
	// Resolved keyword match; nil = unfiltered.
	UserIDs []string
}

type ListConsentsOut struct {
	*commonv1.PageResponse
	Rows []*AdminConsentRow
}

// Port ----------------------------------------------------------------------------------------------------------------

type AgreementRP interface {
	Create(ctx context.Context, v *AgreementVersion) (*AgreementVersion, error)
	// Clears every current flag of one document except keepID ("" clears
	// all) — one half of the create/promote flip, always inside a
	// transaction.
	ClearCurrent(ctx context.Context, doc, keepID string) error
	MarkCurrent(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, doc *string) ([]*AgreementVersion, error)
	Current(ctx context.Context) ([]*AgreementVersion, error)
	Find(ctx context.Context, id string) (*AgreementVersion, error)
	FindByDocNo(ctx context.Context, doc, no string) (*AgreementVersion, error)
	CountAgreedUsers(ctx context.Context) (map[VersionKey]int, error)
	ListConsents(ctx context.Context, in *ListConsentsIn) (*ListConsentsOut, error)
	// The export's unpaged twin — same filters, no page: the paginated
	// contract caps page sizes, and the "all" flag is only legal on an
	// unbounded config, which this table's list deliberately is not.
	ListAllConsents(ctx context.Context, in *ListConsentsIn) ([]*AgreementConsent, error)
	RecordConsents(ctx context.Context, rows []*AgreementConsent) error
}

// UC ------------------------------------------------------------------------------------------------------------------

type AgreementUC struct {
	shared.UC
	agreementRP AgreementRP
}

func NewAgreementUC(logger *slog.Logger, tm shared.Transaction, agreementRP AgreementRP) *AgreementUC {
	return &AgreementUC{
		UC:          shared.NewUC(logger.With("module", "module/agreement"), tm),
		agreementRP: agreementRP,
	}
}

type AgreementAdminUC struct {
	shared.UC
	agreementRP AgreementRP
	profileRP   profile.ProfileRP
	systemRP    member.SystemRP
}

func NewAgreementAdminUC(
	logger *slog.Logger, tm shared.Transaction,
	agreementRP AgreementRP, profileRP profile.ProfileRP, systemRP member.SystemRP,
) *AgreementAdminUC {
	return &AgreementAdminUC{
		UC:          shared.NewUC(logger.With("module", "module/agreement"), tm),
		agreementRP: agreementRP,
		profileRP:   profileRP,
		systemRP:    systemRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AgreementUC) Current(ctx context.Context) ([]*AgreementVersion, error) {
	return uc.agreementRP.Current(ctx)
}

func (uc *AgreementUC) Record(ctx context.Context, loginMethod string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("consent recording requires a member session"))
	}
	current, err := uc.agreementRP.Current(ctx)
	if err != nil {
		return err
	}
	if len(current) == 0 {
		return esportspb.ErrorEsportsAgreementNotConfigured("").WithCause(fmt.Errorf("no current agreement version exists to consent to"))
	}
	// One row per current document, same timestamp: the login gate
	// requires every box checked, so a consent event is the whole set —
	// the version each row carries is what the server resolves now,
	// never what a (possibly stale) client believes.
	rows := make([]*AgreementConsent, 0, len(current))
	for _, v := range current {
		rows = append(rows, &AgreementConsent{
			UserID:      subject.UserID,
			Doc:         v.Doc,
			VersionNo:   v.No,
			LoginMethod: loginMethod,
			ConsentAt:   time.Now(),
		})
	}
	return uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		return uc.agreementRP.RecordConsents(txCtx, rows)
	})
}

func (uc *AgreementAdminUC) CreateVersion(ctx context.Context, doc, no string, effectiveAt time.Time, body string) (*AgreementVersion, error) {
	_, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("version creation requires an admin session"))
	}
	// The title derives from the document type — the create face takes
	// no title input, exactly like the prototype's modal derives it.
	title, known := docTitles[doc]
	if !known {
		return nil, errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("unknown agreement doc %q", doc))
	}
	existing, err := uc.agreementRP.FindByDocNo(ctx, doc, no)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, esportspb.ErrorEsportsAgreementVersionDuplicate("").WithCause(fmt.Errorf("version %s already exists for %s", no, doc))
	}
	var created *AgreementVersion
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// Publishing is atomic: the outgoing version loses current and
		// the new one gains it in one transaction, so no consent can
		// land in a window where a document has two currents — or none.
		if err := uc.agreementRP.ClearCurrent(txCtx, doc, ""); err != nil {
			return err
		}
		created, err = uc.agreementRP.Create(txCtx, &AgreementVersion{
			Doc: doc, No: no, Title: title, Body: body,
			EffectiveAt: effectiveAt, IsCurrent: true,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (uc *AgreementAdminUC) SetCurrent(ctx context.Context, id string) (*AgreementVersion, error) {
	row, err := uc.agreementRP.Find(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, esportspb.ErrorEsportsAgreementVersionNotFound("").WithCause(err)
		}
		return nil, err
	}
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		if err := uc.agreementRP.ClearCurrent(txCtx, row.Doc, id); err != nil {
			return err
		}
		return uc.agreementRP.MarkCurrent(txCtx, id)
	})
	if err != nil {
		return nil, err
	}
	return uc.agreementRP.Find(ctx, id)
}

func (uc *AgreementAdminUC) DeleteVersion(ctx context.Context, id string) error {
	// The current-check and the delete are one transaction: a promote
	// committing between a read and a delete would remove the freshly
	// current version and leave its document with nothing to consent to.
	return uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		row, err := uc.agreementRP.Find(txCtx, id)
		if err != nil {
			if errorspb.IsInfraErrorDbNotFound(err) {
				return esportspb.ErrorEsportsAgreementVersionNotFound("").WithCause(err)
			}
			return err
		}
		if row.IsCurrent {
			return esportspb.ErrorEsportsAgreementVersionIsCurrent("").WithCause(fmt.Errorf("version %s is current for %s; switch current first", row.No, row.Doc))
		}
		return uc.agreementRP.Delete(txCtx, id)
	})
}

func (uc *AgreementAdminUC) ListVersions(ctx context.Context, doc *string) ([]*AgreementVersion, error) {
	rows, err := uc.agreementRP.List(ctx, doc)
	if err != nil {
		return nil, err
	}
	counts, err := uc.agreementRP.CountAgreedUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range rows {
		v.AgreedCount = counts[VersionKey{Doc: v.Doc, No: v.No}]
	}
	return rows, nil
}

func (uc *AgreementAdminUC) ListConsents(ctx context.Context, in *ListConsentsIn) (*ListConsentsOut, error) {
	if in.Keyword != nil {
		// Empty match = empty answer — the honest no-result page, not an
		// unfiltered list.
		ids, err := uc.profileRP.ListUserIDs(ctx, in.Keyword, nil)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return &ListConsentsOut{PageResponse: &commonv1.PageResponse{}}, nil
		}
		in.UserIDs = ids
	}
	out, err := uc.agreementRP.ListConsents(ctx, in)
	if err != nil {
		return nil, err
	}
	if len(out.Rows) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(out.Rows))
	for _, row := range out.Rows {
		ids = append(ids, row.UserID)
	}
	nickname, phone, err := uc.identityFolds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, row := range out.Rows {
		row.Nickname = nickname[row.UserID]
		row.PhoneMasked = phone[row.UserID]
	}
	return out, nil
}

func (uc *AgreementAdminUC) ExportConsents(ctx context.Context, in *ListConsentsIn) ([]*AdminConsentRow, error) {
	if in.Keyword != nil {
		ids, err := uc.profileRP.ListUserIDs(ctx, in.Keyword, nil)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, nil
		}
		in.UserIDs = ids
	}
	rows, err := uc.agreementRP.ListAllConsents(ctx, in)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	nickname, phone, err := uc.identityFolds(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]*AdminConsentRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, &AdminConsentRow{
			AgreementConsent: *row,
			Nickname:         nickname[row.UserID],
			PhoneMasked:      phone[row.UserID],
		})
	}
	return out, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AgreementAdminUC) identityFolds(ctx context.Context, userIDs []string) (map[string]*string, map[string]*string, error) {
	// The consent table's two display columns resolve from the read
	// planes the operator face already composes: nickname is the local
	// profile fold, the masked phone the system roster projection. A
	// member absent from either plane keeps null facts — the honest lazy
	// shape, never an error.
	// One login writes two rows (user + privacy), so any page or export
	// carries repeated ids; the roster's id list is unique-items on the
	// wire, so dedupe before the call.
	ids := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		if _, dup := seen[id]; !dup {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	// The export reads unpaged, so its id set can outgrow one roster
	// page: page until the roster has answered every matching member —
	// its Total ends the loop, not the requested size, because the
	// roster may clamp the page below it.
	phone := make(map[string]*string, len(ids))
	for page := int32(1); ; page++ {
		roster, err := uc.systemRP.ListRoster(ctx, &member.RosterQuery{
			Population: member.DefaultPopulation,
			UserIDs:    ids,
			Page:       &commonv1.PageRequest{PageNo: utils.Ptr(page), PageSize: rosterHydrate},
		})
		if err != nil {
			return nil, nil, err
		}
		for _, u := range roster.Users {
			phone[u.ID] = u.PhoneMasked
		}
		if len(roster.Users) == 0 || int32(len(phone)) >= roster.GetTotal() {
			break
		}
	}
	profiles, err := uc.profileRP.ListByUserIDs(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	nickname := make(map[string]*string, len(profiles))
	for _, p := range profiles {
		nickname[p.UserID] = p.Nickname
	}
	return nickname, phone, nil
}
