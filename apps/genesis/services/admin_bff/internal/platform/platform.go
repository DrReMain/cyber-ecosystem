package platform

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
)

type CacheErrorHandler func(error) error

type Platform struct {
	cache            *cache.Cache
	handleCacheError CacheErrorHandler
	articleClient    genesisV1.ArticleServiceClient
	resourceClient   genesisV1.ResourceServiceClient
}

func NewPlatform(
	logger log.Logger,
	cache *cache.Cache,
	handleCacheError CacheErrorHandler,
	articleClient genesisV1.ArticleServiceClient,
	resourceClient genesisV1.ResourceServiceClient,
) (*Platform, func(), error) {
	p := &Platform{
		cache:            cache,
		handleCacheError: handleCacheError,
		articleClient:    articleClient,
		resourceClient:   resourceClient,
	}
	return p,
		func() {
			if cache != nil && cache.Client != nil {
				_ = cache.Client.Close()
			}
		},
		nil
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (p *Platform) GetArticleClient() genesisV1.ArticleServiceClient {
	return p.articleClient
}

func (p *Platform) GetResourceClient() genesisV1.ResourceServiceClient {
	return p.resourceClient
}

func (p *Platform) GetCache() *cache.Cache {
	return p.cache
}

func (p *Platform) HandleCacheError(err error) error {
	return p.handleCacheError(err)
}

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewCache,
	NewCacheErrorHandler,
	NewGRPCResourceClient,
	NewGRPCArticleClient,
)
