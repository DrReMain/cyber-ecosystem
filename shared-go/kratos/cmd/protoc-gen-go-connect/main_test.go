package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestHandleCall(t *testing.T) {
	cases := []struct {
		name                       string
		clientStream, serverStream bool
		wantHandle                 string
	}{
		{"unary", false, false, "HandleUnary"},
		{"server_stream", false, true, "HandleServerStream"},
		{"client_stream", true, false, "HandleClientStream"},
		{"bidi_stream", true, true, "HandleBidiStream"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handle, proc := handleCall("cyber.transfer.v1.TransferService", "Subscribe", tc.clientStream, tc.serverStream)
			if handle != tc.wantHandle {
				t.Fatalf("handle = %q, want %q", handle, tc.wantHandle)
			}
			wantProc := "/cyber.transfer.v1.TransferService/Subscribe"
			if proc != wantProc {
				t.Fatalf("procedure = %q, want %q", proc, wantProc)
			}
		})
	}
}

// msgType is a tiny helper that builds a trivial message descriptor.
func msgType(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{Name: proto.String(name)}
}

// buildDemoFile builds a FileDescriptorProto fixture with one "Demo" service in
// package cyber.demo.v1 exposing all four streaming variants. The go_package
// option resolves to package name "v1", matching the real generated output.
func buildDemoFile() *descriptorpb.FileDescriptorProto {
	req := ".cyber.demo.v1.Req"
	resp := ".cyber.demo.v1.Resp"
	method := func(name string, client, server bool) *descriptorpb.MethodDescriptorProto {
		m := &descriptorpb.MethodDescriptorProto{
			Name:       proto.String(name),
			InputType:  proto.String(req),
			OutputType: proto.String(resp),
		}
		if client {
			m.ClientStreaming = proto.Bool(true)
		}
		if server {
			m.ServerStreaming = proto.Bool(true)
		}
		return m
	}
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("demo.proto"),
		Package: proto.String("cyber.demo.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("cyber-ecosystem/gen/go/cyber/demo/v1;v1"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgType("Req"), msgType("Resp")},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("Demo"),
			Method: []*descriptorpb.MethodDescriptorProto{
				method("Unary", false, false),
				method("ServerStream", false, true),
				method("ClientStream", true, false),
				method("Bidi", true, true),
			},
		}},
	}
}

// runGen drives the real protogen pipeline against the supplied file
// descriptors and returns the rendered content of the generated connect file
// (or "" with ok=false if no file was produced).
//
// The pipeline mirrors what main() does: Options{}.New builds the Plugin from a
// CodeGeneratorRequest (ProtoFile carries the descriptors; FileToGenerate marks
// which files are active, which sets f.Generate = true), then genFile emits the
// output via p.NewGeneratedFile, and plugin.Response() renders every generated
// file to its final gofmt'd bytes.
func runGen(t *testing.T, files ...*descriptorpb.FileDescriptorProto) *pluginpb.CodeGeneratorResponse {
	t.Helper()
	req := &pluginpb.CodeGeneratorRequest{
		ProtoFile: files,
	}
	for _, f := range files {
		req.FileToGenerate = append(req.FileToGenerate, f.GetName())
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.Options{}.New: %v", err)
	}
	for _, f := range plugin.Files {
		if !f.Generate {
			continue
		}
		genFile(plugin, f)
	}
	return plugin.Response()
}

// generatedContent finds the content of the connect file generated for the
// given source .proto. genFile names the output "<prefix>_connect.pb.go" where
// <prefix> is the file's go_package-derived GeneratedFilenamePrefix (a full
// import path), so we match by the _connect.pb.go suffix anchored to the
// source proto's base name. Returns "" if no such file was produced.
func generatedContent(t *testing.T, resp *pluginpb.CodeGeneratorResponse, sourceProto string) string {
	t.Helper()
	// trim any directory from the source proto, e.g. "demo.proto" -> "demo"
	base := sourceProto
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "."); i >= 0 {
		base = base[:i]
	}
	wantSuffix := base + "_connect.pb.go"
	for _, rf := range resp.File {
		name := rf.GetName()
		if strings.HasSuffix(name, wantSuffix) {
			return rf.GetContent()
		}
	}
	return ""
}

