package tenant

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

type TenantService struct {
	systempb.UnimplementedTenantServiceServer

	log      *slog.Logger
	tenantUC *TenantUC
}

func NewTenantService(logger *slog.Logger, tenantUC *TenantUC) *TenantService {
	return &TenantService{
		log:      logger.With("module", "module/tenant_service"),
		tenantUC: tenantUC,
	}
}

func (s *TenantService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterTenantServiceServer(srv, s)
}

func (s *TenantService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterTenantServiceHTTPServer(srv, s)
}

func (s *TenantService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterTenantServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *TenantService) UpdateTenantStatus(ctx context.Context, in *systempb.UpdateTenantStatusRequest) (*systempb.UpdateTenantStatusResponse, error) {
	if _, err := s.tenantUC.UpdateStatus(ctx, in.Id, in.GetEnabled()); err != nil {
		return nil, err
	}
	return &systempb.UpdateTenantStatusResponse{}, nil
}

func (s *TenantService) ListTenants(ctx context.Context, in *systempb.ListTenantsRequest) (*systempb.ListTenantsResponse, error) {
	out, err := s.tenantUC.List(ctx, &TenantListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Code:        in.Code,
		Name:        in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListTenantsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoTenant),
	}, nil
}

func (s *TenantService) GetTenant(ctx context.Context, in *systempb.GetTenantRequest) (*systempb.GetTenantResponse, error) {
	out, err := s.tenantUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetTenantResponse{Tenant: toProtoTenant(out)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoTenant(t *Tenant) *systempb.Tenant {
	return &systempb.Tenant{
		Id:        utils.StringW(t.ID),
		Code:      utils.StringW(t.Code),
		Name:      utils.StringW(t.Name),
		Enabled:   utils.BoolW(t.Enabled),
		CreatedAt: utils.ToTimestamp(&t.CreatedAt),
		UpdatedAt: utils.ToTimestamp(&t.UpdatedAt),
	}
}
