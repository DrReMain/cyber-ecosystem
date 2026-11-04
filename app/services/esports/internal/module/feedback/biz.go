package feedback

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/message"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	statusPending = "pending"
	statusReplied = "replied"
	statusClosed  = "closed"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Feedback struct {
	ID        string
	UserID    string
	Category  string // feature_bug / assessment / report / payment / account / other
	Content   string
	Contact   *string
	Status    string // pending / replied / closed
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FeedbackMessage struct {
	ID         string
	FeedbackID string
	Sender     string // user / admin
	Content    string
	AdminID    *string
	CreatedAt  time.Time
}

type AdminTicketRow struct {
	// The operator's ticket row: the root with the nickname fold, so
	// the queue reads one table instead of joining the profile plane.
	Feedback
	Nickname *string
}

type ListFeedbackIn struct {
	*commonv1.PageRequest
	UserID string
}

type ListTicketsIn struct {
	*commonv1.PageRequest
	Status *string
}

type ListFeedbackOut struct {
	*commonv1.PageResponse
	Rows []*Feedback
}

// Port ----------------------------------------------------------------------------------------------------------------

type FeedbackRP interface {
	// Root and opening turn land together — a ticket without its first
	// message must not exist. The created root returns with its id.
	Create(ctx context.Context, f *Feedback, content string) (*Feedback, error)
	ListByUser(ctx context.Context, in *ListFeedbackIn) (*ListFeedbackOut, error)
	// nil = absent.
	Find(ctx context.Context, id string) (*Feedback, error)
	Thread(ctx context.Context, feedbackID string) ([]*FeedbackMessage, error)
	// The member's turn lands and the root walks back to pending — one
	// move the caller wraps in a transaction, never two. The flip is
	// conditional on not-closed: false means a close won the ticket
	// underneath the turn, and the caller's transaction unwinds.
	AppendUserTurn(ctx context.Context, id, content string) (bool, error)
	// The operator's turn lands and the root flips to replied — the
	// same one-move shape, same not-closed condition.
	AppendAdminTurn(ctx context.Context, id, adminID, content string) (bool, error)
	// The terminal move, one conditional UPDATE: any earlier state
	// closes, an already-closed row is the no-op repeat.
	Close(ctx context.Context, id string) error
	ListTickets(ctx context.Context, in *ListTicketsIn) (*ListFeedbackOut, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type FeedbackUC struct {
	shared.UC
	feedbackRP FeedbackRP
}

func NewFeedbackUC(logger *slog.Logger, tm shared.Transaction, feedbackRP FeedbackRP) *FeedbackUC {
	return &FeedbackUC{
		UC:         shared.NewUC(logger.With("module", "module/feedback"), tm),
		feedbackRP: feedbackRP,
	}
}

type FeedbackAdminUC struct {
	shared.UC
	feedbackRP FeedbackRP
	profileRP  profile.ProfileRP
	messageRP  message.MessageRP
}

func NewFeedbackAdminUC(
	logger *slog.Logger, tm shared.Transaction,
	feedbackRP FeedbackRP, profileRP profile.ProfileRP, messageRP message.MessageRP,
) *FeedbackAdminUC {
	return &FeedbackAdminUC{
		UC:         shared.NewUC(logger.With("module", "module/feedback"), tm),
		feedbackRP: feedbackRP,
		profileRP:  profileRP,
		messageRP:  messageRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *FeedbackUC) Submit(ctx context.Context, category, content string, contact *string) (*Feedback, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("feedback submission requires a member session"))
	}
	created := &Feedback{
		UserID: subject.UserID, Category: category, Content: content,
		Contact: contact, Status: statusPending,
	}
	var out *Feedback
	err := uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		row, err := uc.feedbackRP.Create(txCtx, created, content)
		out = row
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (uc *FeedbackUC) ListMine(ctx context.Context, page *commonv1.PageRequest) (*ListFeedbackOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("feedback listing requires a member session"))
	}
	return uc.feedbackRP.ListByUser(ctx, &ListFeedbackIn{PageRequest: helper.EnsurePageRequest(page), UserID: subject.UserID})
}

func (uc *FeedbackUC) GetMine(ctx context.Context, id string) (*Feedback, []*FeedbackMessage, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("feedback reading requires a member session"))
	}
	row, err := uc.owned(ctx, id, subject.UserID)
	if err != nil {
		return nil, nil, err
	}
	thread, err := uc.feedbackRP.Thread(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return row, thread, nil
}

