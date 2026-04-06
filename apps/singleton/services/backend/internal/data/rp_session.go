package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/conf"
)

type sessionRP struct {
	RP
}

func NewSessionRP(c *conf.Security, logger log.Logger, store *Store) (biz.SessionRP, error) {
	if store.GetCache() == nil {
		if !c.GetSession() {
			return &noopSessionRP{}, nil
		}
		return nil, fmt.Errorf("security.session requires cache when enabled")
	}
	return &sessionRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_session")),
			store: store,
		},
	}, nil
}

// Repo ----------------------------------------------------------------------------------------------------------------

const sessRevokedKey = "session:revoked:"

func (rp *sessionRP) RevokeSession(ctx context.Context, sid string, ttl time.Duration) error {
	return rp.store.cache.KV.Set(ctx, sessRevokedKey+sid, []byte("1"), ttl)
}

func (rp *sessionRP) IsSessionRevoked(ctx context.Context, sid string) (bool, error) {
	return rp.store.cache.KV.Exist(ctx, sessRevokedKey+sid)
}

// ---------------------------------------------------------------------------------------------------------------------

type noopSessionRP struct{}

func (rp *noopSessionRP) RevokeSession(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (rp *noopSessionRP) IsSessionRevoked(_ context.Context, _ string) (bool, error) {
	return false, nil
}
