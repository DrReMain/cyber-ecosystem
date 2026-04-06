package connect

import (
	connectrpc "connectrpc.com/connect"

	"github.com/go-kratos/kratos/v2/encoding"
)

// JSONCodec returns a Connect JSON codec that uses the globally registered codec.
// This ensures consistent serialization behavior with HTTP transport.
//
// The codec is obtained via encoding.GetCodec("json"), which returns the
// custom protoJSONCodec registered during json.Init().
func JSONCodec() connectrpc.Codec {
	return &connectCodecAdapter{codec: encoding.GetCodec("json")}
}

// connectCodecAdapter adapts Kratos encoding.Codec to Connect Codec interface.
type connectCodecAdapter struct {
	codec encoding.Codec
}

// Name returns the codec name.
func (c *connectCodecAdapter) Name() string {
	return "json"
}

// Marshal serializes a message using the global codec.
func (c *connectCodecAdapter) Marshal(message any) ([]byte, error) {
	return c.codec.Marshal(message)
}

// Unmarshal deserializes data into a message using the global codec.
func (c *connectCodecAdapter) Unmarshal(data []byte, message any) error {
	return c.codec.Unmarshal(data, message)
}
