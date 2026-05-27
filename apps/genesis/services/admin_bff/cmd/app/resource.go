package main

import (
	sourcesdk "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

func newResource() *sourcesdk.Resource {
	return sourcesdk.NewSchemaless(
		semconv.ServiceNameKey.String(Name),
		semconv.ServiceInstanceIDKey.String(id),
		semconv.ServiceVersionKey.String(Version),
	)
}
