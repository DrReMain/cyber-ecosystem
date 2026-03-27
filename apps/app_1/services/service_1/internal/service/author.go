package service

import (
	"context"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/kratos/utils"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	app1ConnectV1 "cyber-ecosystem/apps/app_1/gen/go/v1/app1V1connect"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
)

type AuthorService struct {
	app1V1.UnimplementedAuthorServiceServer

	log      *log.Helper
	authorUC *biz.AuthorUC
}

func NewAuthorService(logger log.Logger, authorUC *biz.AuthorUC) *AuthorService {
	return &AuthorService{
		log:      log.NewHelper(log.With(logger, "module", "service/author")),
		authorUC: authorUC,
	}
}
func (s *AuthorService) RegisterGRPC(srv *grpc.Server) {
	app1V1.RegisterAuthorServiceServer(srv, s)
}
func (s *AuthorService) RegisterHTTP(srv *http.Server) {
	app1V1.RegisterAuthorServiceHTTPServer(srv, s)
}
func (s *AuthorService) RegisterConnect(srv *connect.Server) {
	srv.Register(app1ConnectV1.NewAuthorServiceHandler(s, srv.HandlerOptions()...))
}

// ---------------------------------------------------------------------------------------------------------------------

func (s *AuthorService) CreateAuthor(ctx context.Context, in *app1V1.CreateAuthorRequest) (*app1V1.CreateAuthorResponse, error) {
	entity := &biz.AuthorEntity{
		Name: in.Name,
	}
	if err := s.authorUC.CreateAuthor(ctx, entity); err != nil {
		return nil, err
	}
	return &app1V1.CreateAuthorResponse{}, nil
}

func (s *AuthorService) GetAuthor(ctx context.Context, in *app1V1.GetAuthorRequest) (*app1V1.GetAuthorResponse, error) {
	entity, err := s.authorUC.GetAuthor(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &app1V1.GetAuthorResponse{
		Id:        entity.ID,
		CreatedAt: utils.GetPPbTimeFromPTime(entity.CreatedAt),
		UpdatedAt: utils.GetPPbTimeFromPTime(entity.UpdatedAt),
		Name:      utils.ToPtrWrapper(entity.Name, wrapperspb.String),
	}, nil
}
