package helper

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	extv1 "cyber-ecosystem/gen/go/cyber/shared/ext/v1"
)

// ExtractHTTP returns the HTTP method and path from an HttpRule.
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

// GetServiceComment extracts the service_comment extension from a service descriptor.
func GetServiceComment(sd protoreflect.ServiceDescriptor) string {
	options := sd.Options()
	if options == nil {
		return ""
	}
	if v, ok := proto.GetExtension(options, extv1.E_ServiceComment).(string); ok {
		return v
	}
	return ""
}

// GetMethodComment extracts the method_comment extension from a method descriptor.
func GetMethodComment(md protoreflect.MethodDescriptor) string {
	options := md.Options()
	if options == nil {
		return ""
	}
	if v, ok := proto.GetExtension(options, extv1.E_MethodComment).(string); ok {
		return v
	}
	return ""
}
