package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/conf"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
)

type dataScopeSnapshotRP struct {
	RP
}

func NewDataScopeSnapshotRP(c *conf.Security, logger log.Logger, store *Store) (biz.DataScopeSnapshotRP, error) {
	if store.GetCache() == nil {
		if !c.GetDataScope() {
			return &noopDataScopeSnapshotRP{}, nil
		}
		return nil, fmt.Errorf("security.data_scope requires cache when enabled")
	}
	return &dataScopeSnapshotRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_data_scope_snapshot")),
			store: store,
		},
	}, nil
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *dataScopeSnapshotRP) Get(ctx context.Context, userID string) (*datascope.ScopeSnapshot, bool) {
	key := datascope.SnapshotCacheKey(userID)
	data, err := rp.store.cache.KV.Get(ctx, key)
	if err != nil || data == nil {
		return nil, false
	}
	snapshot, err := datascope.UnmarshalSnapshot(data)
	if err != nil {
		return nil, false
	}
	return snapshot, true
}

func (rp *dataScopeSnapshotRP) Set(ctx context.Context, userID string, snapshot *datascope.ScopeSnapshot) error {
	data, err := datascope.MarshalSnapshot(snapshot)
	if err != nil {
		return err
	}
	key := datascope.SnapshotCacheKey(userID)
	return rp.store.cache.KV.Set(ctx, key, data, 30*time.Minute)
}

func (rp *dataScopeSnapshotRP) Invalidate(ctx context.Context, userID string) error {
	key := datascope.SnapshotCacheKey(userID)
	return rp.store.cache.KV.Delete(ctx, key)
}

type noopDataScopeSnapshotRP struct{}

func (rp *noopDataScopeSnapshotRP) Get(_ context.Context, _ string) (*datascope.ScopeSnapshot, bool) {
	return nil, false
}

func (rp *noopDataScopeSnapshotRP) Set(_ context.Context, _ string, _ *datascope.ScopeSnapshot) error {
	return nil
}

func (rp *noopDataScopeSnapshotRP) Invalidate(_ context.Context, _ string) error {
	return nil
}
