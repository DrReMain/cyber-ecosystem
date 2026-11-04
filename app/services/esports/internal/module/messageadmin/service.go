package messageadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/message"
)

// Struct --------------------------------------------------------------------------------------------------------------

type MessageAdminService struct {
	esportspb.UnimplementedMessageAdminServiceServer

	log            *slog.Logger
	messageAdminUC *message.MessageAdminUC
}

func NewMessageAdminService(logger *slog.Logger, messageAdminUC *message.MessageAdminUC) *MessageAdminService {
	return &MessageAdminService{
		log:            logger.With("module", "module/messageadmin_service"),
		messageAdminUC: messageAdminUC,
	}
}

func (s *MessageAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterMessageAdminServiceServer(srv, s)
}

func (s *MessageAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterMessageAdminServiceHTTPServer(srv, s)
}

func (s *MessageAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterMessageAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *MessageAdminService) CreateAnnouncement(ctx context.Context, in *esportspb.CreateAnnouncementRequest) (*esportspb.CreateAnnouncementResponse, error) {
	row, err := s.messageAdminUC.CreateAnnouncement(ctx,
		kindFromPB(in.GetKind()), scopeFromPB(in.GetScope()), in.GetTitle(), in.GetContent())
	if err != nil {
		return nil, err
	}
	return &esportspb.CreateAnnouncementResponse{Announcement: announcementPB(row)}, nil
}

func (s *MessageAdminService) UpdateAnnouncement(ctx context.Context, in *esportspb.UpdateAnnouncementRequest) (*esportspb.UpdateAnnouncementResponse, error) {
	row, err := s.messageAdminUC.UpdateAnnouncement(ctx, in.GetAnnouncementId(), in.Title, in.Content)
	if err != nil {
		return nil, err
	}
	return &esportspb.UpdateAnnouncementResponse{Announcement: announcementPB(row)}, nil
}

func (s *MessageAdminService) PublishAnnouncement(ctx context.Context, in *esportspb.PublishAnnouncementRequest) (*esportspb.PublishAnnouncementResponse, error) {
	row, err := s.messageAdminUC.PublishAnnouncement(ctx, in.GetAnnouncementId())
	if err != nil {
		return nil, err
	}
	return &esportspb.PublishAnnouncementResponse{Announcement: announcementPB(row)}, nil
}

func (s *MessageAdminService) OfflineAnnouncement(ctx context.Context, in *esportspb.OfflineAnnouncementRequest) (*esportspb.OfflineAnnouncementResponse, error) {
	row, err := s.messageAdminUC.OfflineAnnouncement(ctx, in.GetAnnouncementId())
	if err != nil {
		return nil, err
	}
	return &esportspb.OfflineAnnouncementResponse{Announcement: announcementPB(row)}, nil
}

func (s *MessageAdminService) ListAnnouncements(ctx context.Context, in *esportspb.ListAnnouncementsRequest) (*esportspb.ListAnnouncementsResponse, error) {
	var kind, status *string
	if in.Kind != nil {
		folded := kindFromPB(*in.Kind)
		kind = &folded
	}
	if in.Status != nil {
		folded := statusFromPB(*in.Status)
		status = &folded
	}
	rows, err := s.messageAdminUC.ListAnnouncements(ctx, kind, status)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListAnnouncementsResponse{
		Announcements: utils.SliceMap(rows, announcementPB),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func announcementPB(a *message.Announcement) *esportspb.Announcement {
	return &esportspb.Announcement{
		Id:          utils.StringW(a.ID),
		Kind:        kindPB(a.Kind),
		Scope:       scopePB(a.Scope),
		Title:       utils.StringW(a.Title),
		Content:     utils.StringW(a.Content),
		Status:      statusPB(a.Status),
		PublishedAt: utils.ToTimestamp(a.PublishedAt),
		OfflineAt:   utils.ToTimestamp(a.OfflineAt),
		PublishedBy: utils.Wrap(a.PublishedBy, utils.StringW),
		CreatedAt:   utils.ToTimestamp(&a.CreatedAt),
		UpdatedAt:   utils.ToTimestamp(&a.UpdatedAt),
	}
}

func kindPB(v string) esportspb.AnnouncementKind {
	switch v {
	case "banner":
		return esportspb.AnnouncementKind_ANNOUNCEMENT_KIND_BANNER
	case "message":
		return esportspb.AnnouncementKind_ANNOUNCEMENT_KIND_MESSAGE
	default:
		return esportspb.AnnouncementKind_ANNOUNCEMENT_KIND_UNSPECIFIED
	}
}

func kindFromPB(v esportspb.AnnouncementKind) string {
	switch v {
	case esportspb.AnnouncementKind_ANNOUNCEMENT_KIND_BANNER:
		return "banner"
	case esportspb.AnnouncementKind_ANNOUNCEMENT_KIND_MESSAGE:
		return "message"
	default:
		return ""
	}
}

func scopePB(v string) esportspb.AnnouncementScope {
	switch v {
	case "all":
		return esportspb.AnnouncementScope_ANNOUNCEMENT_SCOPE_ALL
	case "membership":
		return esportspb.AnnouncementScope_ANNOUNCEMENT_SCOPE_MEMBERSHIP
	default:
		return esportspb.AnnouncementScope_ANNOUNCEMENT_SCOPE_UNSPECIFIED
	}
}

func scopeFromPB(v esportspb.AnnouncementScope) string {
	switch v {
	case esportspb.AnnouncementScope_ANNOUNCEMENT_SCOPE_ALL:
		return "all"
	case esportspb.AnnouncementScope_ANNOUNCEMENT_SCOPE_MEMBERSHIP:
		return "membership"
	default:
		return ""
	}
}

func statusPB(v string) esportspb.AnnouncementStatus {
	switch v {
	case "draft":
		return esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_DRAFT
	case "published":
		return esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_PUBLISHED
	case "offline":
		return esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_OFFLINE
	default:
		return esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_UNSPECIFIED
	}
}

func statusFromPB(v esportspb.AnnouncementStatus) string {
	switch v {
	case esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_DRAFT:
		return "draft"
	case esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_PUBLISHED:
		return "published"
	case esportspb.AnnouncementStatus_ANNOUNCEMENT_STATUS_OFFLINE:
		return "offline"
	default:
		return ""
	}
}
