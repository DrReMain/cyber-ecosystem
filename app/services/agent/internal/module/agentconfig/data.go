package agentconfig

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/agent/internal/ent"
	entagentconfig "cyber-ecosystem/app/services/agent/internal/ent/agentconfig"
	"cyber-ecosystem/app/services/agent/internal/platform"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type agentConfigRP struct {
	shared.RP
}

func NewAgentConfigRP(logger *slog.Logger, p *platform.Platform) AgentConfigRP {
	return &agentConfigRP{RP: shared.NewRP(logger.With("module", "module/agentconfig_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *agentConfigRP) Create(ctx context.Context, mc *AgentConfig, apiKeySealed string) (*AgentConfig, error) {
	created, err := rp.Platform.GetClient(ctx).AgentConfig.Create().
		SetUserID(mc.UserID).
		SetBaseURL(mc.BaseURL).
		SetAPIKey(apiKeySealed).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapAgentConfig(created), nil
}

func (rp *agentConfigRP) Update(ctx context.Context, mc *AgentConfig, apiKeySealed *string) error {
	q := rp.Platform.GetClient(ctx).AgentConfig.UpdateOneID(mc.ID).
		SetBaseURL(mc.BaseURL)
	if apiKeySealed != nil {
		q = q.SetAPIKey(*apiKeySealed)
	}
	if err := q.Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *agentConfigRP) DeleteByUserID(ctx context.Context, userID string) error {
	// SoftDeleteMixin rewrites delete into the tombstone update.
	if _, err := rp.Platform.GetClient(ctx).AgentConfig.Delete().
		Where(entagentconfig.UserIDEQ(userID)).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *agentConfigRP) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	query := rp.Platform.GetClient(ctx).AgentConfig.Query().
		Order(ent.Desc(entagentconfig.FieldCreatedAt))
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List: utils.SliceMap(rows, func(d *ent.AgentConfig) *ListItem {
			return &ListItem{Config: mapAgentConfig(d)}
		}),
	}, nil
}

func (rp *agentConfigRP) FindSealedByUserID(ctx context.Context, userID string) (*AgentConfig, string, error) {
	// The partial unique index guarantees at most one live row per user.
	d, err := rp.Platform.GetClient(ctx).AgentConfig.Query().
		Where(entagentconfig.UserIDEQ(userID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", rp.Platform.HandleEntError(err)
	}
	return mapAgentConfig(d), d.APIKey, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapAgentConfig(d *ent.AgentConfig) *AgentConfig {
	return &AgentConfig{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		TenantID:  d.TenantID,
		UserID:    d.UserID,
		BaseURL:   d.BaseURL,
		APIKeySet: d.APIKey != "",
	}
}
