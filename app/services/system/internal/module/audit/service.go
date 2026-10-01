package audit

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
)

// Struct --------------------------------------------------------------------------------------------------------------

type AuditService struct {
	systempb.UnimplementedAuditServiceServer

	log     *slog.Logger
	auditUC *AuditUC
}

func NewAuditService(logger *slog.Logger, auditUC *AuditUC) *AuditService {
	return &AuditService{
		log:     logger.With("module", "module/audit_service"),
		auditUC: auditUC,
	}
}

func (s *AuditService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterAuditServiceServer(srv, s)
}

func (s *AuditService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterAuditServiceHTTPServer(srv, s)
}

func (s *AuditService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterAuditServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *AuditService) ListAuditLogs(ctx context.Context, in *systempb.ListAuditLogsRequest) (*systempb.ListAuditLogsResponse, error) {
	out, err := s.auditUC.List(ctx, &AuditListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Actor:       in.Actor,
		Operation:   in.Operation,
		Status:      in.Status,
		DenyReason:  in.DenyReason,
		Denied:      in.Denied,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListAuditLogsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoAuditLog),
	}, nil
}

func (s *AuditService) ExportAuditLogs(ctx context.Context, in *systempb.ExportAuditLogsRequest) (*systempb.ExportAuditLogsResponse, error) {
	f, err := s.auditUC.Export(ctx, &ExportIn{IDs: in.Ids})
	if err != nil {
		return nil, err
	}
	return &systempb.ExportAuditLogsResponse{File: file.ToProtoFile(f)}, nil
}

func (s *AuditService) ListAuditExports(ctx context.Context, in *systempb.ListAuditExportsRequest) (*systempb.ListAuditExportsResponse, error) {
	out, err := s.auditUC.ListExports(ctx, &ExportListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListAuditExportsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoAuditExport),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoAuditExport(e *AuditExport) *systempb.AuditExport {
	p := &systempb.AuditExport{
		FileId:    e.FileID,
		CreatedAt: utils.ToTimestamp(&e.CreatedAt),
		OwnerId:   e.OwnerID,
	}
	if e.File != nil {
		p.File = file.ToProtoFile(e.File)
	}
	return p
}

func toProtoAuditLog(l *AuditLog) *systempb.AuditLog {
	return &systempb.AuditLog{
		Id:            utils.StringW(l.ID),
		CreatedAt:     utils.ToTimestamp(&l.CreatedAt),
		Tenant:        utils.StringW(l.Tenant),
		Actor:         utils.StringW(l.Actor),
		PrincipalType: utils.StringW(l.PrincipalType),
		Operation:     utils.StringW(l.Operation),
		HttpMethod:    utils.StringW(l.HTTPMethod),
		HttpPath:      utils.StringW(l.HTTPPath),
		Status:        utils.Int32W(utils.ConvNum[int32](l.Status)),
		LatencyMs:     utils.Int32W(utils.ConvNum[int32](l.LatencyMs)),
		Ip:            utils.StringW(l.IP),
		UserAgent:     utils.StringW(l.UserAgent),
		DenyReason:    utils.StringW(l.DenyReason),
	}
}
