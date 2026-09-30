package role

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/helper"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/authz"
)

// Struct --------------------------------------------------------------------------------------------------------------

type RoleService struct {
	systempb.UnimplementedRoleServiceServer

	log    *slog.Logger
	roleUC *RoleUC
}

func NewRoleService(logger *slog.Logger, roleUC *RoleUC) *RoleService {
	return &RoleService{
		log:    logger.With("module", "module/role_service"),
		roleUC: roleUC,
	}
}

func (s *RoleService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterRoleServiceServer(srv, s)
}

func (s *RoleService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterRoleServiceHTTPServer(srv, s)
}

func (s *RoleService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterRoleServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------
func (s *RoleService) CreateRole(ctx context.Context, in *systempb.CreateRoleRequest) (*systempb.CreateRoleResponse, error) {
	created, err := s.roleUC.Create(ctx, &Role{
		Code:   in.GetCode(),
		Name:   in.GetName(),
		Remark: in.Remark,
	}, grantsFromProto(in.Grants))
	if err != nil {
		return nil, err
	}
	return &systempb.CreateRoleResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, in *systempb.UpdateRoleRequest) (*systempb.UpdateRoleResponse, error) {
	if _, err := s.roleUC.Update(ctx, in.FieldsMask, &Role{
		ID:     in.Id,
		Name:   in.GetName(),
		Remark: in.Remark,
	}, grantsFromProto(in.Grants)); err != nil {
		return nil, err
	}
	return &systempb.UpdateRoleResponse{}, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, in *systempb.UpdateRoleStatusRequest) (*systempb.UpdateRoleStatusResponse, error) {
	if _, err := s.roleUC.UpdateStatus(ctx, in.Id, in.GetEnabled()); err != nil {
		return nil, err
	}
	return &systempb.UpdateRoleStatusResponse{}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, in *systempb.DeleteRoleRequest) (*systempb.DeleteRoleResponse, error) {
	if _, err := s.roleUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeleteRoleResponse{}, nil
}

func (s *RoleService) ListRoles(ctx context.Context, in *systempb.ListRolesRequest) (*systempb.ListRolesResponse, error) {
	out, err := s.roleUC.List(ctx, &RoleListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Code:        in.Code,
		Name:        in.Name,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListRolesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoRole),
	}, nil
}

func (s *RoleService) GetRole(ctx context.Context, in *systempb.GetRoleRequest) (*systempb.GetRoleResponse, error) {
	r, err := s.roleUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetRoleResponse{
		Role:   toProtoRole(r),
		Grants: utils.SliceMap(r.Grants, toProtoGrant),
	}, nil
}

func (s *RoleService) ListRolesByPrincipal(ctx context.Context, in *systempb.ListRolesByPrincipalRequest) (*systempb.ListRolesByPrincipalResponse, error) {
	list, err := s.roleUC.ListByPrincipal(ctx, in.PrincipalType, in.PrincipalId)
	if err != nil {
		return nil, err
	}
	return &systempb.ListRolesByPrincipalResponse{
		List: utils.SliceMap(list, toProtoRole),
	}, nil
}

func (s *RoleService) RemoveRoleMember(ctx context.Context, in *systempb.RemoveRoleMemberRequest) (*systempb.RemoveRoleMemberResponse, error) {
	if err := s.roleUC.RemoveMember(ctx, in.RoleId, in.PrincipalType, in.PrincipalId); err != nil {
		return nil, err
	}
	return &systempb.RemoveRoleMemberResponse{}, nil
}

func (s *RoleService) ListRoleMembers(ctx context.Context, in *systempb.ListRoleMembersRequest) (*systempb.ListRoleMembersResponse, error) {
	out, err := s.roleUC.ListMembers(ctx, &MemberListIn{
		RoleID:        in.RoleId,
		PageRequest:   helper.EnsurePageRequest(in.Page),
		PrincipalType: in.PrincipalType,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListRoleMembersResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoMember),
	}, nil
}

