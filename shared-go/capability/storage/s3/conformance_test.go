package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"
)

// Live conformance probe against a real S3-compatible backend; skipped unless
// S3_CONFORMANCE_ENDPOINT is set. One run produces a drift profile: each
// subtest probes one standard-S3 surface this capability targets, and a
// failure names the drift plus what breaks downstream. Drift is absorbed in
// config switches or platform instantiation — never inside this package.
//
//	S3_CONFORMANCE_ENDPOINT=http://localhost:8333 \
//	S3_CONFORMANCE_ACCESS_KEY=admin@cyber-ecosystem.com \
//	S3_CONFORMANCE_SECRET_KEY=Cyber-Ecosystem123 \
//	S3_CONFORMANCE_BUCKET=core \
//	go test ./shared-go/capability/storage/s3/ -run TestLiveConformance -count=1 -v
func TestLiveConformance(t *testing.T) {
	endpoint := os.Getenv("S3_CONFORMANCE_ENDPOINT")
	if endpoint == "" {
		t.Skip("S3_CONFORMANCE_ENDPOINT not set; live conformance is opt-in")
	}
	// Limits pinned to config.yaml's values (zeros would silently resolve to
	// compiled defaults and diverge this instantiation from the service's).
	cfg := &Config{
		Endpoint:           endpoint,
		AccessKey:          os.Getenv("S3_CONFORMANCE_ACCESS_KEY"),
		SecretKey:          os.Getenv("S3_CONFORMANCE_SECRET_KEY"),
		Bucket:             os.Getenv("S3_CONFORMANCE_BUCKET"),
		Region:             "us-east-1",
		UsePathStyle:       true,
		MaxFileSize:        52428800, // 50 MiB
		MultipartThreshold: 8388608,  // 8 MiB
		PartSize:           5242880,  // 5 MiB
		PresignTTL:         15 * time.Minute,
	}
	if cfg.Bucket == "" {
		cfg.Bucket = "core"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, _, cleanup, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer cleanup()
	s := New(client, nil, *cfg)

	suffix := time.Now().UnixNano()
	body := fmt.Sprintf("conformance probe %d — standard S3 surface", suffix)
	key := fmt.Sprintf("conformance/probe-%d.txt", suffix)

	t.Run("presign-put-bare", func(t *testing.T) {
		// Plain net/http PUT, no SDK machinery or checksum header — the browser
		// shape. Failure = checksum/signature drift; direct uploads break.
		url, err := s.Presign.PresignUpload(ctx, key, int64(len(body)), time.Minute)
		if err != nil {
			t.Fatalf("PresignUpload: %v", err)
		}
		put, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		put.Header.Set("Content-Type", "text/plain; charset=utf-8")
		put.ContentLength = int64(len(body))
		resp, err := http.DefaultClient.Do(put)
		if err != nil {
			t.Fatalf("bare PUT: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("bare PUT status = %d, want 200 — checksum/signature drift, direct upload broken", resp.StatusCode)
		}
	})
	defer func() { _ = s.Object.Delete(ctx, key) }()

	t.Run("content-type-passthrough", func(t *testing.T) {
		// Unsigned Content-Type must still be stored; failure = Confirm-time
		// metadata backfill breaks (G4 closure).
		obj, err := s.Object.Stat(ctx, key)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if !strings.HasPrefix(obj.ContentType, "text/plain") {
			t.Fatalf("stored Content-Type = %q, want text/plain — passthrough drift, preview/confirm break", obj.ContentType)
		}
	})

	t.Run("head-metadata", func(t *testing.T) {
		obj, err := s.Object.Stat(ctx, key)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if obj.Size != int64(len(body)) {
			t.Errorf("Size = %d, want %d", obj.Size, len(body))
		}
		if obj.ETag == "" {
			t.Errorf("ETag empty — R2 seam (integrity对照) loses its data source")
		}
	})

	t.Run("get-response-overrides", func(t *testing.T) {
		// response-* overrides signed into a presigned GET; failure = download
		// filename/inline control breaks (G3).
		disposition := `attachment; filename="probe.txt"`
		url, err := s.Presign.PresignDownload(ctx, key, time.Minute, storage.DownloadOptions{
			ContentType:        "text/plain; charset=utf-8",
			ContentDisposition: disposition,
		})
		if err != nil {
			t.Fatalf("PresignDownload: %v", err)
		}
		get, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(get)
		if err != nil {
			t.Fatalf("bare GET: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("bare GET status = %d, want 200", resp.StatusCode)
		}
		got, _ := io.ReadAll(resp.Body)
		if string(got) != body {
			t.Errorf("body mismatch: %q", got)
		}
		if cd := resp.Header.Get("Content-Disposition"); cd != disposition {
			t.Errorf("Content-Disposition = %q, want %q — override drift, download naming breaks", cd, disposition)
		}
	})

	t.Run("range-get", func(t *testing.T) {
		// S3-native Range on a presigned GET; failure = streaming/video drag breaks.
		url, err := s.Presign.PresignDownload(ctx, key, time.Minute, storage.DownloadOptions{})
		if err != nil {
			t.Fatalf("PresignDownload: %v", err)
		}
		get, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		get.Header.Set("Range", "bytes=0-3")
		resp, err := http.DefaultClient.Do(get)
		if err != nil {
			t.Fatalf("ranged GET: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusPartialContent {
			t.Fatalf("ranged GET status = %d, want 206 — Range drift, media streaming breaks", resp.StatusCode)
		}
		got, _ := io.ReadAll(resp.Body)
		if len(got) != 4 || string(got) != body[:4] {
			t.Errorf("range slice = %q, want first 4 bytes %q", got, body[:4])
		}
	})

	t.Run("error-surface", func(t *testing.T) {
		// A missing key must surface as 404 with the standard NoSuchKey code;
		// nonstandard codes defeat mapError and surface as unknown errors.
		url, err := s.Presign.PresignDownload(ctx, fmt.Sprintf("conformance/absent-%d.txt", suffix), time.Minute, storage.DownloadOptions{})
		if err != nil {
			t.Fatalf("PresignDownload: %v", err)
		}
		get, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(get)
		if err != nil {
			t.Fatalf("GET absent: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("absent key status = %d, want 404", resp.StatusCode)
		}
		errBody, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(errBody), "NoSuchKey") {
			t.Errorf("error code not NoSuchKey — drift in error surface, mapError loses it: %q", errBody)
		}
	})

	t.Run("delete-idempotent", func(t *testing.T) {
		// S3 deletes an absent key successfully; a 404 here makes Delete report
		// ErrNotFound on retries.
		if err := s.Object.Delete(ctx, fmt.Sprintf("conformance/absent-%d.txt", suffix)); err != nil {
			t.Errorf("delete absent key: %v — want idempotent success", err)
		}
	})

	t.Run("multipart-lifecycle", func(t *testing.T) {
		// Presigned part URLs + bare part PUTs + ListParts shape + Complete —
		// the resumable-upload foundation. Drift here breaks large uploads.
		mKey := fmt.Sprintf("conformance/mp-%d.bin", suffix)
		uploadID, err := s.Multipart.Create(ctx, mKey, "application/octet-stream")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		part1 := strings.Repeat("a", 5<<20+8) // first part must exceed the 5MiB minimum
		part2 := "tail"
		var parts []storage.CompletedPart
		for i, chunk := range []string{part1, part2} {
			url, err := s.Presign.PresignUploadPart(ctx, mKey, uploadID, int32(i+1), time.Minute)
			if err != nil {
				t.Fatalf("PresignUploadPart(%d): %v", i+1, err)
			}
			put, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(chunk))
			put.ContentLength = int64(len(chunk))
			resp, err := http.DefaultClient.Do(put)
			if err != nil {
				t.Fatalf("bare part PUT(%d): %v", i+1, err)
			}
			etag := resp.Header.Get("ETag")
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("bare part PUT(%d) status = %d, want 200 — multipart drift", i+1, resp.StatusCode)
			}
			if etag == "" {
				t.Fatalf("part PUT(%d) returned no ETag — Complete cannot assemble", i+1)
			}
			parts = append(parts, storage.CompletedPart{PartNumber: int32(i + 1), ETag: etag})
		}

		listed, err := s.Multipart.ListParts(ctx, mKey, uploadID)
		if err != nil {
			t.Fatalf("ListParts: %v", err)
		}
		if len(listed) != 2 {
			t.Errorf("ListParts = %d parts, want 2 — resume diff breaks", len(listed))
		}
		for _, p := range listed {
			if p.ETag == "" || p.Size <= 0 {
				t.Errorf("listed part %+v missing ETag/Size — resume progress display breaks", p)
			}
		}

		info, err := s.Multipart.Complete(ctx, mKey, uploadID, parts)
		if err != nil {
			t.Fatalf("Complete: %v", err)
		}
		// CompleteMultipartUpload's standard response carries only
		// Location/Bucket/Key/ETag — size authority is the post-complete HEAD,
		// i.e. exactly the Confirm closure path.
		if info == nil || info.ETag == "" {
			t.Fatalf("Complete returned no ETag: %+v", info)
		}
		st, err := s.Object.Stat(ctx, mKey)
		if err != nil {
			t.Fatalf("Stat after complete: %v", err)
		}
		if st.Size != int64(len(part1)+len(part2)) {
			t.Errorf("assembled size = %d, want %d — part data lost", st.Size, len(part1)+len(part2))
		}
		defer func() { _ = s.Object.Delete(ctx, mKey) }()
	})

	t.Run("copy-object", func(t *testing.T) {
		dst := fmt.Sprintf("conformance/copy-%d.txt", suffix)
		defer func() { _ = s.Object.Delete(ctx, dst) }()
		if _, err := s.Object.Copy(ctx, key, dst); err != nil {
			t.Fatalf("Copy: %v — server-side copy drift", err)
		}
		obj, err := s.Object.Stat(ctx, dst)
		if err != nil {
			t.Fatalf("Stat copy: %v", err)
		}
		if obj.Size != int64(len(body)) {
			t.Errorf("copy size = %d, want %d", obj.Size, len(body))
		}
	})
}
