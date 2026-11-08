package e2e

// The agreement domain's live probe: the seed pair answering current, a
// member's consent landing one row per document on the versions the
// server resolves (never a client-declared number), the operator's
// publish/promote/delete moves with their refusals (duplicate, current
// delete, unknown id), and the evidence surviving the version's
// deletion. Gated: ENT_E2E=1 with the dev database up; anything else
// skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/agreementconsent"
	"cyber-ecosystem/app/services/esports/internal/ent/agreementversion"
	"cyber-ecosystem/app/services/esports/internal/module/agreement"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

const (
	consentUser  = "agre2euser0000000001"
	e2eVersionNo = "v9.9.9-e2e"
	seededNo     = "v1.0.0"
)

// The paginated contract caps page sizes and the "all" flag rides an
// unbounded config this table's list does not use — the probe reads the
// plain first page (the probe's own rows never reach the cap) and the
// unpaged read through the export face below.
var pageFirst = commonv1.PageRequest{}

// The profile stub answers every keyword with the one member this probe
// consents as — the operator list's keyword arm then reads exactly that
// member's rows, and the hydration carries the nickname back.
type consentProfileStub struct{}

func (consentProfileStub) FindByUserID(context.Context, string) (*profile.Profile, error) {
	return nil, nil
}
func (consentProfileStub) ListByUserIDs(_ context.Context, ids []string) ([]*profile.Profile, error) {
	out := make([]*profile.Profile, 0, len(ids))
	for _, id := range ids {
		nickname := "e2e-" + id
		out = append(out, &profile.Profile{UserID: id, Nickname: &nickname})
	}
	return out, nil
}
func (consentProfileStub) ListUserIDs(context.Context, *string, *string) ([]string, error) {
	return []string{consentUser}, nil
}
func (consentProfileStub) Save(context.Context, *profile.Profile) (*profile.Profile, error) {
	return nil, nil
}

func TestAgreementLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the agreement lifecycle end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: entEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: entEnvOr("DB_PASS", "postgres"), DbName: entEnvOr("DB_NAME", "esports"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	// Registered the moment the closer exists: a wiring failure below
	// must not leak the open client for the rest of the binary's run.
	t.Cleanup(func() {
		cleanAgreementE2E(t, db)
		closeDB()
	})
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanAgreementE2E(t, db)

	rp := agreement.NewAgreementRP(logger, pf)
	memberUC := agreement.NewAgreementUC(logger, pf, rp)
	adminUC := agreement.NewAgreementAdminUC(logger, pf, rp, consentProfileStub{}, stubRoster{})

	// The seed pair answers current before anything moves.
	current, err := memberUC.Current(context.Background())
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	seededByDoc := map[string]*agreement.AgreementVersion{}
	for _, v := range current {
		seededByDoc[v.Doc] = v
	}
	if len(seededByDoc) != 2 || seededByDoc["user"] == nil || seededByDoc["privacy"] == nil {
		t.Fatalf("current: docs %v, want the seeded user+privacy pair", seededByDoc)
	}

	// No session, no consent — the record is evidence, it never lands
	// anonymous.
	if err := memberUC.Record(context.Background(), "sms_code"); err == nil || !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("anonymous record: err=%v, want unauthenticated", err)
	}

	// A login writes one row per document on the current versions — the
	// version each row carries is what the server resolves now.
	if err := memberUC.Record(entSubject(consentUser), "sms_code"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := memberUC.Record(entSubject(consentUser), "password"); err != nil {
		t.Fatalf("record again: %v", err)
	}
	mine, err := adminUC.ListConsents(entSubject(consentUser), &agreement.ListConsentsIn{
		PageRequest: &pageFirst,
		Keyword:     utils.Ptr("nickname-anything"),
	})
	if err != nil {
		t.Fatalf("list consents: %v", err)
	}
	if len(mine.Rows) != 4 {
		t.Fatalf("consent rows: %d, want 4 (two logins x two docs)", len(mine.Rows))
	}
	methods := map[string]int{}
	for _, row := range mine.Rows {
		if row.VersionNo != seededByDoc[row.Doc].No {
			t.Fatalf("consent landed on %s/%s, want the seeded current", row.Doc, row.VersionNo)
		}
		methods[row.LoginMethod]++
	}
	if methods["sms_code"] != 2 || methods["password"] != 2 {
		t.Fatalf("login methods: %v, want one row per doc per login", methods)
	}

	// A caller that carries no page at all gets the pagination refusal —
	// never a panic on the timestamp reads promoted through the nil
	// embedded page request.
	if _, err := adminUC.ListConsents(entSubject(consentUser), &agreement.ListConsentsIn{
		Keyword: utils.Ptr("nickname-anything"),
	}); err == nil || !errorspb.IsGeneralErrorPaginationInvalidArgument(err) {
		t.Fatalf("nil-page list: err=%v, want pagination invalid argument", err)
	}

	// Publishing: create flips current atomically; the demoted seeded
	// version keeps the count it earned (distinct users, so two logins by
	// one member still count once).
	versions, err := adminUC.ListVersions(entSubject(consentUser), nil)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	var seededUserCount int
	for _, v := range versions {
		if v.Doc == "user" && v.No == seededByDoc["user"].No {
			seededUserCount = v.AgreedCount
		}
	}
	if seededUserCount < 1 {
		t.Fatalf("seeded user version agreed count: %d, want >=1", seededUserCount)
	}
	published, err := adminUC.CreateVersion(entSubject(consentUser), "user", e2eVersionNo, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), "e2e body")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if !published.IsCurrent || published.Title != "《用户服务协议》" {
		t.Fatalf("created version fold: %+v", published)
	}
	if _, err := adminUC.CreateVersion(entSubject(consentUser), "user", e2eVersionNo, time.Now(), "dup"); err == nil || !esportspb.IsEsportsAgreementVersionDuplicate(err) {
		t.Fatalf("duplicate create: err=%v, want duplicate", err)
	}

	// The next consent lands on the published version only — privacy's
	// current is untouched by a user-doc publish.
	if err := memberUC.Record(entSubject(consentUser), "sms_code"); err != nil {
		t.Fatalf("record after publish: %v", err)
	}
	mine, err = adminUC.ListConsents(entSubject(consentUser), &agreement.ListConsentsIn{
		PageRequest: &pageFirst,
		VersionNo:   utils.Ptr(e2eVersionNo),
	})
	if err != nil {
		t.Fatalf("list consents by version: %v", err)
	}
	if len(mine.Rows) != 1 || mine.Rows[0].Doc != "user" {
		t.Fatalf("post-publish rows: %+v, want one user row on the new version", mine.Rows)
	}

	// Refusals: the current version cannot be deleted; unknown ids answer
	// not-found.
	if err := adminUC.DeleteVersion(entSubject(consentUser), published.ID); err == nil || !esportspb.IsEsportsAgreementVersionIsCurrent(err) {
		t.Fatalf("delete current: err=%v, want is-current", err)
	}
	if err := adminUC.DeleteVersion(entSubject(consentUser), "nosuchversion00001"); err == nil || !esportspb.IsEsportsAgreementVersionNotFound(err) {
		t.Fatalf("delete unknown: err=%v, want not found", err)
	}
	if _, err := adminUC.SetCurrent(entSubject(consentUser), "nosuchversion00001"); err == nil || !esportspb.IsEsportsAgreementVersionNotFound(err) {
		t.Fatalf("promote unknown: err=%v, want not found", err)
	}

	// Promote the seeded version back, delete the published one — and the
	// evidence rows it earned outlive it, queryable by their string.
	promoted, err := adminUC.SetCurrent(entSubject(consentUser), seededByDoc["user"].ID)
	if err != nil {
		t.Fatalf("promote back: %v", err)
	}
	if !promoted.IsCurrent {
		t.Fatalf("promote back: seeded version not current")
	}
	if err := adminUC.DeleteVersion(entSubject(consentUser), published.ID); err != nil {
		t.Fatalf("delete published: %v", err)
	}
	mine, err = adminUC.ListConsents(entSubject(consentUser), &agreement.ListConsentsIn{
		PageRequest: &pageFirst,
		VersionNo:   utils.Ptr(e2eVersionNo),
	})
	if err != nil {
		t.Fatalf("list consents of deleted version: %v", err)
	}
	if len(mine.Rows) != 1 {
		t.Fatalf("evidence after version delete: %d rows, want the earned row to survive", len(mine.Rows))
	}

	// The export reads the same filters unpaged — the deleted version's
	// evidence row exports with its nickname fold.
	exported, err := adminUC.ExportConsents(entSubject(consentUser), &agreement.ListConsentsIn{
		PageRequest: &pageFirst,
		VersionNo:   utils.Ptr(e2eVersionNo),
	})
	if err != nil {
		t.Fatalf("export consents: %v", err)
	}
	if len(exported) != 1 || exported[0].Nickname == nil || *exported[0].Nickname != "e2e-"+consentUser {
		t.Fatalf("export fold: %+v, want the deleted version's row with its nickname", exported)
	}
}

func cleanAgreementE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	if _, err := db.AgreementConsent.Delete().Where(
		agreementconsent.VersionNoEQ(e2eVersionNo),
	).Exec(context.Background()); err != nil {
		t.Fatalf("clean consent rows: %v", err)
	}
	if _, err := db.AgreementConsent.Delete().Where(
		agreementconsent.UserIDEQ(consentUser),
	).Exec(context.Background()); err != nil {
		t.Fatalf("clean consent user rows: %v", err)
	}
	if _, err := db.AgreementVersion.Delete().Where(
		agreementversion.NoEQ(e2eVersionNo),
	).Exec(context.Background()); err != nil {
		t.Fatalf("clean version rows: %v", err)
	}
	// A run that died between moves can leave a document with no current
	// at all; the seed row is the cure. A current that exists — including
	// an operator's walkthrough publish — stays exactly as it is.
	rows, err := db.AgreementVersion.Query().All(context.Background())
	if err != nil {
		t.Fatalf("clean query versions: %v", err)
	}
	hasCurrent := map[string]bool{}
	byDoc := map[string]*ent.AgreementVersion{}
	for _, row := range rows {
		if row.IsCurrent {
			hasCurrent[string(row.Doc)] = true
		}
		if row.No == seededNo {
			byDoc[string(row.Doc)] = row
		}
	}
	for doc, seeded := range byDoc {
		if hasCurrent[doc] {
			continue
		}
		if _, err := seeded.Update().SetIsCurrent(true).Save(context.Background()); err != nil {
			t.Fatalf("clean restore current %s/%s: %v", doc, seeded.No, err)
		}
	}
}
