package message

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type MessageService struct {
	esportspb.UnimplementedMessageServiceServer

	log       *slog.Logger
	messageUC *MessageUC
}

func NewMessageService(logger *slog.Logger, messageUC *MessageUC) *MessageService {
	return &MessageService{
		log:       logger.With("module", "module/message_service"),
		messageUC: messageUC,
	}
}

func (s *MessageService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterMessageServiceServer(srv, s)
}

func (s *MessageService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterMessageServiceHTTPServer(srv, s)
}

func (s *MessageService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterMessageServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *MessageService) ListMessages(ctx context.Context, in *esportspb.ListMessagesRequest) (*esportspb.ListMessagesResponse, error) {
	out, err := s.messageUC.ListMessages(ctx, in.GetPage())
	if err != nil {
		return nil, err
	}
	return &esportspb.ListMessagesResponse{
		Page:        out.PageResponse,
		Messages:    utils.SliceMap(out.Rows, MessagePB),
		UnreadCount: utils.ConvNum[int32](out.Unread),
	}, nil
}

func (s *MessageService) GetUnreadCount(ctx context.Context, in *esportspb.GetUnreadCountRequest) (*esportspb.GetUnreadCountResponse, error) {
	count, err := s.messageUC.CountUnread(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.GetUnreadCountResponse{Count: utils.ConvNum[int32](count)}, nil
}

func (s *MessageService) MarkMessageRead(ctx context.Context, in *esportspb.MarkMessageReadRequest) (*esportspb.MarkMessageReadResponse, error) {
	if err := s.messageUC.MarkRead(ctx, in.GetMessageId()); err != nil {
		return nil, err
	}
	return &esportspb.MarkMessageReadResponse{}, nil
}

func (s *MessageService) MarkAllMessagesRead(ctx context.Context, in *esportspb.MarkAllMessagesReadRequest) (*esportspb.MarkAllMessagesReadResponse, error) {
	if err := s.messageUC.MarkAllRead(ctx); err != nil {
		return nil, err
	}
	return &esportspb.MarkAllMessagesReadResponse{}, nil
}

func (s *MessageService) ListBanners(ctx context.Context, in *esportspb.ListBannersRequest) (*esportspb.ListBannersResponse, error) {
	rows, err := s.messageUC.ListBanners(ctx)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListBannersResponse{Banners: utils.SliceMap(rows, bannerPB)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func MessagePB(m *Message) *esportspb.Message {
	// messageadmin's transport consumes this; one mapper owning the type
	// switch is what keeps both faces agreeing on what a letter carries.
	return &esportspb.Message{
		Id:        utils.StringW(m.ID),
		Type:      messageTypePB(m.Type),
		Title:     utils.StringW(m.Title),
		Content:   utils.StringW(m.Content),
		RefId:     utils.Wrap(m.RefID, utils.StringW),
		IsRead:    utils.BoolW(m.IsRead),
		CreatedAt: utils.ToTimestamp(&m.CreatedAt),
	}
}

func bannerPB(a *Announcement) *esportspb.Banner {
	return &esportspb.Banner{
		Title:       utils.StringW(a.Title),
		Content:     utils.StringW(a.Content),
		PublishedAt: utils.ToTimestamp(a.PublishedAt),
	}
}

func messageTypePB(v string) esportspb.MessageType {
	switch v {
	case typeAnnouncement:
		return esportspb.MessageType_MESSAGE_TYPE_ANNOUNCEMENT
	case "report":
		return esportspb.MessageType_MESSAGE_TYPE_REPORT
	case "membership":
		return esportspb.MessageType_MESSAGE_TYPE_MEMBERSHIP
	case feedbackReplyKind:
		return esportspb.MessageType_MESSAGE_TYPE_FEEDBACK_REPLY
	default:
		return esportspb.MessageType_MESSAGE_TYPE_UNSPECIFIED
	}
}
