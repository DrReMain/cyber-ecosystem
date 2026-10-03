package introspect

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// Struct --------------------------------------------------------------------------------------------------------------

type IntrospectService struct {
	systempb.UnimplementedIntrospectServiceServer

	log        *slog.Logger
	authorizer kauthz.Authorizer
}

func NewIntrospectService(logger *slog.Logger, authorizer kauthz.Authorizer) *IntrospectService {
	return &IntrospectService{
		log:        logger.With("module", "module/introspect_service"),
		authorizer: authorizer,
	}
}

func (s *IntrospectService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterIntrospectServiceServer(srv, s)
}

func (s *IntrospectService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterIntrospectServiceHTTPServer(srv, s)
}

func (s *IntrospectService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterIntrospectServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *IntrospectService) VerifySession(ctx context.Context, in *systempb.VerifySessionRequest) (*systempb.VerifySessionResponse, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, kauthz.ErrMissingSubject
	}
	email, _ := subject.Extra[shared.ExtraEmail].(string)
	return &systempb.VerifySessionResponse{
		UserId:    utils.StringW(subject.UserID),
		TenantId:  utils.StringW(subject.TenantID),
		SessionId: utils.StringW(subject.SessionID),
		Email:     utils.StringW(email),
	}, nil
}

func (s *IntrospectService) CheckOperation(ctx context.Context, in *systempb.CheckOperationRequest) (*systempb.CheckOperationResponse, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, kauthz.ErrMissingSubject
	}
	d, err := s.authorizer.Decide(ctx, subject, in.Operation)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, kauthz.ErrPermissionDenied
	}
	return &systempb.CheckOperationResponse{Allowed: d.Allowed, Reason: d.Reason}, nil
}
