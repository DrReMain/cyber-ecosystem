package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/biz"
)

// Struct --------------------------------------------------------------------------------------------------------------

type UserService struct {
	systempb.UnimplementedUserServiceServer

	log    *slog.Logger
	userUC *biz.UserUC
}

func NewUserService(logger *slog.Logger, userUC *biz.UserUC) *UserService {
	return &UserService{
		log:    logger.With("module", "service/user"),
		userUC: userUC,
	}
}

func (s *UserService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterUserServiceServer(srv, s)
}

func (s *UserService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterUserServiceHTTPServer(srv, s)
}

func (s *UserService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterUserServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *UserService) CreateUser(ctx context.Context, in *systempb.CreateUserRequest) (*systempb.CreateUserResponse, error) {
	created, err := s.userUC.Create(ctx, *in.Email, *in.Password, in.DeptId)
	if err != nil {
		return nil, err
	}
	return &systempb.CreateUserResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *systempb.GetUserRequest) (*systempb.GetUserResponse, error) {
	a, err := s.userUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetUserResponse{
		User: s.userToProto(a),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *UserService) userToProto(a *biz.User) *systempb.User {
	return &systempb.User{
		Id:        utils.StringW(a.ID),
		CreatedAt: utils.ToTimestamp(&a.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&a.UpdatedAt),
		Email:     utils.StringW(a.Email),
		DeptId:    utils.Wrap(a.DeptID, utils.StringW),
	}
}
