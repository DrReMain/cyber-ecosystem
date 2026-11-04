package message

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/announcement"
	"cyber-ecosystem/app/services/esports/internal/ent/usermessage"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const emitChunk = 500 // one bulk insert's parameter budget — a whole-roster fan-out rides many bounded statements

// Repo ----------------------------------------------------------------------------------------------------------------

type messageRP struct {
	shared.RP
}

func NewMessageRP(logger *slog.Logger, p *platform.Platform) MessageRP {
	return &messageRP{RP: shared.NewRP(logger.With("module", "module/message_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *messageRP) ListUserMessages(ctx context.Context, in *ListMessagesIn) (*ListMessagesOut, error) {
	query := rp.Platform.GetClient(ctx).UserMessage.Query().
		Where(usermessage.UserIDEQ(in.UserID))
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.Order(ent.Desc(usermessage.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	unread, err := rp.CountUnread(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	return &ListMessagesOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Rows:         utils.SliceMap(rows, mapMessage),
		Unread:       unread,
	}, nil
}

func (rp *messageRP) CountUnread(ctx context.Context, userID string) (int, error) {
	count, err := rp.Platform.GetClient(ctx).UserMessage.Query().
		Where(usermessage.UserIDEQ(userID), usermessage.IsReadEQ(false)).
		Count(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return count, nil
}

func (rp *messageRP) MarkRead(ctx context.Context, userID, messageID string) error {
	// Exec, not Save: zero matching rows is the no-op outcome (already
	// read, another member's row, or a racing id), not a miss to surface.
	if err := rp.Platform.GetClient(ctx).UserMessage.Update().
		Where(usermessage.IDEQ(messageID), usermessage.UserIDEQ(userID), usermessage.IsReadEQ(false)).
		SetIsRead(true).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *messageRP) MarkAllRead(ctx context.Context, userID string) error {
	// The same conditional shape as the single-row mark: unread rows
	// flip, everything else is untouched.
	if err := rp.Platform.GetClient(ctx).UserMessage.Update().
		Where(usermessage.UserIDEQ(userID), usermessage.IsReadEQ(false)).
		SetIsRead(true).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *messageRP) ListPublishedBanners(ctx context.Context) ([]*Announcement, error) {
	// Published implies visible: the status flip and its delivery
	// moment commit together, so no time guard rides the query.
	rows, err := rp.Platform.GetClient(ctx).Announcement.Query().
		Where(
			announcement.KindEQ(announcement.Kind(kindBanner)),
			announcement.StatusEQ(announcement.Status(statusPublished)),
		).
		Order(ent.Desc(announcement.FieldPublishedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapAnnouncement), nil
}

func (rp *messageRP) EmitToUsers(ctx context.Context, userIDs []string, m *Message) error {
	client := rp.Platform.GetClient(ctx)
	for start := 0; start < len(userIDs); start += emitChunk {
		end := start + emitChunk
		if end > len(userIDs) {
			end = len(userIDs)
		}
		bulk := make([]*ent.UserMessageCreate, 0, end-start)
		for _, userID := range userIDs[start:end] {
			bulk = append(bulk, client.UserMessage.Create().
				SetUserID(userID).
				SetType(usermessage.Type(m.Type)).
				SetTitle(m.Title).
				SetContent(m.Content).
				SetNillableRefID(m.RefID))
		}
		if _, err := client.UserMessage.CreateBulk(bulk...).Save(ctx); err != nil {
			return rp.Platform.HandleEntError(err)
		}
	}
	return nil
}

func (rp *messageRP) RecordFeedbackReply(ctx context.Context, userID, feedbackID, content string) error {
	_, err := rp.Platform.GetClient(ctx).UserMessage.Create().
		SetUserID(userID).
		SetType(usermessage.Type(feedbackReplyKind)).
		SetTitle(feedbackReplyTitle).
		SetContent(content).
		SetRefID(feedbackID).
		Save(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *messageRP) CreateAnnouncement(ctx context.Context, a *Announcement) (*Announcement, error) {
	row, err := rp.Platform.GetClient(ctx).Announcement.Create().
		SetKind(announcement.Kind(a.Kind)).
		SetScope(announcement.Scope(a.Scope)).
		SetTitle(a.Title).
		SetContent(a.Content).
		SetStatus(announcement.Status(a.Status)).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapAnnouncement(row), nil
}

func (rp *messageRP) FindAnnouncement(ctx context.Context, id string) (*Announcement, error) {
	// nil row = absent; the callers read absence, not an error.
	row, err := rp.Platform.GetClient(ctx).Announcement.Query().
		Where(announcement.IDEQ(id)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapAnnouncement(row), nil
}

func (rp *messageRP) UpdateAnnouncementDraft(ctx context.Context, id string, title, content *string) (bool, error) {
	rows, err := rp.Platform.GetClient(ctx).Announcement.Update().
		Where(announcement.IDEQ(id), announcement.StatusEQ(announcement.Status(statusDraft))).
		SetNillableTitle(title).
		SetNillableContent(content).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *messageRP) PublishAnnouncement(ctx context.Context, id string, publishedAt time.Time, publishedBy string) (bool, error) {
	// Conditional on draft — the flip is the publish arbitration, so a
	// rival delivery that won the row answers zero rows here.
	rows, err := rp.Platform.GetClient(ctx).Announcement.Update().
		Where(announcement.IDEQ(id), announcement.StatusEQ(announcement.Status(statusDraft))).
		SetStatus(announcement.Status(statusPublished)).
		SetPublishedAt(publishedAt).
		SetPublishedBy(publishedBy).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *messageRP) OfflineAnnouncement(ctx context.Context, id string, offlineAt time.Time) (bool, error) {
	rows, err := rp.Platform.GetClient(ctx).Announcement.Update().
		Where(announcement.IDEQ(id), announcement.StatusEQ(announcement.Status(statusPublished))).
		SetStatus(announcement.Status(statusOffline)).
		SetOfflineAt(offlineAt).
		Save(ctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return rows > 0, nil
}

func (rp *messageRP) ListAnnouncements(ctx context.Context, kind, status *string) ([]*Announcement, error) {
	query := rp.Platform.GetClient(ctx).Announcement.Query()
	if kind != nil {
		query = query.Where(announcement.KindEQ(announcement.Kind(*kind)))
	}
	if status != nil {
		query = query.Where(announcement.StatusEQ(announcement.Status(*status)))
	}
	rows, err := query.Order(ent.Desc(announcement.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapAnnouncement), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapMessage(row *ent.UserMessage) *Message {
	return &Message{
		ID:        row.ID,
		UserID:    row.UserID,
		Type:      string(row.Type),
		Title:     row.Title,
		Content:   row.Content,
		RefID:     row.RefID,
		IsRead:    row.IsRead,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func mapAnnouncement(row *ent.Announcement) *Announcement {
	return &Announcement{
		ID:          row.ID,
		Kind:        string(row.Kind),
		Scope:       string(row.Scope),
		Title:       row.Title,
		Content:     row.Content,
		Status:      string(row.Status),
		PublishedAt: row.PublishedAt,
		OfflineAt:   row.OfflineAt,
		PublishedBy: row.PublishedBy,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
