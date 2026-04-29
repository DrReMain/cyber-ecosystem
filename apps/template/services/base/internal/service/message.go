package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	templateV1 "cyber-ecosystem/apps/template/gen/go/v1"
	templateV1connect "cyber-ecosystem/apps/template/gen/go/v1/templateV1connect"
	"cyber-ecosystem/apps/template/services/base/internal/biz"
)

// region[rgba(236,64,122,0.15)] 🩷 Struct -------------------------------------------------------------------------------

type MessageService struct {
	templateV1.UnimplementedMessageServiceServer
	log       *log.Helper
	messageUC *biz.MessageUC
}

func NewMessageService(logger log.Logger, messageUC *biz.MessageUC) *MessageService {
	return &MessageService{
		log:       log.NewHelper(log.With(logger, "module", "service/message")),
		messageUC: messageUC,
	}
}
func (s *MessageService) RegisterGRPC(srv *grpc.Server) {
	templateV1.RegisterMessageServiceServer(srv, s)
}
func (s *MessageService) RegisterHTTP(srv *http.Server) {
	templateV1.RegisterMessageServiceHTTPServer(srv, s)
}
func (s *MessageService) RegisterConnect(srv *connect.Server) {
	srv.Register(templateV1connect.NewMessageServiceHandler(s, srv.HandlerOptions()...))
}

// region[rgba(255,167,38,0.15)] 🟠 Handler -----------------------------------------------------------------------------

func (s *MessageService) CreateMessage(ctx context.Context, in *templateV1.CreateMessageRequest) (*templateV1.CreateMessageResponse, error) {
	m := &biz.Message{
		Title:   in.Title,
		Content: in.Content,
	}
	created, err := s.messageUC.Create(ctx, m)
	if err != nil {
		return nil, err
	}
	return &templateV1.CreateMessageResponse{
		Id: utils.Wrap(created.ID, utils.StringW),
	}, nil
}

func (s *MessageService) UpdateMessage(ctx context.Context, in *templateV1.UpdateMessageRequest) (*templateV1.UpdateMessageResponse, error) {
	m := &biz.Message{
		ID:      in.Id,
		Title:   in.Title,
		Content: in.Content,
	}
	updated, err := s.messageUC.Update(ctx, in.FieldsMask, m)
	if err != nil {
		return nil, err
	}
	return &templateV1.UpdateMessageResponse{
		Id: utils.Wrap(updated.ID, utils.StringW),
	}, nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, in *templateV1.DeleteMessageRequest) (*templateV1.DeleteMessageResponse, error) {
	deletedID, err := s.messageUC.Delete(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return &templateV1.DeleteMessageResponse{
		Id: utils.StringW(deletedID),
	}, nil
}

func (s *MessageService) GetMessage(ctx context.Context, in *templateV1.GetMessageRequest) (*templateV1.GetMessageResponse, error) {
	m, err := s.messageUC.Get(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return s.messageToProto(m), nil
}

func (s *MessageService) QueryMessage(ctx context.Context, in *templateV1.QueryMessageRequest) (*templateV1.QueryMessageResponse, error) {
	out, err := s.messageUC.Query(ctx, &biz.MessageQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
		Title:       in.Title,
		Status:      in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &templateV1.QueryMessageResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.messageToProto),
	}, nil
}

func (s *MessageService) SortMessage(ctx context.Context, in *templateV1.SortMessageRequest) (*templateV1.SortMessageResponse, error) {
	sorted, err := s.messageUC.Sort(ctx, *in.Id, in.PrevId, in.NextId)
	if err != nil {
		return nil, err
	}
	return &templateV1.SortMessageResponse{
		Id: utils.Wrap(sorted.ID, utils.StringW),
	}, nil
}

func (s *MessageService) UpdateMessageStatus(ctx context.Context, in *templateV1.UpdateMessageStatusRequest) (*templateV1.UpdateMessageStatusResponse, error) {
	_, err := s.messageUC.UpdateStatus(ctx, *in.Id, *in.Status)
	if err != nil {
		return nil, err
	}
	return &templateV1.UpdateMessageStatusResponse{
		Id: utils.StringW(*in.Id),
	}, nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func (s *MessageService) messageToProto(m *biz.Message) *templateV1.GetMessageResponse {
	return &templateV1.GetMessageResponse{
		Id:        utils.Wrap(m.ID, utils.StringW),
		CreatedAt: utils.ToTimestamp(m.CreatedAt),
		UpdatedAt: utils.ToTimestamp(m.UpdatedAt),
		Title:     utils.Wrap(m.Title, utils.StringW),
		Content:   utils.Wrap(m.Content, utils.StringW),
		Status:    utils.Wrap(m.Status, utils.StringW),
	}
}
