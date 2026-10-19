package s3

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"part-size below s3 minimum", Config{PartSize: 4 << 20}, true},
		{"part-size at minimum", Config{PartSize: minPartSize}, false},
		{"zero part-size skips check", Config{}, false},
	}
	for _, c := range cases {
		if err := c.cfg.validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: validate() err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestSizeOrDefaultHelpers(t *testing.T) {
	if got := partSizeOrDefault(0); got != defaultPartSize {
		t.Errorf("partSizeOrDefault(0) = %d, want %d", got, defaultPartSize)
	}
	if got := partSizeOrDefault(1 << 20); got != 1<<20 {
		t.Errorf("partSizeOrDefault(explicit) = %d, want passthrough", got)
	}
	if got := multipartThresholdOrDefault(0); got != defaultMultipartThreshold {
		t.Errorf("multipartThresholdOrDefault(0) = %d, want %d", got, defaultMultipartThreshold)
	}
	if got := multipartThresholdOrDefault(1 << 20); got != 1<<20 {
		t.Errorf("multipartThresholdOrDefault(explicit) = %d, want passthrough", got)
	}
}

func TestPresignTTLOrDefault(t *testing.T) {
	if got := presignTTLOrDefault(0); got != defaultPresignTTL {
		t.Errorf("zero: got %v, want %v", got, defaultPresignTTL)
	}
	if got := presignTTLOrDefault(48 * time.Hour); got != 48*time.Hour {
		t.Errorf("explicit: got %v, want passthrough", got)
	}
	if got := presignTTLOrDefault(30 * 24 * time.Hour); got != maxPresignTTL {
		t.Errorf("over cap: got %v, want clamp to %v", got, maxPresignTTL)
	}
}