func TestGenerateGolden(t *testing.T) {
	resp := runGen(t, buildDemoFile())
	content := generatedContent(t, resp, "demo.proto")
	if content == "" {
		t.Fatalf("no connect file generated for demo.proto; response files: %d", len(resp.File))
	}

	// Package declaration derived from go_package option.
	if !strings.Contains(content, "package v1") {
		t.Errorf("generated content missing package declaration\ngot:\n%s", content)
	}
	// Kratos connect transport import, rendered as an aliased import.
	if !strings.Contains(content, `connect "cyber-ecosystem/shared-go/kratos/transport/connect"`) {
		t.Errorf("generated content missing connect import\ngot:\n%s", content)
	}
	// DO NOT EDIT header.
	if !strings.Contains(content, "Code generated by protoc-gen-go-connect. DO NOT EDIT.") {
		t.Errorf("generated content missing DO NOT EDIT header\ngot:\n%s", content)
	}
	// Register function signature.
	wantSig := "func RegisterDemoConnectServer(srv *connect.Server, svc DemoServer) {"
	if !strings.Contains(content, wantSig) {
		t.Errorf("generated content missing register signature %q\ngot:\n%s", wantSig, content)
	}

	// All four streaming variants mapped to their Handle* helpers with the
	// exact fully-qualified procedure paths.
	wantLines := []string{
		`connect.HandleUnary(srv, "/cyber.demo.v1.Demo/Unary", svc.Unary)`,
		`connect.HandleServerStream(srv, "/cyber.demo.v1.Demo/ServerStream", svc.ServerStream)`,
		`connect.HandleClientStream(srv, "/cyber.demo.v1.Demo/ClientStream", svc.ClientStream)`,
		`connect.HandleBidiStream(srv, "/cyber.demo.v1.Demo/Bidi", svc.Bidi)`,
	}
	for _, line := range wantLines {
		if !strings.Contains(content, line) {
			t.Errorf("generated content missing expected line:\n  %s\ngot:\n%s", line, content)
		}
	}

	// Sanity: the generated code must be valid Go (Response() already gofmt'd
	// it via GeneratedFile.Content(); if parsing had failed Response().Error
	// would be set instead of File). Assert no error surfaced.
	if e := resp.GetError(); e != "" {
		t.Fatalf("plugin reported error rendering output: %s", e)
	}
}

func TestGenerateEmptyServiceFileSkipped(t *testing.T) {
	// A file with no services must produce no generated connect file.
	noService := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("empty.proto"),
		Package: proto.String("cyber.empty.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("cyber-ecosystem/gen/go/cyber/empty/v1;v1"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgType("Ping")},
	}
	resp := runGen(t, noService)

	if got := len(resp.File); got != 0 {
		t.Errorf("expected no generated files for service-less input, got %d", got)
		for i, rf := range resp.File {
			t.Logf("  file[%d] = %s", i, rf.GetName())
		}
	}
	// No error either — skipping is a clean no-op.
	if e := resp.GetError(); e != "" {
		t.Errorf("expected clean skip, got plugin error: %s", e)
	}
}

func TestGenerateMultipleServicesPerFile(t *testing.T) {
	req := ".cyber.demo.v1.Req"
	resp := ".cyber.demo.v1.Resp"
	twin := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("twin.proto"),
		Package: proto.String("cyber.demo.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("cyber-ecosystem/gen/go/cyber/demo/v1;v1"),
		},
		MessageType: []*descriptorpb.DescriptorProto{msgType("Req"), msgType("Resp")},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: proto.String("Alpha"),
				Method: []*descriptorpb.MethodDescriptorProto{{
					Name:       proto.String("Do"),
					InputType:  proto.String(req),
					OutputType: proto.String(resp),
				}},
			},
			{
				Name: proto.String("Beta"),
				Method: []*descriptorpb.MethodDescriptorProto{{
					Name:       proto.String("Run"),
					InputType:  proto.String(req),
					OutputType: proto.String(resp),
					ServerStreaming: proto.Bool(true),
				}},
			},
		},
	}

	got := runGen(t, twin)
	content := generatedContent(t, got, "twin.proto")
	if content == "" {
		t.Fatalf("no connect file generated for twin.proto")
	}

	// Both services get their own Register entry point.
	if !strings.Contains(content, "func RegisterAlphaConnectServer(srv *connect.Server, svc AlphaServer) {") {
		t.Errorf("missing RegisterAlphaConnectServer\ngot:\n%s", content)
	}
	if !strings.Contains(content, "func RegisterBetaConnectServer(srv *connect.Server, svc BetaServer) {") {
		t.Errorf("missing RegisterBetaConnectServer\ngot:\n%s", content)
	}

	// Procedure paths use each service's own fully-qualified name.
	if !strings.Contains(content, `connect.HandleUnary(srv, "/cyber.demo.v1.Alpha/Do", svc.Do)`) {
		t.Errorf("missing Alpha/Do procedure line\ngot:\n%s", content)
	}
	if !strings.Contains(content, `connect.HandleServerStream(srv, "/cyber.demo.v1.Beta/Run", svc.Run)`) {
		t.Errorf("missing Beta/Run procedure line\ngot:\n%s", content)
	}
}
