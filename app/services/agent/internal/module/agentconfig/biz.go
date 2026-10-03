package agentconfig

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type AgentConfig struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantID  string
	UserID    string
	BaseURL   string
	APIKeySet bool
}

type UserBrief struct {
	ID     string
	Email  *string
	Avatar *string
}

type ListIn struct {
	*commonpb.PageRequest
}

type ListItem struct {
	Config *AgentConfig
	User   *UserBrief
}

type ListOut struct {
	*commonpb.PageResponse
	List []*ListItem
}

type UpdateIn struct {
	BaseURL string
	APIKey  *string // nil = keep the stored ciphertext
}

// Port ----------------------------------------------------------------------------------------------------------------

type AgentConfigRP interface {
	FindByUserID(ctx context.Context, userID string) (*AgentConfig, error) // nil, nil = not configured
	Create(ctx context.Context, mc *AgentConfig, apiKeySealed string) (*AgentConfig, error)
	Update(ctx context.Context, mc *AgentConfig, apiKeySealed *string) error // nil = keep
	DeleteByUserID(ctx context.Context, userID string) error
	List(ctx context.Context, in *ListIn) (*ListOut, error)
}

// UserRP hydrates user identity via system RPCs (cross-service read).
type UserRP interface {
	Hydrate(ctx context.Context, ids []string) map[string]*UserBrief
}

// UC ------------------------------------------------------------------------------------------------------------------

type AgentConfigUC struct {
	shared.UC
	masterKey string

	agentConfigRP AgentConfigRP
	userRP        UserRP
}

func NewAgentConfigUC(logger *slog.Logger, tm shared.Transaction, c *conf.Crypto, agentConfigRP AgentConfigRP, userRP UserRP) (*AgentConfigUC, error) {
	// Fail at boot, not at first seal: an empty master key bricks every write.
	if c == nil || c.GetMasterKey() == "" {
		return nil, fmt.Errorf("crypto.master_key is required")
	}
	return &AgentConfigUC{
		UC:        shared.NewUC(logger.With("module", "module/agentconfig"), tm),
		masterKey: c.GetMasterKey(),

		agentConfigRP: agentConfigRP,
		userRP:        userRP,
	}, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AgentConfigUC) Get(ctx context.Context) (*AgentConfig, error) {
	subject, err := uc.subject(ctx)
	if err != nil {
		return nil, err
	}
	return uc.agentConfigRP.FindByUserID(ctx, subject.UserID)
}

func (uc *AgentConfigUC) Update(ctx context.Context, in *UpdateIn) error {
	subject, err := uc.subject(ctx)
	if err != nil {
		return err
	}
	var sealed *string
	if in.APIKey != nil {
		s, serr := utils.Seal(*in.APIKey, uc.masterKey)
		if serr != nil {
			return errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("seal api key: %w", serr))
		}
		sealed = &s
	}
	// Upsert: one live row per user, so absent means create, present means
	// rewrite — both inside one transaction.
	return uc.Tm.InTx(ctx, func(tctx context.Context) error {
		mc, ferr := uc.agentConfigRP.FindByUserID(tctx, subject.UserID)
		if ferr != nil {
			return ferr
		}
		if mc == nil {
			empty := ""
			if sealed != nil {
				empty = *sealed
			}
			_, cerr := uc.agentConfigRP.Create(tctx, &AgentConfig{UserID: subject.UserID, BaseURL: in.BaseURL}, empty)
			return cerr
		}
		mc.BaseURL = in.BaseURL
		return uc.agentConfigRP.Update(tctx, mc, sealed)
	})
}

func (uc *AgentConfigUC) Delete(ctx context.Context) error {
	subject, err := uc.subject(ctx)
	if err != nil {
		return err
	}
	// Idempotent: upsert makes absence a normal state.
	return uc.agentConfigRP.DeleteByUserID(ctx, subject.UserID)
}

func (uc *AgentConfigUC) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	if _, err := uc.subject(ctx); err != nil {
		return nil, err
	}
	out, err := uc.agentConfigRP.List(ctx, in)
	if err != nil {
		return nil, err
	}
	// Hydration is one eligibility check plus one read set per request; a
	// denied or missing profile renders as absent fields, never a failed row.
	ids := make([]string, 0, len(out.List))
	seen := make(map[string]struct{}, len(out.List))
	for _, it := range out.List {
		if _, ok := seen[it.Config.UserID]; !ok {
			seen[it.Config.UserID] = struct{}{}
			ids = append(ids, it.Config.UserID)
		}
	}
	briefs := uc.userRP.Hydrate(ctx, ids)
	for _, it := range out.List {
		it.User = briefs[it.Config.UserID]
	}
	return out, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AgentConfigUC) subject(ctx context.Context) (*security.Subject, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, kauthz.ErrMissingSubject
	}
	return subject, nil
}
