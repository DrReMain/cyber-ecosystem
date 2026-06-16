package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	pb "cyber-ecosystem/gen/go/cyber/mobile/v1"

	"cyber-ecosystem/app/services/edge_mobile/internal/biz"
)

type UserService struct {
	pb.UnimplementedMobileUserServiceServer

	log *slog.Logger
	uc  *biz.MobileUserUC
}

func NewUserService(logger *slog.Logger, uc *biz.MobileUserUC) *UserService {
	return &UserService{
		log: logger.With("module", "service/user"),
		uc:  uc,
	}
}

func (s *UserService) RegisterGRPC(srv *grpc.Server) {
	pb.RegisterMobileUserServiceServer(srv, s)
}

func (s *UserService) RegisterHTTP(srv *http.Server) {
	pb.RegisterMobileUserServiceHTTPServer(srv, s)
}

func (s *UserService) RegisterConnect(srv *connecttransport.Server) {
	pb.RegisterMobileUserServiceConnectServer(srv, s)
}

// Method --------------------------------------------------------------------------------------------------------

func (s *UserService) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	u := &biz.MobileUser{
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Password: in.Password,
		Avatar:   in.Avatar,
	}
	created, err := s.uc.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Id: utils.StringW(created.ID)}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	u, err := s.uc.Get(ctx, *in.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{User: toUser(u)}, nil
}

func (s *UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	out, err := s.uc.List(ctx, &biz.MobileUserListIn{
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
		List: utils.SliceMap(out.List, func(u *biz.MobileUser) *pb.GetUserResponse {
			return &pb.GetUserResponse{User: toUser(u)}
		}),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	u := &biz.MobileUser{
		ID:       *in.Id,
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Password: in.Password,
		Avatar:   in.Avatar,
	}
	if _, err := s.uc.Update(ctx, in.FieldsMask, u); err != nil {
		return nil, err
	}
	return &pb.UpdateUserResponse{}, nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, in *pb.UpdateUserStatusRequest) (*pb.UpdateUserStatusResponse, error) {
	if _, err := s.uc.UpdateStatus(ctx, *in.Id, *in.Status); err != nil {
		return nil, err
	}
	return &pb.UpdateUserStatusResponse{}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if _, err := s.uc.Delete(ctx, *in.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}

func (s *UserService) SortUser(ctx context.Context, in *pb.SortUserRequest) (*pb.SortUserResponse, error) {
	if _, err := s.uc.Sort(ctx, *in.Id, in.PrevId, in.NextId); err != nil {
		return nil, err
	}
	return &pb.SortUserResponse{}, nil
}

// Private --------------------------------------------------------------------------------------------------------

func toUser(u *biz.MobileUser) *pb.User {
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
