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

type UserAttributeService struct {
	singletonV1.UnimplementedUserAttributeServiceServer
	log             *log.Helper
	userAttributeUC *biz.UserAttributeUC
}

func NewUserAttributeService(logger log.Logger, userAttributeUC *biz.UserAttributeUC) *UserAttributeService {
	return &UserAttributeService{
		log:             log.NewHelper(log.With(logger, "module", "service/user_attribute")),
		userAttributeUC: userAttributeUC,
	}
}
func (s *UserAttributeService) RegisterGRPC(srv *grpc.Server) {
	singletonV1.RegisterUserAttributeServiceServer(srv, s)
}
func (s *UserAttributeService) RegisterHTTP(srv *http.Server) {
	singletonV1.RegisterUserAttributeServiceHTTPServer(srv, s)
}
func (s *UserAttributeService) RegisterConnect(srv *connect.Server) {
	srv.Register(singletonV1connect.NewUserAttributeServiceHandler(s, srv.HandlerOptions()...))
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *UserAttributeService) SetUserAttribute(ctx context.Context, in *singletonV1.SetUserAttributeRequest) (*singletonV1.SetUserAttributeResponse, error) {
	attr := &biz.UserAttribute{
		UserID: &in.UserId,
		Key:    &in.Key,
		Value:  &in.Value,
	}
	if err := s.userAttributeUC.Set(ctx, attr); err != nil {
		return nil, err
	}
	return &singletonV1.SetUserAttributeResponse{}, nil
}

func (s *UserAttributeService) RemoveUserAttribute(ctx context.Context, in *singletonV1.RemoveUserAttributeRequest) (*singletonV1.RemoveUserAttributeResponse, error) {
	if err := s.userAttributeUC.Remove(ctx, in.UserId, in.Key); err != nil {
		return nil, err
	}
	return &singletonV1.RemoveUserAttributeResponse{}, nil
}

func (s *UserAttributeService) QueryUserAttributes(ctx context.Context, in *singletonV1.QueryUserAttributesRequest) (*singletonV1.QueryUserAttributesResponse, error) {
	attrs, err := s.userAttributeUC.Query(ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &singletonV1.QueryUserAttributesResponse{
		List: utils.SliceMap(attrs, s.attrToProto),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (s *UserAttributeService) attrToProto(e *biz.UserAttribute) *singletonV1.GetUserAttributeResponse {
	return &singletonV1.GetUserAttributeResponse{
		Id:        *e.ID,
		CreatedAt: utils.ToTimestamp(e.CreatedAt),
		UpdatedAt: utils.ToTimestamp(e.UpdatedAt),
		UserId:    *e.UserID,
		Key:       *e.Key,
		Value:     *e.Value,
	}
}
