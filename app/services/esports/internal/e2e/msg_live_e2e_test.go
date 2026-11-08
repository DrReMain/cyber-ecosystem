package e2e

// The message/feedback domain's live probe: the real database, the
// domain's own faces end to end — announcement draft→publish→delivery
// into letterboxes (all-scope and membership-scope), unread counting
// and mark-read, the banner read face, the feedback conversation
// (submit→reply with its same-transaction letterbox closure→follow-up
// walking back to pending→close) and the refusal arms. Gated: MSG_E2E=1
// with the dev database up runs it; anything else skips.

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/utils"

	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/feedback"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/message"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

// The publish fan-out reads the roster; this stub answers the two probe
// members, one page, so the all-scope delivery lands on exactly them.
type msgRosterRP struct {
	ids []string
}

func (r msgRosterRP) ListRoster(ctx context.Context, q *member.RosterQuery) (*member.RosterPage, error) {
	return &member.RosterPage{
		PageResponse: &commonv1.PageResponse{Total: utils.ConvNum[int32](len(r.ids)), More: false},
		Users: []*member.RosterUser{
			{ID: r.ids[0], Enabled: true, CreatedAt: time.Now()},
			{ID: r.ids[1], Enabled: true, CreatedAt: time.Now()},
		},
	}, nil
}

