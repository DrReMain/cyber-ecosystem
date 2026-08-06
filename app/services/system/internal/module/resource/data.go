package resource

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cyber-ecosystem/shared-go/helper"

	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

type resourceRP struct {
	shared.RP
}

func NewResourceRP(logger *slog.Logger, p *platform.Platform) ResourceRP {
	return &resourceRP{
		RP: shared.NewRP(logger.With("module", "module/resource_rp"), p),
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *resourceRP) ListResource(ctx context.Context) ([]*ServiceMeta, error) {
	const protoPrefix = "cyber/system/v1/"
	var services []*ServiceMeta

	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(fd.Path(), protoPrefix) {
			return true
		}
		for i := 0; i < fd.Services().Len(); i++ {
			services = append(services, buildServiceMeta(fd.Services().Get(i), fd))
		}
		return true
	})

	return services, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func buildServiceMeta(sd protoreflect.ServiceDescriptor, fd protoreflect.FileDescriptor) *ServiceMeta {
	svc := &ServiceMeta{
		Name:       string(sd.Name()),
		FullName:   string(sd.FullName()),
		Package:    string(fd.Package()),
		SourceFile: fd.Path(),
		Comment:    helper.GetServiceComment(sd),
	}

	methods := make([]*ResourceMethod, sd.Methods().Len())
	for i := 0; i < sd.Methods().Len(); i++ {
		methods[i] = buildResourceMethod(sd.Methods().Get(i))
	}
	svc.Methods = methods

	return svc
}

func buildResourceMethod(md protoreflect.MethodDescriptor) *ResourceMethod {
	m := &ResourceMethod{
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