func (s *RoleService) ExplainOperation(ctx context.Context, in *systempb.ExplainOperationRequest) (*systempb.ExplainOperationResponse, error) {
	res, err := s.roleUC.ExplainOperation(ctx, in.PrincipalType, in.PrincipalId, in.Operation)
	if err != nil {
		return nil, err
	}
	verdict := "DENY"
	if res.Allowed {
		verdict = "ALLOW"
	}
	return &systempb.ExplainOperationResponse{
		Verdict:   verdict,
		Builtin:   res.Builtin,
		RoleCodes: res.RoleCodes,
		Hits:      utils.SliceMap(res.Hits, toProtoHit),
	}, nil
}

func (s *RoleService) PreviewGrants(ctx context.Context, in *systempb.PreviewGrantsRequest) (*systempb.PreviewGrantsResponse, error) {
	ops, err := s.roleUC.PreviewGrants(ctx, in.RoleCodes)
	if err != nil {
		return nil, err
	}
	return &systempb.PreviewGrantsResponse{Operations: ops}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toProtoRole(r *Role) *systempb.Role {
	return &systempb.Role{
		Id:            utils.StringW(r.ID),
		CreatedAt:     utils.ToTimestamp(&r.CreatedAt),
		UpdatedAt:     utils.ToTimestamp(&r.UpdatedAt),
		Code:          utils.StringW(r.Code),
		Name:          utils.StringW(r.Name),
		Enabled:       utils.BoolW(r.Enabled),
		Remark:        utils.Wrap(r.Remark, utils.StringW),
		GrantCount:    utils.Int64W(r.GrantCount),
		BoundCount:    utils.Int64W(r.BoundCount),
		GlobalGrant:   utils.BoolW(r.GlobalGrant),
		WildcardCount: utils.Int64W(r.WildcardCount),
		NarrowedCount: utils.Int64W(r.NarrowedCount),
		PolicyCount:   utils.Int64W(r.PolicyCount),
	}
}

func toProtoGrant(p *Permission) *systempb.RoleGrant {
	return &systempb.RoleGrant{
		Operation: p.Operation,
		ScopeKind: scopeKindToProto(p.ScopeKind),
		PolicyIds: p.PolicyIDs,
	}
}

func toProtoMember(m *RoleMember) *systempb.RoleMember {
	return &systempb.RoleMember{
		PrincipalType: utils.StringW(m.PrincipalType),
		PrincipalId:   utils.StringW(m.PrincipalID),
		CreatedAt:     utils.ToTimestamp(&m.CreatedAt),
	}
}

func grantsFromProto(gs []*systempb.RoleGrant) []*Permission {
	out := make([]*Permission, 0, len(gs))
	for _, g := range gs {
		out = append(out, &Permission{
			Operation: g.GetOperation(),
			ScopeKind: scopeKindFromProto(g.GetScopeKind()),
			PolicyIDs: g.GetPolicyIds(),
		})
	}
	return out
}

func toProtoHit(h authz.GrantHit) *systempb.ExplainHit {
	return &systempb.ExplainHit{
		Pattern:   h.Pattern,
		RoleCode:  h.RoleCode,
		ScopeKind: scopeKindToProto(h.ScopeKind),
		Policies:  utils.SliceMap(h.Policies, toProtoPolicyState),
	}
}

func toProtoPolicyState(p authz.PolicyState) *systempb.ExplainPolicy {
	return &systempb.ExplainPolicy{
		Kind:  p.Kind,
		Name:  p.Name,
		State: p.State,
	}
}

func scopeKindFromProto(k systempb.ScopeKind) string {
	switch k {
	case systempb.ScopeKind_SCOPE_KIND_ALL:
		return kauthz.ScopeKindAll
	case systempb.ScopeKind_SCOPE_KIND_SELF:
		return kauthz.ScopeKindSelf
	case systempb.ScopeKind_SCOPE_KIND_DEPT_TREE:
		return kauthz.ScopeKindDeptTree
	default:
		return ""
	}
}

func scopeKindToProto(k string) systempb.ScopeKind {
	switch k {
	case kauthz.ScopeKindAll:
		return systempb.ScopeKind_SCOPE_KIND_ALL
	case kauthz.ScopeKindSelf:
		return systempb.ScopeKind_SCOPE_KIND_SELF
	case kauthz.ScopeKindDeptTree:
		return systempb.ScopeKind_SCOPE_KIND_DEPT_TREE
	default:
		return systempb.ScopeKind_SCOPE_KIND_UNSPECIFIED
	}
}
