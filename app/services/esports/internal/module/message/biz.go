package message

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

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	typeAnnouncement  = "announcement"
	feedbackReplyKind = "feedback_reply"

	scopeAll        = "all"
	scopeMembership = "membership"

	kindBanner = "banner"

	statusDraft     = "draft"
	statusPublished = "published"
	statusOffline   = "offline"

	feedbackReplyTitle = "反馈回复" // the closure letter's fixed headline, zh because the member base reads zh — every writer sets a title, this one has only one
)

var rosterHydrate = utils.Ptr(int32(helper.DefaultPageSizeMax)) // one roster page per round-trip; the publish loop pages until the roster answers everyone

// DO ------------------------------------------------------------------------------------------------------------------

type Message struct {
	ID        string
	UserID    string
	Type      string // announcement / report / membership / feedback_reply
	Title     string
	Content   string
	RefID     *string
	IsRead    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Announcement struct {
	ID          string
	Kind        string // banner / message
	Scope       string // all / membership
	Title       string
	Content     string
	Status      string // draft / published / offline
	PublishedAt *time.Time
	OfflineAt   *time.Time
	PublishedBy *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListMessagesIn struct {
	*commonv1.PageRequest
	UserID string
}

type ListMessagesOut struct {
	*commonv1.PageResponse
	Rows   []*Message
	Unread int
}

// Port ----------------------------------------------------------------------------------------------------------------

type MessageRP interface {
	ListUserMessages(ctx context.Context, in *ListMessagesIn) (*ListMessagesOut, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	// The mark-read arbitration: one conditional UPDATE on the member's
	// own row — another device's earlier read, or a racing id, is the
	// same outcome as success, never an error.
	MarkRead(ctx context.Context, userID, messageID string) error
	MarkAllRead(ctx context.Context, userID string) error
	ListPublishedBanners(ctx context.Context) ([]*Announcement, error)
	// Bulk delivery, insert-only — the publish transaction's message
	// half; chunked so a whole-roster fan-out stays under the statement
	// parameter ceiling.
	EmitToUsers(ctx context.Context, userIDs []string, m *Message) error
	// The feedback reply closure: exactly one letter into the asker's
	// box, insert-only, inside the reply's transaction.
	RecordFeedbackReply(ctx context.Context, userID, feedbackID, content string) error
	CreateAnnouncement(ctx context.Context, a *Announcement) (*Announcement, error)
	// nil = absent.
	FindAnnouncement(ctx context.Context, id string) (*Announcement, error)
	// Draft-edit write: title/content, either optional. False = the row
	// left draft underneath the call — the caller answers the conflict.
	UpdateAnnouncementDraft(ctx context.Context, id string, title, content *string) (bool, error)
	// The flip is conditional on draft: false means a rival publish won
	// the row, and the caller's delivery unwinds with it.
	PublishAnnouncement(ctx context.Context, id string, publishedAt time.Time, publishedBy string) (bool, error)
	// Conditional on published, mirroring the publish flip.
	OfflineAnnouncement(ctx context.Context, id string, offlineAt time.Time) (bool, error)
	ListAnnouncements(ctx context.Context, kind, status *string) ([]*Announcement, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type MessageUC struct {
	shared.UC
	messageRP MessageRP
}

func NewMessageUC(logger *slog.Logger, tm shared.Transaction, messageRP MessageRP) *MessageUC {
	return &MessageUC{
		UC:        shared.NewUC(logger.With("module", "module/message"), tm),
		messageRP: messageRP,
	}
}

type MessageAdminUC struct {
	shared.UC
	messageRP MessageRP
	systemRP  member.SystemRP
	entRP     entitle.EntitlementRP
}

func NewMessageAdminUC(
	logger *slog.Logger, tm shared.Transaction,
	messageRP MessageRP, systemRP member.SystemRP, entRP entitle.EntitlementRP,
) *MessageAdminUC {
	return &MessageAdminUC{
		UC:        shared.NewUC(logger.With("module", "module/message"), tm),
		messageRP: messageRP,
		systemRP:  systemRP,
		entRP:     entRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *MessageUC) ListMessages(ctx context.Context, page *commonv1.PageRequest) (*ListMessagesOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("message listing requires a member session"))
	}
	return uc.messageRP.ListUserMessages(ctx, &ListMessagesIn{PageRequest: helper.EnsurePageRequest(page), UserID: subject.UserID})
}

func (uc *MessageUC) CountUnread(ctx context.Context) (int, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return 0, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("unread counting requires a member session"))
	}
	return uc.messageRP.CountUnread(ctx, subject.UserID)
}

func (uc *MessageUC) MarkRead(ctx context.Context, messageID string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("marking read requires a member session"))
	}
	return uc.messageRP.MarkRead(ctx, subject.UserID, messageID)
}

func (uc *MessageUC) MarkAllRead(ctx context.Context) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("marking read requires a member session"))
	}
	return uc.messageRP.MarkAllRead(ctx, subject.UserID)
}

func (uc *MessageUC) ListBanners(ctx context.Context) ([]*Announcement, error) {
	return uc.messageRP.ListPublishedBanners(ctx)
}

func (uc *MessageAdminUC) CreateAnnouncement(ctx context.Context, kind, scope, title, content string) (*Announcement, error) {
	_, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("announcement creation requires an admin session"))
	}
	return uc.messageRP.CreateAnnouncement(ctx, &Announcement{
		Kind: kind, Scope: scope, Title: title, Content: content, Status: statusDraft,
	})
}

