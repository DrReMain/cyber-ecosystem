package chat

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	kratosgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/grpc"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	agentpb "cyber-ecosystem/gen/go/cyber/agent/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type ChatService struct {
	agentpb.UnimplementedChatServiceServer

	log    *slog.Logger
	chatUC *ChatUC
}

func NewChatService(logger *slog.Logger, chatUC *ChatUC) *ChatService {
	return &ChatService{
		log:    logger.With("module", "module/chat_service"),
		chatUC: chatUC,
	}
}

func (s *ChatService) RegisterGRPC(srv *kratosgrpc.Server) {
	agentpb.RegisterChatServiceServer(srv, s)
}

func (s *ChatService) RegisterHTTP(srv *kratoshttp.Server) {
	agentpb.RegisterChatServiceHTTPServer(srv, s)
}

func (s *ChatService) RegisterConnect(srv *connecttransport.Server) {
	agentpb.RegisterChatServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *ChatService) DeleteSession(ctx context.Context, in *agentpb.DeleteSessionRequest) (*agentpb.DeleteSessionResponse, error) {
	if err := s.chatUC.DeleteSession(ctx, in.GetSessionId()); err != nil {
		return nil, err
	}
	return &agentpb.DeleteSessionResponse{}, nil
}

func (s *ChatService) ListModels(ctx context.Context, in *agentpb.ListModelsRequest) (*agentpb.ListModelsResponse, error) {
	models, err := s.chatUC.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	return &agentpb.ListModelsResponse{Models: models}, nil
}

func (s *ChatService) ListSessions(ctx context.Context, in *agentpb.ListSessionsRequest) (*agentpb.ListSessionsResponse, error) {
	out, err := s.chatUC.ListSessions(ctx, in.GetPage())
	if err != nil {
		return nil, err
	}
	return &agentpb.ListSessionsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, mapSessionPB),
	}, nil
}

func (s *ChatService) GetSessionMessages(ctx context.Context, in *agentpb.GetSessionMessagesRequest) (*agentpb.GetSessionMessagesResponse, error) {
	out, err := s.chatUC.ListMessages(ctx, in.GetSessionId(), in.GetPage())
	if err != nil {
		return nil, err
	}
	return &agentpb.GetSessionMessagesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, mapSessionMessagePB),
	}, nil
}

func (s *ChatService) UpdateSession(ctx context.Context, in *agentpb.UpdateSessionRequest) (*agentpb.UpdateSessionResponse, error) {
	if err := s.chatUC.RenameSession(ctx, in.GetSessionId(), in.GetTitle()); err != nil {
		return nil, err
	}
	return &agentpb.UpdateSessionResponse{}, nil
}

func (s *ChatService) Chat(in *agentpb.ChatRequest, stream grpc.ServerStreamingServer[agentpb.ChatResponse]) error {
	// Streaming requests bypass the unary protovalidate middleware, so the
	// ChatRequest constraints are enforced here (see the message comment).
	if err := validateChatRequest(in); err != nil {
		return err
	}
	ctx := stream.Context()
	prep, err := s.chatUC.CreateOrLoad(ctx, in.GetSessionId(), in.GetContent())
	if err != nil {
		return err
	}
	if prep.IsNew {
		// Echo the id before any delta; a failed send means the reader is gone.
		if serr := stream.Send(&agentpb.ChatResponse{SessionId: prep.ID}); serr != nil {
			return nil
		}
	}
	s.log.Info("Chat", "model", in.GetModel(), "session", prep.ID, "new", prep.IsNew)
	// Clone so the appended user turn never aliases prep.History's backing array.
	turns := append(slices.Clone(prep.History), &Turn{Role: "user", Content: in.GetContent()})
	var content, reasoning strings.Builder
	deltas := 0
	finish, err := s.chatUC.Stream(ctx, in.GetModel(), turns, func(d *Delta) error {
		content.WriteString(d.Content)
		reasoning.WriteString(d.Reasoning)
		deltas++
		return stream.Send(&agentpb.ChatResponse{ContentDelta: d.Content, ReasoningDelta: d.Reasoning})
	})
	if err != nil {
		if ctx.Err() != nil {
			// Hang-up after visible content still deserves persistence. The
			// stream ctx is dead: keep its values (tenant, subject), drop the
			// cancellation, bound the write.
			if deltas > 0 {
				pctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
				defer cancel()
				if perr := s.chatUC.AppendExchange(pctx, &ExchangeIn{
					Session:            prep,
					Model:              in.GetModel(),
					UserContent:        in.GetContent(),
					AssistantContent:   content.String(),
					AssistantReasoning: reasoning.String(),
					Finish:             "aborted",
				}); perr != nil {
					s.log.Error("persist aborted exchange", "session", prep.ID, "err", perr)
				}
			}
			return nil
		}
		return err
	}
	// Persist before the final frame: a persist failure surfaces as a Connect
	// error, the streamed text stays visible, and a retry is duplicate-free.
	if perr := s.chatUC.AppendExchange(ctx, &ExchangeIn{
		Session:            prep,
		Model:              in.GetModel(),
		UserContent:        in.GetContent(),
		AssistantContent:   content.String(),
		AssistantReasoning: reasoning.String(),
		Finish:             finish,
	}); perr != nil {
		return perr
	}
	return stream.Send(&agentpb.ChatResponse{FinishReason: finish})
}

// Private -------------------------------------------------------------------------------------------------------------

func validateChatRequest(in *agentpb.ChatRequest) error {
	if m := in.GetModel(); m == "" || len(m) > 128 {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("model: required, 1..128 chars"))
	}
	if len(in.GetSessionId()) > 20 {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("session_id: max 20 chars"))
	}
	if c := in.GetContent(); c == "" || len(c) > 32768 {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("content: required, 1..32768 chars"))
	}
	return nil
}

func mapSessionPB(sess *ChatSession) *agentpb.ChatSession {
	return &agentpb.ChatSession{
		Id:        utils.StringW(sess.ID),
		CreatedAt: utils.ToTimestamp(&sess.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&sess.UpdatedAt),
		Title:     utils.StringW(sess.Title),
	}
}

func mapSessionMessagePB(m *SessionMessage) *agentpb.SessionMessage {
	return &agentpb.SessionMessage{
		Id:        utils.StringW(m.ID),
		CreatedAt: utils.ToTimestamp(&m.CreatedAt),
		Role:      utils.StringW(m.Role),
		Content:   utils.StringW(m.Content),
		Reasoning: utils.StringW(m.Reasoning),
		Model:     utils.StringW(m.Model),
	}
}
