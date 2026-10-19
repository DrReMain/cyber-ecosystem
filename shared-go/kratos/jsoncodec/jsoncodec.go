// Package jsoncodec provides a "json" codec built on protojson that emits
// unpopulated fields. kratos's default encoding/json codec is Go std json,
// which honors the proto struct tags' omitempty and drops zero-value scalar
// fields — so HTTP/Connect JSON responses omit count=0, status=0, etc.
//
// Codec's methods (Name/Marshal/Unmarshal) satisfy both kratos
// encoding.Codec and connect.Codec (identical signatures), so one value serves
// the "json" codec for HTTP (via Register) and for Connect (via
// connect.WithCodec). proto messages serialize with protojson; non-proto values
// fall back to encoding/json.
package jsoncodec

import (
	jsonstd "encoding/json"

	kencoding "github.com/go-kratos/kratos/v3/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// EmitUnpopulated outputs zero-value scalar fields instead of omitting them.
// UseProtoNames=false keeps camelCase field names (the proto JSON standard);
// enums render as their string names.
var (
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false,
	}
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

// Codec is the protojson-based "json" codec. Register (HTTP) or pass via
// connect.WithCodec (Connect). The zero value is the "json" codec; use NewCodec
// to also cover the "json; charset=utf-8" Content-Type variant for connect —
// connect-go registers both names separately, so both must be overridden or
// clients sending the charset variant silently fall back to connect-go's
// default codec (which omits empty fields).
type Codec struct{ name string }

// NewCodec returns a Codec whose Name is name. Use "json; charset=utf-8" so
// connect handles that Content-Type variant with the same protojson behavior.
func NewCodec(name string) Codec { return Codec{name: name} }

func (c Codec) Name() string {
	if c.name != "" {
		return c.name
	}
	return "json"
}

func (Codec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return MarshalOptions.Marshal(msg)
	}
	return jsonstd.Marshal(v)
}

func (Codec) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	if msg, ok := v.(proto.Message); ok {
		return UnmarshalOptions.Unmarshal(data, msg)
	}
	return jsonstd.Unmarshal(data, v)
}

// Register overrides kratos's default "json" codec with Codec, so HTTP JSON
// responses include unpopulated fields. Call once at startup (e.g. in main
// init()). Connect uses connect.WithCodec(Codec{}) separately.
func Register() {
	kencoding.RegisterCodec(Codec{})
}
