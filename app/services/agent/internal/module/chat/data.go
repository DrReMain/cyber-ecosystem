package chat

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/agent/internal/ent"
	"cyber-ecosystem/app/services/agent/internal/ent/chatmessage"
	"cyber-ecosystem/app/services/agent/internal/ent/chatsession"
	"cyber-ecosystem/app/services/agent/internal/platform"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Repo ------------------------------------------------------------------------------------------------------------------

type chatRP struct {
	shared.RP
}

func NewChatRP(logger *slog.Logger, p *platform.Platform) ChatRP {
	return &chatRP{RP: shared.NewRP(logger.With("module", "module/chat_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *chatRP) FindSessionByID(ctx context.Context, ownerID, sessionID string) (*ChatSession, error) {
	d, err := rp.Platform.GetClient(ctx).ChatSession.Query().
		Where(chatsession.IDEQ(sessionID), chatsession.OwnerIDEQ(ownerID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapChatSession(d), nil
}

func (rp *chatRP) LatestTurns(ctx context.Context, sessionID string, limit int) ([]*Turn, error) {
	rows, err := rp.Platform.GetClient(ctx).ChatMessage.Query().
		Where(chatmessage.SessionIDEQ(sessionID)).
		Order(ent.Desc(chatmessage.FieldCreatedAt), ent.Desc(chatmessage.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, func(d *ent.ChatMessage) *Turn {
		return &Turn{Role: d.Role, Content: d.Content}
	}), nil
}

func (rp *chatRP) CreateWithExchange(ctx context.Context, sess *ChatSession, user, assistant *SessionMessage) error {
	c := rp.Platform.GetClient(ctx)
	if err := c.ChatSession.Create().
		SetID(sess.ID).
		SetOwnerID(sess.OwnerID).
		SetTitle(sess.Title).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return rp.createMessages(ctx, c, sess.ID, user, assistant)
}

func (rp *chatRP) AppendExchange(ctx context.Context, sessionID string, user, assistant *SessionMessage) error {
	c := rp.Platform.GetClient(ctx)
	// Message inserts do not touch the session row; the explicit bump keeps
	// the (owner_id, updated_at) rail ordering truthful.
	if err := c.ChatSession.UpdateOneID(sessionID).SetUpdatedAt(time.Now()).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return rp.createMessages(ctx, c, sessionID, user, assistant)
}

func (rp *chatRP) TouchSession(ctx context.Context, ownerID, sessionID string) error {
	// Send-time ordering bump; a concurrent delete makes it a no-op.
	if _, err := rp.Platform.GetClient(ctx).ChatSession.Update().
		Where(chatsession.IDEQ(sessionID), chatsession.OwnerIDEQ(ownerID)).
		SetUpdatedAt(time.Now()).
		Save(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *chatRP) UpdateSessionTitle(ctx context.Context, ownerID, sessionID, title string) (bool, error) {
	n, err := rp.Platform.GetClient(ctx).ChatSession.Update().
		Where(chatsession.IDEQ(sessionID), chatsession.OwnerIDEQ(ownerID)).
		SetTitle(title).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return n > 0, nil
}

func (rp *chatRP) ListSessions(ctx context.Context, ownerID string, page *commonpb.PageRequest) (*ListSessionsOut, error) {
	query := rp.Platform.GetClient(ctx).ChatSession.Query().
		Where(chatsession.OwnerIDEQ(ownerID)).
		Order(ent.Desc(chatsession.FieldUpdatedAt), ent.Desc(chatsession.FieldID))
	total, offset, limit, err := shared.Paginate(ctx, query, page, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, err
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListSessionsOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(rows, mapChatSession),
	}, nil
}

func (rp *chatRP) ListMessages(ctx context.Context, sessionID string, page *commonpb.PageRequest) (*ListMessagesOut, error) {
	query := rp.Platform.GetClient(ctx).ChatMessage.Query().
		Where(chatmessage.SessionIDEQ(sessionID)).
		Order(ent.Desc(chatmessage.FieldCreatedAt), ent.Desc(chatmessage.FieldID))
	total, offset, limit, err := shared.Paginate(ctx, query, page, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, err
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListMessagesOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(rows, mapChatMessage),
	}, nil
}

func (rp *chatRP) DeleteMessagesBySession(ctx context.Context, sessionID string) error {
	if _, err := rp.Platform.GetClient(ctx).ChatMessage.Delete().
		Where(chatmessage.SessionIDEQ(sessionID)).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *chatRP) DeleteSessionByID(ctx context.Context, ownerID, sessionID string) error {
	if _, err := rp.Platform.GetClient(ctx).ChatSession.Delete().
		Where(chatsession.IDEQ(sessionID), chatsession.OwnerIDEQ(ownerID)).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *chatRP) createMessages(ctx context.Context, c *ent.Client, sessionID string, user, assistant *SessionMessage) error {
	// One bulk insert keeps the pair in a single statement: same created_at,
	// xid counter order puts the user turn before the assistant one.
	creates := []*ent.ChatMessageCreate{
		c.ChatMessage.Create().SetSessionID(sessionID).SetRole(user.Role).SetContent(user.Content),
		c.ChatMessage.Create().SetSessionID(sessionID).SetRole(assistant.Role).SetContent(assistant.Content).
			SetReasoning(assistant.Reasoning).SetModel(assistant.Model).SetFinish(assistant.Finish),
	}
	if err := c.ChatMessage.CreateBulk(creates...).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func mapChatSession(d *ent.ChatSession) *ChatSession {
	return &ChatSession{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		OwnerID:   d.OwnerID,
		Title:     d.Title,
	}
}

func mapChatMessage(d *ent.ChatMessage) *SessionMessage {
	return &SessionMessage{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		Role:      d.Role,
		Content:   d.Content,
		Reasoning: d.Reasoning,
		Model:     d.Model,
		Finish:    d.Finish,
	}
}
