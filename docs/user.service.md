```go
package service

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/gen/common"
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
	if _, err := s.uc.CreateUser(ctx, &biz.User{
		Username: in.Username,
		Email:    in.Email,
		Age:      in.Age,
		Password: in.Password,
	}); err != nil {
		return nil, err
	}
	return &v1.CreateUserResponse{}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	u := &biz.User{
		ID:       in.Id,
		Username: in.GetUsername(),
		Email:    in.GetEmail(),
		Age:      in.GetAge(),
	}
	if _, err := s.uc.UpdateUser(ctx, u, in.UpdateMask.GetPaths()); err != nil {
		return nil, err
	}
	return &v1.UpdateUserResponse{}, nil
}

func (s *UserService) DeleteBatchUser(ctx context.Context, in *v1.DeleteBatchUserRequest) (*v1.DeleteBatchUserResponse, error) {
	if err := s.uc.DeleteBatchUser(ctx, in.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteBatchUserResponse{}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, in *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	if err := s.uc.DeleteUser(ctx, in.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteUserResponse{}, nil
}

func (s *UserService) GetUser(ctx context.Context, in *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	user, err := s.uc.GetUser(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetUserResponse{
		User: serializeUser(user),
	}, nil
}

func (s *UserService) QueryUser(ctx context.Context, in *v1.QueryUserRequest) (*v1.QueryUserResponse, error) {
	pageNo := in.Page.GetPageNo()
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := in.Page.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10
	}

	users, total, err := s.uc.QueryUser(ctx, &biz.UserQueryOption{
		PageNo:   pageNo,
		PageSize: pageSize,
		Email:    in.GetEmail(),
		Fields:   in.ReadMask.GetPaths(),
	})
	if err != nil {
		return nil, err
	}
	res := &v1.QueryUserResponse{
		Page: &common.PageResponse{
			PageNo:   pageNo,
			PageSize: pageSize,
			Total:    total,
			More:     total > int64(pageNo*pageSize),
		},
		List: make([]*v1.UserEntity, 0, len(users)),
	}
	for _, user := range users {
		res.List = append(res.List, serializeUser(user))
	}
	return res, nil
}

func (s *UserService) ListUser(ctx context.Context, in *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	users, err := s.uc.ListUser(ctx, &biz.UserQueryOption{
		Fields: in.ReadMask.GetPaths(),
	})
	if err != nil {
		return nil, err
	}
	res := &v1.ListUserResponse{
		List: make([]*v1.UserEntity, 0, len(users)),
	}
	for _, user := range users {
		res.List = append(res.List, serializeUser(user))
	}
	return res, nil
}

// ---

func serializeUser(u *biz.User) *v1.UserEntity {
	if u != nil {
		return &v1.UserEntity{
			Id:        u.ID,
			CreatedAt: u.CreatedAt.UnixMilli(),
			UpdatedAt: u.UpdatedAt.UnixMilli(),
			Username:  u.Username,
			Email:     u.Email,
			Age:       u.Age,
		}
	}
	return nil
}

```
