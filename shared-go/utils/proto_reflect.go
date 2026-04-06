package utils

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cyber-ecosystem/contracts/go/desc"
)

func ExtractHTTP(rule *annotations.HttpRule) (method, path string) {
	switch p := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		return "GET", p.Get
	case *annotations.HttpRule_Put:
		return "PUT", p.Put
	case *annotations.HttpRule_Post:
		return "POST", p.Post
	case *annotations.HttpRule_Delete:
		return "DELETE", p.Delete
	case *annotations.HttpRule_Patch:
		return "PATCH", p.Patch
	}
	return "", ""
}

func GetServiceComment(sd protoreflect.ServiceDescriptor) string {
	var options = sd.Options()
	if options == nil {
		return ""
	}
	if v, ok := proto.GetExtension(options, desc.E_ServiceComment).(string); ok {
		return v
	}
	return ""
}

func GetMethodComment(md protoreflect.MethodDescriptor) string {
	var options = md.Options()
	if options == nil {
		return ""
	}
	if v, ok := proto.GetExtension(options, desc.E_MethodComment).(string); ok {
		return v
	}
	return ""
}
