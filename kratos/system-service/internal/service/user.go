package service

import (
	"context"

	v1 "github.com/DrReMain/cyber-ecosystem/kratos/system-service/gen/v1"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type UserService struct {
	v1.UnimplementedUserServiceServer

	log *log.Helper
	uc  *biz.UserUC
}

func (s *UserService) RegisterGRPC(srv *grpc.Server) {
	v1.RegisterUserServiceServer(srv, s)
}
func (s *UserService) RegisterHTTP(srv *http.Server) {
	v1.RegisterUserServiceHTTPServer(srv, s)
}

func NewUserService(uc *biz.UserUC, logger log.Logger) *UserService {
	return &UserService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *UserService) CreateUser(ctx context.Context, in *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	return &v1.CreateUserResponse{}, nil
}
