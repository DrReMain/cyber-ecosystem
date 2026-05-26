package data

import (
	"context"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/genesis/services/base/internal/biz"
	"cyber-ecosystem/apps/genesis/services/base/internal/platform"
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

func (rp *resourceRP) ListServices(ctx context.Context) ([]*biz.ResourceService, error) {
	const (
		project = "genesis"
		dir     = "/api/v1/"
	)
	var result []*biz.ResourceService
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(fd.Path(), project+dir) {
			return true
		}

		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			sd := services.Get(i)
			methods := sd.Methods()

			service := &biz.ResourceService{
				Name:       string(sd.Name()),
				FullName:   string(sd.FullName()),
				Package:    string(fd.Package()),
				SourceFile: fd.Path(),
				Comment:    utils.GetServiceComment(sd),
			}

			for j := 0; j < methods.Len(); j++ {
				md := methods.Get(j)
				method := &biz.ResourceMethod{
					Name:             string(md.Name()),
					FullName:         string(md.FullName()),
					RequestName:      string(md.Input().Name()),
					RequestFullName:  string(md.Input().FullName()),
					ResponseName:     string(md.Output().Name()),
					ResponseFullName: string(md.Output().FullName()),
					Comment:          utils.GetMethodComment(md),
				}
				if md.Options() != nil {
					httpRule, ok := proto.GetExtension(md.Options(), annotations.E_Http).(*annotations.HttpRule)
					if ok && httpRule != nil {
						method.HttpMethod, method.HttpPath = utils.ExtractHTTP(httpRule)
					}
				}
				service.Methods = append(service.Methods, method)
			}

			result = append(result, service)
		}
		return true
	})

	return result, nil
}
