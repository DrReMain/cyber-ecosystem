package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/biz"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/platform"
)

type resourceRP struct {
	RP
}

func NewResourceRP(logger log.Logger, p *platform.Platform) biz.ResourceRP {
	return &resourceRP{
		RP: RP{
			log:      log.NewHelper(log.With(logger, "module", "data/resource_rp")),
			platform: p,
		},
	}
}

// region[rgba(0,188,212,0.12)] 🩵 Repo --------------------------------------------------------------------------------

func (rp *resourceRP) ListResource(ctx context.Context) ([]*biz.ResourceService, error) {
	resp, err := rp.platform.GetResourceClient().ListResource(ctx, &genesisV1.ListResourceRequest{})
	if err != nil {
		return nil, err
	}
	return utils.SliceMap(resp.List, protoToResourceService), nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func protoToResourceService(s *genesisV1.Service) *biz.ResourceService {
	return &biz.ResourceService{
		Name:       s.Name,
		FullName:   s.FullName,
		Package:    s.Package,
		SourceFile: s.SourceFile,
		Comment:    s.Comment,
		Methods: func() []*biz.ResourceMethod {
			r := make([]*biz.ResourceMethod, 0, len(s.Methods))
			for _, m := range s.Methods {
				r = append(r, &biz.ResourceMethod{
					Name:             m.Name,
					FullName:         m.FullName,
					RequestName:      m.RequestName,
					RequestFullName:  m.RequestFullName,
					ResponseName:     m.ResponseName,
					ResponseFullName: m.ResponseFullName,
					HttpMethod:       m.HttpMethod,
					HttpPath:         m.HttpPath,
					Comment:          m.Comment,
				})
			}
			return r
		}(),
	}
}
