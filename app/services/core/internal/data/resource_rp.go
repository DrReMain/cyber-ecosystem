package data

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cyber-ecosystem/shared-go/helper"

	"cyber-ecosystem/app/services/core/internal/biz"
	"cyber-ecosystem/app/services/core/internal/platform"
)

type resourceRP struct {
	RP
}

func NewResourceRP(logger *slog.Logger, p *platform.Platform) biz.ResourceRP {
	return &resourceRP{
		RP: RP{
			log:      logger.With("module", "data/resource_rp"),
			platform: p,
		},
	}
}

// Repo --------------------------------------------------------------------------------------------------------

func (rp *resourceRP) ListResource(ctx context.Context) ([]*biz.ResourceService, error) {
	const protoPrefix = "cyber/core/v1/"
	var services []*biz.ResourceService

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(fd.Path(), protoPrefix) {
			return true
		}
		for i := 0; i < fd.Services().Len(); i++ {
			services = append(services, buildResourceService(fd.Services().Get(i), fd))
		}
		return true
	})

	return services, nil
}

// Private --------------------------------------------------------------------------------------------------------

func buildResourceService(sd protoreflect.ServiceDescriptor, fd protoreflect.FileDescriptor) *biz.ResourceService {
	svc := &biz.ResourceService{
		Name:       string(sd.Name()),
		FullName:   string(sd.FullName()),
		Package:    string(fd.Package()),
		SourceFile: fd.Path(),
		Comment:    helper.GetServiceComment(sd),
	}

	methods := make([]*biz.ResourceMethod, sd.Methods().Len())
	for i := 0; i < sd.Methods().Len(); i++ {
		methods[i] = buildResourceMethod(sd.Methods().Get(i))
	}
	svc.Methods = methods

	return svc
}

func buildResourceMethod(md protoreflect.MethodDescriptor) *biz.ResourceMethod {
	m := &biz.ResourceMethod{
		Name:             string(md.Name()),
		FullName:         string(md.FullName()),
		RequestName:      string(md.Input().Name()),
		RequestFullName:  string(md.Input().FullName()),
		ResponseName:     string(md.Output().Name()),
		ResponseFullName: string(md.Output().FullName()),
		Comment:          helper.GetMethodComment(md),
	}

	if options := md.Options(); options != nil {
		if rule, ok := proto.GetExtension(options, annotations.E_Http).(*annotations.HttpRule); ok && rule != nil {
			m.HttpMethod, m.HttpPath = helper.ExtractHTTP(rule)
		}
	}

	return m
}