func TestMsgLiveE2E(t *testing.T) {
	if os.Getenv("MSG_E2E") != "1" {
		t.Skip("set MSG_E2E=1 with the dev database up to run the message/feedback end-to-end")
	}
	const (
		uidA     = "e2emsguser0000000010" // MaxRuneLen 20
		uidB     = "e2emsguser0000000020"
		admin1   = "e2emsgoper0000000001" // 20 runes, like every subject id the store expects
		absentID = "e2emsgabsent00000000" // matches no row ever created
	)
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: envOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: envOr("DB_PASS", "postgres"), DbName: envOr("DB_NAME", "esports"),
		MaxOpenConns: 4,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	payments, closePayments, err := platform.NewPayments(&conf.Payment{
		DefaultProvider: "mock",
		NotifyBaseUrl:   "http://localhost:5174",
		Providers:       []*conf.Payment_Provider{{Name: "mock", Kind: "mock"}},
	})
	if err != nil {
		t.Fatalf("payments: %v", err)
	}
	entHandler := platform.NewEntErrorHandler()
	payHandler, err := platform.NewPaymentErrorHandler()
	if err != nil {
		t.Fatalf("pay handler: %v", err)
	}
	pf, err := platform.NewPlatform(db, entHandler, payments, payHandler, nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanMsgE2E(t, db, uidA, uidB)
	t.Cleanup(func() {
		cleanMsgE2E(t, db, uidA, uidB)
		closePayments()
		closeDB()
	})

	msgRP := message.NewMessageRP(logger, pf)
	msgUC := message.NewMessageUC(logger, pf, msgRP)
	msgAdmin := message.NewMessageAdminUC(logger, pf, msgRP, msgRosterRP{ids: []string{uidA, uidB}}, entitle.NewEntitlementRP(logger, pf))
	fbRP := feedback.NewFeedbackRP(logger, pf)
	fbUC := feedback.NewFeedbackUC(logger, pf, fbRP)
	fbAdmin := feedback.NewFeedbackAdminUC(logger, pf, fbRP, profile.NewProfileRP(logger, pf), msgRP)

	adminCtx := entSubject(admin1)
	userA, userB := entSubject(uidA), entSubject(uidB)

	// Announcements: draft → publish (all scope) → both letterboxes.
	draft, err := msgAdmin.CreateAnnouncement(adminCtx, "message", "all", "E2E MSG 赛季公告", "反应速度权重由 1.0 调整为 1.15。")
	failChain(t, "create announcement", err)
	if draft.Status != "draft" {
		t.Fatalf("created status = %s, want draft", draft.Status)
	}
	published, err := msgAdmin.PublishAnnouncement(adminCtx, draft.ID)
	failChain(t, "publish announcement", err)
	if published.Status != "published" || published.PublishedBy == nil || *published.PublishedBy != admin1 {
		t.Fatalf("published shape: %+v", published)
	}

	// Delivery landed on both roster members, unread and listed.
	for _, uid := range []string{uidA, uidB} {
		out, err := msgUC.ListMessages(entSubject(uid), nil)
		failChain(t, "list messages "+uid, err)
		if out.Unread != 1 || len(out.Rows) != 1 || out.Rows[0].Type != "announcement" || out.Rows[0].Title != "E2E MSG 赛季公告" {
			t.Fatalf("delivery %s: unread=%d rows=%+v", uid, out.Unread, out.Rows)
		}
	}

	// Mark one read: that member's badge empties, the other's stays.
	if err := msgUC.MarkRead(userA, mustOneMessage(t, msgUC, uidA)); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if n, err := msgUC.CountUnread(userA); err != nil || n != 0 {
		t.Fatalf("unread after read: n=%d err=%v, want 0", n, err)
	}
	if n, err := msgUC.CountUnread(userB); err != nil || n != 1 {
		t.Fatalf("other member's unread: n=%d err=%v, want 1", n, err)
	}
	if err := msgUC.MarkAllRead(userB); err != nil {
		t.Fatalf("mark all read: %v", err)
	}

	// Republish refuses: the lifecycle never walks back.
	if _, err := msgAdmin.PublishAnnouncement(adminCtx, draft.ID); err == nil || !strings.Contains(err.Error(), "ANNOUNCEMENT_STATE_CONFLICT") {
		t.Fatalf("republish refusal: err=%v, want state conflict", err)
	}
	// Absent announcement answers not-found, never a phantom flip.
	if _, err := msgAdmin.PublishAnnouncement(adminCtx, absentID); err == nil || !strings.Contains(err.Error(), "ANNOUNCEMENT_NOT_FOUND") {
		t.Fatalf("absent publish refusal: err=%v, want not found", err)
	}

	// Banner kind: publish delivers nothing, the read face answers.
	banner, err := msgAdmin.CreateAnnouncement(adminCtx, "banner", "all", "E2E MSG 横幅", "新版本已上线。")
	failChain(t, "create banner", err)
	if _, err := msgAdmin.PublishAnnouncement(adminCtx, banner.ID); err != nil {
		t.Fatalf("publish banner: %v", err)
	}
	banners, err := msgUC.ListBanners(userA)
	failChain(t, "list banners", err)
	if len(banners) != 1 || banners[0].Title != "E2E MSG 横幅" || banners[0].PublishedAt == nil {
		t.Fatalf("banner read: %+v", banners)
	}
	if n, err := msgUC.CountUnread(userA); err != nil || n != 0 {
		t.Fatalf("banner delivered letters: n=%d err=%v, want 0 (banner kind delivers nothing)", n, err)
	}

	// Membership scope: only the live member receives the letter.
	if _, err := pf.GetClient(context.Background()).Entitlement.Create().
		SetUserID(uidA).
		SetScopeKind(entitlement.ScopeKind("membership")).
		SetStatus(entitlement.Status("active")).
		SetExpiresAt(time.Now().AddDate(0, 0, 31)).
		SetSource("e2e-msg").
		Save(context.Background()); err != nil {
		t.Fatalf("membership seed: %v", err)
	}
	mDraft, err := msgAdmin.CreateAnnouncement(adminCtx, "message", "membership", "E2E MSG 会员公告", "会员权益说明更新。")
	failChain(t, "create membership announcement", err)
	if _, err := msgAdmin.PublishAnnouncement(adminCtx, mDraft.ID); err != nil {
		t.Fatalf("publish membership announcement: %v", err)
	}
	if n, err := msgUC.CountUnread(userA); err != nil || n != 1 {
		t.Fatalf("membership delivery to holder: n=%d err=%v, want 1", n, err)
	}
	if n, err := msgUC.CountUnread(userB); err != nil || n != 0 {
		t.Fatalf("membership delivery to non-holder: n=%d err=%v, want 0", n, err)
	}

	// The feedback conversation: submit → reply (closure letter in the
	// same transaction) → follow-up walks back to pending → close →
	// follow-up refused.
	fb, err := fbUC.Submit(userA, "feature_bug", "匹配总是崩,开始测评有时白屏,重进又好了,请排查", nil)
	failChain(t, "submit feedback", err)
	if fb.Status != "pending" {
		t.Fatalf("submitted status = %s, want pending", fb.Status)
	}
	if _, err := fbAdmin.Reply(adminCtx, fb.ID, "已定位:微信内 H5 唤起被回收导致白屏,v1.3.x 已修复,请更新后重试。"); err != nil {
		t.Fatalf("reply: %v", err)
	}
	// The closure letter: unread +1, typed feedback_reply, ref = the ticket.
	out, err := msgUC.ListMessages(userA, nil)
	failChain(t, "list after reply", err)
	var closure *message.Message
	for _, m := range out.Rows {
		if m.Type == "feedback_reply" {
			closure = m
		}
	}
	if closure == nil || closure.RefID == nil || *closure.RefID != fb.ID {
		t.Fatalf("closure letter: %+v (feedback %s)", closure, fb.ID)
	}
	if n, err := msgUC.CountUnread(userA); err != nil || n != 2 {
		t.Fatalf("unread after reply: n=%d err=%v, want 2 (membership + closure)", n, err)
	}
	afterFollowUp, err := fbUC.FollowUp(userA, fb.ID, "更新后仍复现,继续追问")
	failChain(t, "follow up", err)
	if afterFollowUp.Status != "pending" {
		t.Fatalf("follow-up status = %s, want pending again", afterFollowUp.Status)
	}
	rowFollowUp, followUpThread, err := fbUC.GetMine(userA, fb.ID)
	if err != nil || rowFollowUp.Status != "pending" {
		t.Fatalf("member detail after follow-up: %+v err=%v", rowFollowUp, err)
	}
	if len(followUpThread) != 3 { // submit + admin reply + user follow-up
		t.Fatalf("member thread after follow-up: n=%d, want 3", len(followUpThread))
	}
	// A foreign member reads the ticket: not-owned, never silent.
	if _, _, err := fbUC.GetMine(userB, fb.ID); err == nil || !strings.Contains(err.Error(), "FEEDBACK_NOT_OWNED") {
		t.Fatalf("foreign read refusal: err=%v, want not-owned", err)
	}
	// Absent ticket answers not-found on the member face.
	if _, _, err := fbUC.GetMine(userA, absentID); err == nil || !strings.Contains(err.Error(), "FEEDBACK_NOT_FOUND") {
		t.Fatalf("absent read refusal: err=%v, want not found", err)
	}
	if _, err := fbAdmin.Close(adminCtx, fb.ID); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := fbUC.FollowUp(userA, fb.ID, "已关闭后追问"); err == nil || !strings.Contains(err.Error(), "FEEDBACK_CLOSED") {
		t.Fatalf("closed follow-up refusal: err=%v, want closed", err)
	}
	// The operator's queue read: row present, nickname the lazy null.
	tickets, page, err := fbAdmin.List(adminCtx, nil, nil)
	failChain(t, "admin list", err)
	if page.Total == 0 {
		t.Fatalf("admin list empty")
	}
	var mine *feedback.AdminTicketRow
	for _, row := range tickets {
		if row.ID == fb.ID {
			mine = row
		}
	}
	if mine == nil || mine.Nickname != nil {
		t.Fatalf("admin ticket row: %+v", mine)
	}
	detail, detailThread, err := fbAdmin.Get(adminCtx, fb.ID)
	failChain(t, "admin detail", err)
	if detail == nil || len(detailThread) != 3 {
		t.Fatalf("admin thread: n=%d, want 3", len(detailThread))
	}
}

// Private -------------------------------------------------------------------------------------------------------------

func mustOneMessage(t *testing.T, uc *message.MessageUC, uid string) string {
	t.Helper()
	out, err := uc.ListMessages(entSubject(uid), nil)
	if err != nil || len(out.Rows) == 0 {
		t.Fatalf("messages for mark-read: n=%d err=%v", len(out.Rows), err)
	}
	return out.Rows[0].ID
}

// failChain prints the whole unwrap chain — the raw store error at the
// bottom is the diagnosis; a bare %v would only name the wrapper.
func failChain(t *testing.T, stage string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	var sb strings.Builder
	for e := err; e != nil; e = errors.Unwrap(e) {
		sb.WriteString("\n  -> ")
		sb.WriteString(e.Error())
	}
	t.Fatalf("%s:%s", stage, sb.String())
}

func cleanMsgE2E(t *testing.T, db *ent.Client, uidA, uidB string) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []struct {
		q    string
		args []any
	}{
		{`DELETE FROM user_message WHERE user_id IN ($1, $2)`, []any{uidA, uidB}},
		{`DELETE FROM feedback_message WHERE feedback_id IN (SELECT id FROM user_feedback WHERE user_id IN ($1, $2))`, []any{uidA, uidB}},
		{`DELETE FROM user_feedback WHERE user_id IN ($1, $2)`, []any{uidA, uidB}},
		{`DELETE FROM announcement WHERE title LIKE 'E2E MSG%'`, nil},
		{`DELETE FROM entitlement WHERE user_id IN ($1, $2)`, []any{uidA, uidB}},
	} {
		if _, err := db.ExecContext(ctx, stmt.q, stmt.args...); err != nil {
			t.Fatalf("msg e2e clean: %v", err)
		}
	}
}
