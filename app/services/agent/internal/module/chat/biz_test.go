package chat

import (
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/app/services/agent/internal/conf"
)

// The master-key floor for the sealing key's consumer: HKDF derivation
// cannot add entropy, so a below-floor key silently weakens every sealed
// API key; construction must refuse it. The floor matches the other
// consumer (agentconfig) so one key passes both constructors or neither.
func TestNewChatUCMasterKeyFloor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctor := func(key string) error {
		_, err := NewChatUC(logger, nil, &conf.Crypto{MasterKey: key}, nil, nil)
		return err
	}
	for _, key := range []string{"", "0123456789abcde"} { // empty; 15 bytes, one short
		if err := ctor(key); err == nil {
			t.Fatalf("master key %q must fail construction", key)
		}
	}
	if _, err := NewChatUC(nil, nil, nil, nil, nil); err == nil {
		t.Fatal("nil crypto config must fail construction")
	}
	if err := ctor("0123456789abcdef"); err != nil { // 16 bytes, the floor
		t.Fatalf("16-byte master key must construct, got: %v", err)
	}
}
