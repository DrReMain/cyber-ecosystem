package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cyber-ecosystem/shared-go/helper"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type PolicyService struct {
	systempb.UnimplementedPolicyServiceServer

	log      *slog.Logger
	policyUC *PolicyUC
}

func NewPolicyService(logger *slog.Logger, policyUC *PolicyUC) *PolicyService {
	return &PolicyService{
		log:      logger.With("module", "module/policy_service"),
		policyUC: policyUC,
	}
}

func (s *PolicyService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterPolicyServiceServer(srv, s)
}

func (s *PolicyService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterPolicyServiceHTTPServer(srv, s)
}

func (s *PolicyService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterPolicyServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *PolicyService) CreatePolicy(ctx context.Context, in *systempb.CreatePolicyRequest) (*systempb.CreatePolicyResponse, error) {
	kind, params, err := paramsFromProto(in.Params)
	if err != nil {
		return nil, err
	}
	created, err := s.policyUC.Create(ctx, &Policy{
		Kind:   kind,
		Name:   in.GetName(),
		Params: params,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.CreatePolicyResponse{
		Id: utils.StringW(created.ID),
	}, nil
}

func (s *PolicyService) UpdatePolicy(ctx context.Context, in *systempb.UpdatePolicyRequest) (*systempb.UpdatePolicyResponse, error) {
	kind, params, err := paramsFromProto(in.Params)
	if err != nil {
		return nil, err
	}
	if _, err := s.policyUC.Update(ctx, in.FieldsMask, &Policy{
		ID:      in.Id,
		Kind:    kind,
		Name:    in.GetName(),
		Params:  params,
		Enabled: in.GetEnabled(),
	}); err != nil {
		return nil, err
	}
	return &systempb.UpdatePolicyResponse{}, nil
}

func (s *PolicyService) DeletePolicy(ctx context.Context, in *systempb.DeletePolicyRequest) (*systempb.DeletePolicyResponse, error) {
	if _, err := s.policyUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}
	return &systempb.DeletePolicyResponse{}, nil
}

func (s *PolicyService) ListPolicies(ctx context.Context, in *systempb.ListPoliciesRequest) (*systempb.ListPoliciesResponse, error) {
	out, err := s.policyUC.List(ctx, &PolicyListIn{
		PageRequest: helper.EnsurePageRequest(in.Page),
		OrderBy:     in.OrderBy,
		Kind:        in.Kind,
		Name:        in.Name,
		Enabled:     in.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &systempb.ListPoliciesResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, toProtoPolicy),
	}, nil
}

func (s *PolicyService) GetPolicy(ctx context.Context, in *systempb.GetPolicyRequest) (*systempb.GetPolicyResponse, error) {
	p, err := s.policyUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &systempb.GetPolicyResponse{
		Policy: toProtoPolicy(p),
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func paramsFromProto(in *systempb.PolicyParams) (string, map[string]any, error) {
	if in == nil {
		return "", nil, nil
	}
	// The kind is the oneof case name — the kauthz constants verbatim — and
	// the params map is the member message protojson-encoded with proto field
	// names, the same keys the engine plugins decode. protojson omits
	// zero-valued fields, which matches the missing-key-means-default
	// contract plugins already implement.
	m := in.ProtoReflect()
	od := m.Descriptor().Oneofs().ByName("params")
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if !m.Has(fd) {
			continue
		}
		raw, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(m.Get(fd).Message().Interface())
		if err != nil {
			return "", nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("encode policy params: %w", err))
		}
		var params map[string]any
		if err := json.Unmarshal(raw, &params); err != nil {
			return "", nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("decode policy params: %w", err))
		}
		return string(fd.Name()), params, nil
	}
	return "", nil, nil
}

func paramsToProto(kind string, params map[string]any) *systempb.PolicyParams {
	if kind == "" {
		return nil
	}
	// Mirror of paramsFromProto: the oneof descriptor resolves the member
	// message, so a new kind needs no edit here. Errors degrade to nil params
	// rather than failing the read — the kind still shows, unknown keys are
	// discarded, and evaluation fail-closes at the engine.
	raw, err := json.Marshal(params)
	if err != nil {
		return nil
	}
	out := &systempb.PolicyParams{}
	m := out.ProtoReflect()
	od := m.Descriptor().Oneofs().ByName("params")
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if string(fd.Name()) != kind {
			continue
		}
		mt, err := protoregistry.GlobalTypes.FindMessageByName(fd.Message().FullName())
		if err != nil {
			return nil
		}
		msg := mt.New()
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(raw, msg.Interface()); err != nil {
			return nil
		}
		m.Set(fd, protoreflect.ValueOfMessage(msg))
		return out
	}
	return nil
}

func toProtoPolicy(p *Policy) *systempb.Policy {
	return &systempb.Policy{
		Id:         utils.StringW(p.ID),
		CreatedAt:  utils.ToTimestamp(&p.CreatedAt),
		UpdatedAt:  utils.ToTimestamp(&p.UpdatedAt),
		Kind:       utils.StringW(p.Kind),
		Name:       utils.StringW(p.Name),
		Params:     paramsToProto(p.Kind, p.Params),
		Enabled:    utils.BoolW(p.Enabled),
		BoundCount: utils.Int64W(p.BoundCount),
	}
}
