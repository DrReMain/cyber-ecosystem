package feedbackadmin

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/feedback"
)

// Struct --------------------------------------------------------------------------------------------------------------

type FeedbackAdminService struct {
	esportspb.UnimplementedFeedbackAdminServiceServer

	log             *slog.Logger
	feedbackAdminUC *feedback.FeedbackAdminUC
}

func NewFeedbackAdminService(logger *slog.Logger, feedbackAdminUC *feedback.FeedbackAdminUC) *FeedbackAdminService {
	return &FeedbackAdminService{
		log:             logger.With("module", "module/feedbackadmin_service"),
		feedbackAdminUC: feedbackAdminUC,
	}
}

func (s *FeedbackAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterFeedbackAdminServiceServer(srv, s)
}

func (s *FeedbackAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterFeedbackAdminServiceHTTPServer(srv, s)
}

func (s *FeedbackAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterFeedbackAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *FeedbackAdminService) ListFeedbacks(ctx context.Context, in *esportspb.ListFeedbacksRequest) (*esportspb.ListFeedbacksResponse, error) {
	var status *string
	if in.Status != nil {
		folded := feedback.StatusFromPB(*in.Status)
		status = &folded
	}
	rows, page, err := s.feedbackAdminUC.List(ctx, in.GetPage(), status)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListFeedbacksResponse{
		Page:    page,
		Tickets: utils.SliceMap(rows, ticketPB),
	}, nil
}

func (s *FeedbackAdminService) GetFeedback(ctx context.Context, in *esportspb.GetFeedbackRequest) (*esportspb.GetFeedbackResponse, error) {
	row, thread, err := s.feedbackAdminUC.Get(ctx, in.GetFeedbackId())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetFeedbackResponse{
		Ticket:   ticketPB(row),
		Messages: utils.SliceMap(thread, feedback.FeedbackMessagePB),
	}, nil
}

func (s *FeedbackAdminService) ReplyFeedback(ctx context.Context, in *esportspb.ReplyFeedbackRequest) (*esportspb.ReplyFeedbackResponse, error) {
	row, err := s.feedbackAdminUC.Reply(ctx, in.GetFeedbackId(), in.GetContent())
	if err != nil {
		return nil, err
	}
	return &esportspb.ReplyFeedbackResponse{Ticket: ticketPB(&feedback.AdminTicketRow{Feedback: *row})}, nil
}

func (s *FeedbackAdminService) CloseFeedback(ctx context.Context, in *esportspb.CloseFeedbackRequest) (*esportspb.CloseFeedbackResponse, error) {
	row, err := s.feedbackAdminUC.Close(ctx, in.GetFeedbackId())
	if err != nil {
		return nil, err
	}
	return &esportspb.CloseFeedbackResponse{Ticket: ticketPB(&feedback.AdminTicketRow{Feedback: *row})}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func ticketPB(row *feedback.AdminTicketRow) *esportspb.FeedbackTicket {
	return &esportspb.FeedbackTicket{
		Id:        utils.StringW(row.ID),
		UserId:    utils.StringW(row.UserID),
		Nickname:  utils.Wrap(row.Nickname, utils.StringW),
		Category:  feedback.CategoryPB(row.Category),
		Content:   utils.StringW(row.Content),
		Contact:   utils.Wrap(row.Contact, utils.StringW),
		Status:    feedback.StatusPB(row.Status),
		CreatedAt: utils.ToTimestamp(&row.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&row.UpdatedAt),
	}
}
