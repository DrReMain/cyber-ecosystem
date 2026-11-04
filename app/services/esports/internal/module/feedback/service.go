package feedback

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

type FeedbackService struct {
	esportspb.UnimplementedFeedbackServiceServer

	log        *slog.Logger
	feedbackUC *FeedbackUC
}

func NewFeedbackService(logger *slog.Logger, feedbackUC *FeedbackUC) *FeedbackService {
	return &FeedbackService{
		log:        logger.With("module", "module/feedback_service"),
		feedbackUC: feedbackUC,
	}
}

func (s *FeedbackService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterFeedbackServiceServer(srv, s)
}

func (s *FeedbackService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterFeedbackServiceHTTPServer(srv, s)
}

func (s *FeedbackService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterFeedbackServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *FeedbackService) SubmitFeedback(ctx context.Context, in *esportspb.SubmitFeedbackRequest) (*esportspb.SubmitFeedbackResponse, error) {
	row, err := s.feedbackUC.Submit(ctx, CategoryFromPB(in.GetCategory()), in.GetContent(), in.Contact)
	if err != nil {
		return nil, err
	}
	return &esportspb.SubmitFeedbackResponse{Feedback: feedbackPB(row)}, nil
}

func (s *FeedbackService) ListMyFeedbacks(ctx context.Context, in *esportspb.ListMyFeedbacksRequest) (*esportspb.ListMyFeedbacksResponse, error) {
	out, err := s.feedbackUC.ListMine(ctx, in.GetPage())
	if err != nil {
		return nil, err
	}
	return &esportspb.ListMyFeedbacksResponse{
		Page:      out.PageResponse,
		Feedbacks: utils.SliceMap(out.Rows, feedbackPB),
	}, nil
}

func (s *FeedbackService) GetMyFeedback(ctx context.Context, in *esportspb.GetMyFeedbackRequest) (*esportspb.GetMyFeedbackResponse, error) {
	row, thread, err := s.feedbackUC.GetMine(ctx, in.GetFeedbackId())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetMyFeedbackResponse{
		Feedback: feedbackPB(row),
		Messages: utils.SliceMap(thread, FeedbackMessagePB),
	}, nil
}

func (s *FeedbackService) FollowUpFeedback(ctx context.Context, in *esportspb.FollowUpFeedbackRequest) (*esportspb.FollowUpFeedbackResponse, error) {
	row, err := s.feedbackUC.FollowUp(ctx, in.GetFeedbackId(), in.GetContent())
	if err != nil {
		return nil, err
	}
	return &esportspb.FollowUpFeedbackResponse{Feedback: feedbackPB(row)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func feedbackPB(f *Feedback) *esportspb.Feedback {
	return &esportspb.Feedback{
		Id:        utils.StringW(f.ID),
		Category:  CategoryPB(f.Category),
		Content:   utils.StringW(f.Content),
		Contact:   utils.Wrap(f.Contact, utils.StringW),
		Status:    StatusPB(f.Status),
		CreatedAt: utils.ToTimestamp(&f.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&f.UpdatedAt),
	}
}

func FeedbackMessagePB(m *FeedbackMessage) *esportspb.FeedbackMessage {
	return &esportspb.FeedbackMessage{
		Id:        utils.StringW(m.ID),
		Sender:    SenderPB(m.Sender),
		Content:   utils.StringW(m.Content),
		CreatedAt: utils.ToTimestamp(&m.CreatedAt),
	}
}

func CategoryPB(v string) esportspb.FeedbackCategory {
	switch v {
	case "feature_bug":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_FEATURE_BUG
	case "assessment":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_ASSESSMENT
	case "report":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_REPORT
	case "payment":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_PAYMENT
	case "account":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_ACCOUNT
	case "other":
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_OTHER
	default:
		return esportspb.FeedbackCategory_FEEDBACK_CATEGORY_UNSPECIFIED
	}
}

func CategoryFromPB(v esportspb.FeedbackCategory) string {
	switch v {
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_FEATURE_BUG:
		return "feature_bug"
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_ASSESSMENT:
		return "assessment"
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_REPORT:
		return "report"
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_PAYMENT:
		return "payment"
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_ACCOUNT:
		return "account"
	case esportspb.FeedbackCategory_FEEDBACK_CATEGORY_OTHER:
		return "other"
	default:
		return ""
	}
}

func StatusPB(v string) esportspb.FeedbackStatus {
	switch v {
	case statusPending:
		return esportspb.FeedbackStatus_FEEDBACK_STATUS_PENDING
	case statusReplied:
		return esportspb.FeedbackStatus_FEEDBACK_STATUS_REPLIED
	case statusClosed:
		return esportspb.FeedbackStatus_FEEDBACK_STATUS_CLOSED
	default:
		return esportspb.FeedbackStatus_FEEDBACK_STATUS_UNSPECIFIED
	}
}

func StatusFromPB(v esportspb.FeedbackStatus) string {
	switch v {
	case esportspb.FeedbackStatus_FEEDBACK_STATUS_PENDING:
		return statusPending
	case esportspb.FeedbackStatus_FEEDBACK_STATUS_REPLIED:
		return statusReplied
	case esportspb.FeedbackStatus_FEEDBACK_STATUS_CLOSED:
		return statusClosed
	default:
		return ""
	}
}

func SenderPB(v string) esportspb.FeedbackMessageSender {
	switch v {
	case "user":
		return esportspb.FeedbackMessageSender_FEEDBACK_MESSAGE_SENDER_USER
	case "admin":
		return esportspb.FeedbackMessageSender_FEEDBACK_MESSAGE_SENDER_ADMIN
	default:
		return esportspb.FeedbackMessageSender_FEEDBACK_MESSAGE_SENDER_UNSPECIFIED
	}
}
