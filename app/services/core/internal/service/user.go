package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	pb "cyber-ecosystem/gen/go/cyber/core/v1"

	"cyber-ecosystem/app/services/core/internal/biz"
)

// Struct ----------------------------------------------------------------------------------------------------------------

type UserService struct {
	pb.UnimplementedUserServiceServer

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
	pb.RegisterUserServiceServer(srv, s)
}

func (s *UserService) RegisterHTTP(srv *http.Server) {
	pb.RegisterUserServiceHTTPServer(srv, s)
}

func (s *UserService) RegisterConnect(srv *connecttransport.Server) {
	pb.RegisterUserServiceConnectServer(srv, s)
}

// Handler ---------------------------------------------------------------------------------------------------------------

func (s *UserService) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	u := &biz.User{
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Password: in.Password,
		Avatar:   in.Avatar,
	}
	created, err := s.userUC.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Id: utils.StringW(created.ID)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	u := &biz.User{
		ID:       *in.Id,
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Password: in.Password,
		Avatar:   in.Avatar,
	}
	if _, err := s.userUC.Update(ctx, in.FieldsMask, u); err != nil {
		return nil, err
	}
	return &pb.UpdateUserResponse{}, nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, in *pb.UpdateUserStatusRequest) (*pb.UpdateUserStatusResponse, error) {
	if _, err := s.userUC.UpdateStatus(ctx, *in.Id, *in.Status); err != nil {
		return nil, err
	}
	return &pb.UpdateUserStatusResponse{}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if _, err := s.userUC.Delete(ctx, *in.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}

func (s *UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	out, err := s.userUC.List(ctx, &biz.UserListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     helper.ParseOrderBy(in.OrderBy),
		Phone:       in.Phone,
		Status:      in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListUsersResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, func(u *biz.User) *pb.GetUserResponse {
			return &pb.GetUserResponse{User: toUser(u)}
		}),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	u, err := s.userUC.Get(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{User: toUser(u)}, nil
}

func (s *UserService) SortUser(ctx context.Context, in *pb.SortUserRequest) (*pb.SortUserResponse, error) {
	if _, err := s.userUC.Sort(ctx, *in.Id, in.PrevId, in.NextId); err != nil {
		return nil, err
	}
	return &pb.SortUserResponse{}, nil
}

// Private ---------------------------------------------------------------------------------------------------------------

func toUser(u *biz.User) *pb.User {
	return &pb.User{
		Id:        utils.StringW(u.ID),
		CreatedAt: utils.ToTimestamp(&u.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&u.UpdatedAt),
		Nickname:  utils.Wrap(u.Nickname, utils.StringW),
		Avatar:    utils.Wrap(u.Avatar, utils.StringW),
		Phone:     utils.Wrap(u.Phone, utils.StringW),
		Status:    utils.Wrap(u.Status, utils.StringW),
	}
}