func (uc *FeedbackUC) FollowUp(ctx context.Context, id, content string) (*Feedback, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("a follow-up requires a member session"))
	}
	row, err := uc.owned(ctx, id, subject.UserID)
	if err != nil {
		return nil, err
	}
	if row.Status == statusClosed {
		return nil, esportspb.ErrorEsportsFeedbackClosed("").WithCause(fmt.Errorf("feedback %s is closed; a new question opens a new ticket", id))
	}
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		flipped, err := uc.feedbackRP.AppendUserTurn(txCtx, id, content)
		if err != nil {
			return err
		}
		if !flipped {
			return esportspb.ErrorEsportsFeedbackClosed("").WithCause(fmt.Errorf("feedback %s closed underneath the follow-up; a new question opens a new ticket", id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return uc.feedbackRP.Find(ctx, id)
}

func (uc *FeedbackAdminUC) List(ctx context.Context, page *commonv1.PageRequest, status *string) ([]*AdminTicketRow, *commonv1.PageResponse, error) {
	out, err := uc.feedbackRP.ListTickets(ctx, &ListTicketsIn{PageRequest: helper.EnsurePageRequest(page), Status: status})
	if err != nil {
		return nil, nil, err
	}
	rows := make([]*AdminTicketRow, 0, len(out.Rows))
	ids := make([]string, 0, len(out.Rows))
	for _, f := range out.Rows {
		rows = append(rows, &AdminTicketRow{Feedback: *f})
		ids = append(ids, f.UserID)
	}
	if len(ids) > 0 {
		profiles, err := uc.profileRP.ListByUserIDs(ctx, ids)
		if err != nil {
			return nil, nil, err
		}
		nickname := make(map[string]*string, len(profiles))
		for _, p := range profiles {
			nickname[p.UserID] = p.Nickname
		}
		for _, row := range rows {
			row.Nickname = nickname[row.UserID]
		}
	}
	return rows, out.PageResponse, nil
}

func (uc *FeedbackAdminUC) Get(ctx context.Context, id string) (*AdminTicketRow, []*FeedbackMessage, error) {
	row, err := uc.requireTicket(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	thread, err := uc.feedbackRP.Thread(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	out := &AdminTicketRow{Feedback: *row}
	if profiles, err := uc.profileRP.ListByUserIDs(ctx, []string{row.UserID}); err != nil {
		return nil, nil, err
	} else if len(profiles) > 0 {
		out.Nickname = profiles[0].Nickname
	}
	return out, thread, nil
}

func (uc *FeedbackAdminUC) Reply(ctx context.Context, id, content string) (*Feedback, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("replying requires an admin session"))
	}
	row, err := uc.requireTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == statusClosed {
		return nil, esportspb.ErrorEsportsFeedbackClosed("").WithCause(fmt.Errorf("feedback %s is closed; closed tickets take no further turns", id))
	}
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// The answer is one move with two effects: the ticket's turn
		// and the member's letterbox letter commit together, so the
		// message center's notice can never point at an answer that
		// did not land.
		flipped, err := uc.feedbackRP.AppendAdminTurn(txCtx, id, subject.UserID, content)
		if err != nil {
			return err
		}
		if !flipped {
			return esportspb.ErrorEsportsFeedbackClosed("").WithCause(fmt.Errorf("feedback %s closed underneath the reply; closed tickets take no further turns", id))
		}
		return uc.messageRP.RecordFeedbackReply(txCtx, row.UserID, id, content)
	})
	if err != nil {
		return nil, err
	}
	return uc.feedbackRP.Find(ctx, id)
}

func (uc *FeedbackAdminUC) Close(ctx context.Context, id string) (*Feedback, error) {
	_, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("closing requires an admin session"))
	}
	if _, err := uc.requireTicket(ctx, id); err != nil {
		return nil, err
	}
	if err := uc.feedbackRP.Close(ctx, id); err != nil {
		return nil, err
	}
	return uc.feedbackRP.Find(ctx, id)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *FeedbackUC) owned(ctx context.Context, id, userID string) (*Feedback, error) {
	// The member plane's own-key check: absent answers not-found,
	// present but foreign answers not-owned — support can tell a lost
	// id from a borrowed one.
	row, err := uc.feedbackRP.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, esportspb.ErrorEsportsFeedbackNotFound("").WithCause(fmt.Errorf("feedback %s not found", id))
	}
	if row.UserID != userID {
		return nil, esportspb.ErrorEsportsFeedbackNotOwned("").WithCause(fmt.Errorf("feedback %s belongs to another member", id))
	}
	return row, nil
}

func (uc *FeedbackAdminUC) requireTicket(ctx context.Context, id string) (*Feedback, error) {
	// The operator plane's find-or-404: the member plane answers absent
	// and foreign differently, the queue only knows present or not.
	row, err := uc.feedbackRP.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, esportspb.ErrorEsportsFeedbackNotFound("").WithCause(fmt.Errorf("feedback %s not found", id))
	}
	return row, nil
}
