package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	singletonV1connect "cyber-ecosystem/apps/singleton/gen/go/v1/singletonV1connect"
	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
)

type UserService struct {
	singletonV1.UnimplementedUserServiceServer
	log    *log.Helper
	userUC *biz.UserUC
}

func NewUserService(logger log.Logger, userUC *biz.UserUC) *UserService {
	return &UserService{
		log:    log.NewHelper(log.With(logger, "module", "service/user")),
		userUC: userUC,
	}
}
func (s *UserService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterUserServiceServer(srv, s)
}
func (s *UserService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterUserServiceHTTPServer(srv, s)
}
func (s *UserService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewUserServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *UserService) CreateUser(ctx context.Context, in *singletonV1.CreateUserRequest) (*singletonV1.CreateUserResponse, error) {
	user := &biz.User{
		PasswordPlain: in.Password,
		Email:         in.Email,
	}
	if err := s.userUC.Create(ctx, user); err != nil {
		return nil, err
	}
	return &singletonV1.CreateUserResponse{}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *singletonV1.UpdateUserRequest) (*singletonV1.UpdateUserResponse, error) {
	user := &biz.User{
		ID:            &in.Id,
		PasswordPlain: in.Password,
	}
	if err := s.userUC.Update(ctx, in.FieldsMask, user); err != nil {
		return nil, err
	}
	return &singletonV1.UpdateUserResponse{}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, in *singletonV1.DeleteUserRequest) (*singletonV1.DeleteUserResponse, error) {
	if err := s.userUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &singletonV1.DeleteUserResponse{}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *singletonV1.GetUserRequest) (*singletonV1.GetUserResponse, error) {
	user, err := s.userUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.userToProto(user), nil
}

func (s *UserService) QueryUser(ctx context.Context, in *singletonV1.QueryUserRequest) (*singletonV1.QueryUserResponse, error) {
	out, err := s.userUC.Query(ctx, &biz.UserQueryIn{
		PageRequest: utils.EnsurePageRequest(in.Page),
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
		ID:          in.Id,
		Email:       in.Email,
	})
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryUserResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.userToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *UserService) userToProto(e *biz.User) *singletonV1.GetUserResponse {
	return &singletonV1.GetUserResponse{
		Id:        *e.ID,
		CreatedAt: utils.ToTimestamp(e.CreatedAt),
		UpdatedAt: utils.ToTimestamp(e.UpdatedAt),
		Email:     utils.Wrap(e.Email, utils.StringW),
	}
}