func (uc *MessageAdminUC) UpdateAnnouncement(ctx context.Context, id string, title, content *string) (*Announcement, error) {
	row, err := uc.requireDraft(ctx, id)
	if err != nil {
		return nil, err
	}
	if title != nil || content != nil {
		flipped, err := uc.messageRP.UpdateAnnouncementDraft(ctx, id, title, content)
		if err != nil {
			return nil, err
		}
		if !flipped {
			return nil, esportspb.ErrorEsportsAnnouncementStateConflict("").WithCause(fmt.Errorf("announcement %s left draft underneath the edit; the lifecycle never walks back", id))
		}
	}
	row.Title = utils.Deref(title, row.Title)
	row.Content = utils.Deref(content, row.Content)
	return row, nil
}

func (uc *MessageAdminUC) PublishAnnouncement(ctx context.Context, id string) (*Announcement, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("publishing requires an admin session"))
	}
	row, err := uc.requireDraft(ctx, id)
	if err != nil {
		return nil, err
	}
	// The member set resolves before the transaction opens — delivery
	// is a snapshot of who is addressable now, like a letter mailed on
	// paper; members arriving later never receive it.
	var targets []string
	if row.Kind != kindBanner {
		switch row.Scope {
		case scopeMembership:
			targets, err = uc.liveMembers(ctx)
		default:
			targets, err = uc.wholeRoster(ctx)
		}
		if err != nil {
			return nil, err
		}
	}
	now := time.Now()
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// Status flip and delivery commit together: "published" and
		// "delivered" cannot disagree — a failed insert rolls the flip
		// back with it.
		if row.Kind != kindBanner {
			if err := uc.messageRP.EmitToUsers(txCtx, targets, &Message{
				Type: typeAnnouncement, Title: row.Title, Content: row.Content,
			}); err != nil {
				return err
			}
		}
		flipped, err := uc.messageRP.PublishAnnouncement(txCtx, id, now, subject.UserID)
		if err != nil {
			return err
		}
		if !flipped {
			return esportspb.ErrorEsportsAnnouncementStateConflict("").WithCause(fmt.Errorf("announcement %s left draft underneath the publish; a rival delivery already won it", id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return uc.messageRP.FindAnnouncement(ctx, id)
}

func (uc *MessageAdminUC) OfflineAnnouncement(ctx context.Context, id string) (*Announcement, error) {
	row, err := uc.messageRP.FindAnnouncement(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, esportspb.ErrorEsportsAnnouncementNotFound("").WithCause(fmt.Errorf("announcement %s not found", id))
	}
	if row.Status != statusPublished {
		return nil, esportspb.ErrorEsportsAnnouncementStateConflict("").WithCause(fmt.Errorf("announcement %s is %s; only a published row goes offline", id, row.Status))
	}
	flipped, err := uc.messageRP.OfflineAnnouncement(ctx, id, time.Now())
	if err != nil {
		return nil, err
	}
	if !flipped {
		return nil, esportspb.ErrorEsportsAnnouncementStateConflict("").WithCause(fmt.Errorf("announcement %s left published underneath the call; only a published row goes offline", id))
	}
	return uc.messageRP.FindAnnouncement(ctx, id)
}

func (uc *MessageAdminUC) ListAnnouncements(ctx context.Context, kind, status *string) ([]*Announcement, error) {
	return uc.messageRP.ListAnnouncements(ctx, kind, status)
}

func (uc *MessageAdminUC) ListUserMessages(ctx context.Context, userID string, page *commonv1.PageRequest) (*ListMessagesOut, error) {
	_, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("letterbox listing requires an admin session"))
	}
	return uc.messageRP.ListUserMessages(ctx, &ListMessagesIn{PageRequest: helper.EnsurePageRequest(page), UserID: userID})
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *MessageAdminUC) requireDraft(ctx context.Context, id string) (*Announcement, error) {
	row, err := uc.messageRP.FindAnnouncement(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, esportspb.ErrorEsportsAnnouncementNotFound("").WithCause(fmt.Errorf("announcement %s not found", id))
	}
	if row.Status != statusDraft {
		return nil, esportspb.ErrorEsportsAnnouncementStateConflict("").WithCause(fmt.Errorf("announcement %s is %s; the lifecycle never walks back", id, row.Status))
	}
	return row, nil
}

func (uc *MessageAdminUC) wholeRoster(ctx context.Context) ([]string, error) {
	// Pages until the roster has answered everyone — its Total ends the
	// loop, not the requested size, because the roster may clamp the
	// page below it.
	ids := make([]string, 0, 256)
	for page := int32(1); ; page++ {
		roster, err := uc.systemRP.ListRoster(ctx, &member.RosterQuery{
			Population: member.DefaultPopulation,
			// A pinned order keeps every page over one axis — an
			// unpinned offset walk can hand a member two letters or none.
			OrderBy: []string{"createdAt:asc"},
			Page:    &commonv1.PageRequest{PageNo: utils.Ptr(page), PageSize: rosterHydrate},
		})
		if err != nil {
			return nil, err
		}
		for _, u := range roster.Users {
			ids = append(ids, u.ID)
		}
		if len(roster.Users) == 0 || utils.ConvNum[int32](len(ids)) >= roster.GetTotal() {
			return ids, nil
		}
	}
}

func (uc *MessageAdminUC) liveMembers(ctx context.Context) ([]string, error) {
	// The whole membership ledger in one read; the live judgement — an
	// active row whose window still runs — is the same fold the
	// operator plane makes over these rows.
	rows, err := uc.entRP.ListMembershipRows(ctx, nil)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	seen := make(map[string]struct{}, len(rows))
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Status != shared.StatusActive || row.ExpiresAt == nil || !row.ExpiresAt.After(now) {
			continue
		}
		if _, dup := seen[row.UserID]; dup {
			continue
		}
		seen[row.UserID] = struct{}{}
		ids = append(ids, row.UserID)
	}
	return ids, nil
}
