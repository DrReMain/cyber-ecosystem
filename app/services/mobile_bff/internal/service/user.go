package service

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	pb "cyber-ecosystem/gen/go/cyber/mobile/v1"
)

// UserService implements MobileUserService for all three transports.
//
// This is TEMPORARY scaffolding for verifying error serialization across the
// grpc/http/connect protocols. Methods do not implement real logic yet: invalid
// requests are rejected by the validator middleware (GENERAL_ERROR_VALIDATION_
// FAILED, 400) before reaching the handler; valid requests fall through to these
// stubs, which return the mobile-owned ErrorReason business error directly. Both
// error kinds are exercised so their 3-protocol wire form can be observed. Real
// biz/data wiring comes later (Phase A), at which point these stub bodies are
// replaced.
type UserService struct {
	pb.UnimplementedMobileUserServiceServer

	log *slog.Logger
}

func NewUserService(logger *slog.Logger) *UserService {
	return &UserService{
		log: logger.With("module", "service/user"),
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

// --- stubs: return the ErrorReason business error (temporary verification code).
// Only one business error is defined for now (STATUS_INVALID_TRANSITION); every
// valid request returns it so the business-error path is exercised on all RPCs. ---

func (s *UserService) CreateUser(_ context.Context, _ *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) UpdateUser(_ context.Context, _ *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) DeleteUser(_ context.Context, _ *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) GetUser(_ context.Context, _ *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) ListUsers(_ context.Context, _ *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) UpdateUserStatus(_ context.Context, _ *pb.UpdateUserStatusRequest) (*pb.UpdateUserStatusResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}

func (s *UserService) SortUser(_ context.Context, _ *pb.SortUserRequest) (*pb.SortUserResponse, error) {
	return nil, pb.ErrorErrorReasonStatusInvalidTransition("not implemented")
}
