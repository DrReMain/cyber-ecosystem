package feedback

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/feedbackmessage"
	"cyber-ecosystem/app/services/esports/internal/ent/userfeedback"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type feedbackRP struct {
	shared.RP
}

func NewFeedbackRP(logger *slog.Logger, p *platform.Platform) FeedbackRP {
	return &feedbackRP{RP: shared.NewRP(logger.With("module", "module/feedback_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *feedbackRP) Create(ctx context.Context, f *Feedback, content string) (*Feedback, error) {
	client := rp.Platform.GetClient(ctx)
	// Root and opening turn in the caller's transaction: a ticket whose
	// first message never landed must not exist.
	root, err := client.UserFeedback.Create().
		SetUserID(f.UserID).
		SetCategory(userfeedback.Category(f.Category)).
		SetContent(content).
		SetNillableContact(f.Contact).
		SetStatus(userfeedback.Status(f.Status)).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if _, err := client.FeedbackMessage.Create().
		SetFeedbackID(root.ID).
		SetSender(feedbackmessage.Sender("user")).
		SetContent(content).
		Save(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFeedback(root), nil
}

func (rp *feedbackRP) ListByUser(ctx context.Context, in *ListFeedbackIn) (*ListFeedbackOut, error) {
	query := rp.Platform.GetClient(ctx).UserFeedback.Query().
		Where(userfeedback.UserIDEQ(in.UserID))
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	// Most recently active first — a fresh answer or follow-up jumps
	// the queue; updated_at moves with every turn.
	rows, err := query.Order(ent.Desc(userfeedback.FieldUpdatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListFeedbackOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Rows:         utils.SliceMap(rows, mapFeedback),
	}, nil
}

func (rp *feedbackRP) Find(ctx context.Context, id string) (*Feedback, error) {
	// nil row = absent; the callers read absence, not an error.
	row, err := rp.Platform.GetClient(ctx).UserFeedback.Query().
		Where(userfeedback.IDEQ(id)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFeedback(row), nil
}

func (rp *feedbackRP) Thread(ctx context.Context, feedbackID string) ([]*FeedbackMessage, error) {
	rows, err := rp.Platform.GetClient(ctx).FeedbackMessage.Query().
		Where(feedbackmessage.FeedbackIDEQ(feedbackID)).
		Order(ent.Asc(feedbackmessage.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapFeedbackMessage), nil
}

func (rp *feedbackRP) AppendUserTurn(ctx context.Context, id, content string) (bool, error) {
	client := rp.Platform.GetClient(ctx)
	// The flip arbitrates first, conditional on not-closed — a close
	// that wins the ticket underneath this turn leaves zero rows and
	// the caller unwinds instead of reopening a terminal ticket.
	rows, err := client.UserFeedback.Update().
		Where(userfeedback.IDEQ(id), userfeedback.StatusNEQ(userfeedback.Status(statusClosed))).
		SetStatus(userfeedback.Status(statusPending)).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	if rows == 0 {
		return false, nil
	}
	if _, err := client.FeedbackMessage.Create().
		SetFeedbackID(id).
		SetSender(feedbackmessage.Sender("user")).
		SetContent(content).
		Save(ctx); err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return true, nil
}

func (rp *feedbackRP) AppendAdminTurn(ctx context.Context, id, adminID, content string) (bool, error) {
	client := rp.Platform.GetClient(ctx)
	// Same not-closed arbitration as the member turn; replied is the
	// walk this side makes.
	rows, err := client.UserFeedback.Update().
		Where(userfeedback.IDEQ(id), userfeedback.StatusNEQ(userfeedback.Status(statusClosed))).
		SetStatus(userfeedback.Status(statusReplied)).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	if rows == 0 {
		return false, nil
	}
	if _, err := client.FeedbackMessage.Create().
		SetFeedbackID(id).
		SetSender(feedbackmessage.Sender("admin")).
		SetContent(content).
		SetAdminID(adminID).
		Save(ctx); err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return true, nil
}

func (rp *feedbackRP) Close(ctx context.Context, id string) error {
	// Exec, not Save: zero matching rows is the idempotent repeat (the
	// row already closed), not a miss to surface.
	if err := rp.Platform.GetClient(ctx).UserFeedback.Update().
		Where(userfeedback.IDEQ(id), userfeedback.StatusNEQ(userfeedback.Status(statusClosed))).
		SetStatus(userfeedback.Status(statusClosed)).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *feedbackRP) ListTickets(ctx context.Context, in *ListTicketsIn) (*ListFeedbackOut, error) {
	query := rp.Platform.GetClient(ctx).UserFeedback.Query()
	if in.Status != nil {
		query = query.Where(userfeedback.StatusEQ(userfeedback.Status(*in.Status)))
	}
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(userfeedback.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListFeedbackOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Rows:         utils.SliceMap(rows, mapFeedback),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapFeedback(row *ent.UserFeedback) *Feedback {
	return &Feedback{
		ID:        row.ID,
		UserID:    row.UserID,
		Category:  string(row.Category),
		Content:   row.Content,
		Contact:   row.Contact,
		Status:    string(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func mapFeedbackMessage(row *ent.FeedbackMessage) *FeedbackMessage {
	return &FeedbackMessage{
		ID:         row.ID,
		FeedbackID: row.FeedbackID,
		Sender:     string(row.Sender),
		Content:    row.Content,
		AdminID:    row.AdminID,
		CreatedAt:  row.CreatedAt,
	}
}
