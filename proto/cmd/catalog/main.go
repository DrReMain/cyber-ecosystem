// Command catalog derives per-package operation manifests from the compiled
// proto image. Fed by `buf build -o -#format=binpb`, it writes one JSON file
// per proto package under gen/catalog; the system service loads these
// manifests so its grant catalog spans every deployed service without
// importing their code (composition = which manifests ship).
package main

import (
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

// serviceMeta/methodMeta mirror the system resource module's ServiceMeta /
// ResourceMethod DOs — the json tags are the manifest contract; keep both
// sides in sync when a field moves.
type serviceMeta struct {
	Name       string        `json:"name"`
	FullName   string        `json:"full_name"`
	Package    string        `json:"package"`
	SourceFile string        `json:"source_file"`
	Comment    string        `json:"comment"`
	Methods    []*methodMeta `json:"methods"`
}

type methodMeta struct {
	Name             string `json:"name"`
	FullName         string `json:"full_name"`
	RequestName      string `json:"request_name"`
	RequestFullName  string `json:"request_full_name"`
	ResponseName     string `json:"response_name"`
	ResponseFullName string `json:"response_full_name"`
	HttpMethod       string `json:"http_method"`
	HttpPath         string `json:"http_path"`
	Comment          string `json:"comment"`
	Builtin          bool   `json:"builtin"`
	Access           string `json:"access"`
	Datascope        bool   `json:"datascope"`
}

func main() {
	out := flag.String("out", "gen/catalog", "output directory for per-package manifests")
	flag.Parse()

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail(fmt.Errorf("read image from stdin: %w", err))
	}
	var fds descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(raw, &fds); err != nil {
		fail(fmt.Errorf("parse FileDescriptorSet: %w", err))
	}

	byPkg := make(map[string][]*serviceMeta)
	for _, fd := range fds.GetFile() {
		if !strings.HasPrefix(fd.GetName(), "cyber/") || len(fd.GetService()) == 0 {
			continue
		}
		pkg := fd.GetPackage()
		for _, sd := range fd.GetService() {
			byPkg[pkg] = append(byPkg[pkg], buildService(fd, sd))
		}
	}
	if len(byPkg) == 0 {
		fail(fmt.Errorf("no service-bearing proto files under cyber/ in the image"))
	}

	if err := os.MkdirAll(*out, 0o750); err != nil {
		fail(fmt.Errorf("create output dir: %w", err))
	}
	for _, pkg := range slices.Sorted(maps.Keys(byPkg)) {
		services := byPkg[pkg]
		sort.Slice(services, func(i, j int) bool { return services[i].FullName < services[j].FullName })
		blob, err := utils.MarshalIndent(services, "", "  ")
		if err != nil {
			fail(fmt.Errorf("marshal %s: %w", pkg, err))
		}
		path := filepath.Join(*out, pkg+".json")
		if err := os.WriteFile(path, append(blob, '\n'), 0o600); err != nil {
			fail(fmt.Errorf("write %s: %w", path, err))
		}
	}
}

func buildService(fd *descriptorpb.FileDescriptorProto, sd *descriptorpb.ServiceDescriptorProto) *serviceMeta {
	svc := &serviceMeta{
		Name:       sd.GetName(),
		FullName:   fd.GetPackage() + "." + sd.GetName(),
		Package:    fd.GetPackage(),
		SourceFile: fd.GetName(),
	}
	if opts := sd.GetOptions(); opts != nil {
		if v, ok := proto.GetExtension(opts, extv1.E_ServiceComment).(string); ok {
			svc.Comment = v
		}
	}
	methods := make([]*methodMeta, len(sd.GetMethod()))
	for i, md := range sd.GetMethod() {
		methods[i] = buildMethod(svc.FullName, md)
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Name < methods[j].Name })
	svc.Methods = methods
	return svc
}

func buildMethod(serviceFullName string, md *descriptorpb.MethodDescriptorProto) *methodMeta {
	m := &methodMeta{
		Name:             md.GetName(),
		FullName:         serviceFullName + "." + md.GetName(),
		RequestFullName:  strings.TrimPrefix(md.GetInputType(), "."),
		ResponseFullName: strings.TrimPrefix(md.GetOutputType(), "."),
	}
	m.RequestName = m.RequestFullName[strings.LastIndex(m.RequestFullName, ".")+1:]
	m.ResponseName = m.ResponseFullName[strings.LastIndex(m.ResponseFullName, ".")+1:]
	if opts := md.GetOptions(); opts != nil {
		if rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule); ok && rule != nil {
			m.HttpMethod, m.HttpPath = helper.ExtractHTTP(rule)
		}
		m.Builtin, _ = proto.GetExtension(opts, extv1.E_Builtin).(bool)
		m.Datascope, _ = proto.GetExtension(opts, extv1.E_Datascope).(bool)
		if proto.HasExtension(opts, extv1.E_Access) {
			acc, _ := proto.GetExtension(opts, extv1.E_Access).(extv1.Access)
			m.Access = strings.TrimPrefix(acc.String(), "ACCESS_")
		}
		if d, ok := proto.GetExtension(opts, extv1.E_Method).(*extv1.MethodDesc); ok && d != nil {
			m.Comment = d.GetComment()
		}
	}
	return m
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "catalog:", err)
	os.Exit(1)
}
