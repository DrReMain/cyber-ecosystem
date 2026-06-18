package s3

import (
	"context"
	"os"
	"testing"

	"cyber-ecosystem/shared-go/storage"
)

// testConfig points at the test MinIO (MINIO_ENDPOINT env, or localhost:30090).
func testConfig() Config {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:9000" // k3d LB maps host:9000 → MinIO NodePort 30090
	}
	return Config{
		Endpoint:     endpoint,
		AccessKey:    "admin@cyber-ecosystem.com",
		SecretKey:    "Cyber-Ecosystem123",
		Bucket:       "storage-test",
		Region:       "us-east-1",
		UsePathStyle: true,
		MaxFileSize:  5 << 20, // 5 MiB for tests
	}
}

// newTestStorage builds a wired *storage.Storage against the test MinIO,
// skipping if MinIO is unavailable.
func newTestStorage(t *testing.T) (*storage.Storage, func()) {
	t.Helper()
	cfg := testConfig()
	client, cleanup, err := NewClient(&cfg)
	if err != nil {
		t.Skipf("minio unavailable: %v", err)
	}
	return New(client, cfg), cleanup
}

// resetObjects deletes the given keys immediately and on cleanup, keeping the
// shared bucket hermetic across repeated runs.
func resetObjects(t *testing.T, s *storage.Storage, keys ...string) {
	t.Helper()
	ctx := context.Background()
	del := func() {
		for _, k := range keys {
			_ = s.Object.Delete(ctx, k)
		}
	}
	del()
	t.Cleanup(del)
}
