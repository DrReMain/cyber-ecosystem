package user

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type UserService struct {
	systempb.UnimplementedUserServiceServer

	log    *slog.Logger
	userUC *UserUC
}

func NewUserService(logger *slog.Logger, userUC *UserUC) *UserService {
	return &UserService{
		log:    logger.With("module", "module/user_service"),
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
	created, err := s.userUC.Create(ctx, in.Email, in.Password, in.DeptId)
	if err != nil {
		return nil, err
	}
	return &systempb.CreateUserResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *systempb.UpdateUserRequest) (*systempb.UpdateUserResponse, error) {
	if _, err := s.userUC.Update(ctx, in.FieldsMask, &User{
		ID:     in.Id,
		Email:  in.Email,
		DeptID: in.DeptId,
	}, in.Password); err != nil {
		return nil, err
	}
	return &systempb.UpdateUserResponse{}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, in *systempb.DeleteUserRequest) (*systempb.DeleteUserResponse, error) {
	if _, err := s.userUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeleteUserResponse{}, nil
}

func (s *UserService) ListUsers(ctx context.Context, in *systempb.ListUsersRequest) (*systempb.ListUsersResponse, error) {
	out, err := s.userUC.List(ctx, &UserListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Email:       in.Email,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListUsersResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoUser),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *systempb.GetUserRequest) (*systempb.GetUserResponse, error) {
	a, err := s.userUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetUserResponse{
		User: toProtoUser(a),
	}, nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, in *systempb.GetCurrentUserRequest) (*systempb.GetCurrentUserResponse, error) {
	a, err := s.userUC.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &systempb.GetCurrentUserResponse{
		User: toProtoUser(a),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoUser(a *User) *systempb.User {
	return &systempb.User{
		Id:        utils.StringW(a.ID),
		CreatedAt: utils.ToTimestamp(&a.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&a.UpdatedAt),
		Email:     utils.Wrap(a.Email, utils.StringW),
		DeptId:    utils.Wrap(a.DeptID, utils.StringW),
	}
}
