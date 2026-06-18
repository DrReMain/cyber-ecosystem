package s3

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestPresignRoundTrip does a real PUT then GET against the presigned URLs to
// prove the signatures are valid against MinIO (not just that a URL is returned).
func TestPresignRoundTrip(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "presign/rt")

	upURL, err := s.Presign.PresignUpload(ctx, "presign/rt", 6, time.Minute)
	if err != nil {
		t.Fatalf("PresignUpload: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upURL, bytes.NewReader([]byte("signed")))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	// No Content-Type header: presign no longer binds it, so any format works.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT presigned: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("PUT status: %d", resp.StatusCode)
	}

	downURL, err := s.Presign.PresignDownload(ctx, "presign/rt", time.Minute)
	if err != nil {
		t.Fatalf("PresignDownload: %v", err)
	}
	resp, err = http.DefaultClient.Get(downURL)
	if err != nil {
		t.Fatalf("GET presigned: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	got, _ := io.ReadAll(resp.Body)
	if string(got) != "signed" {
		t.Errorf("presign download body: got %q want signed", got)
	}
}
