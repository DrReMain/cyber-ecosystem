package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/storage"
)

// newTestStorageCfg is newTestStorage with a config mutation hook (e.g. to lift
// MaxFileSize for big-file tests).
func newTestStorageCfg(t *testing.T, mutate func(*Config)) (*storage.Storage, func()) {
	t.Helper()
	cfg := testConfig()
	if mutate != nil {
		mutate(&cfg)
	}
	client, cleanup, err := NewClient(&cfg)
	if err != nil {
		t.Skipf("minio unavailable: %v", err)
	}
	return New(client, cfg), cleanup
}

// TestUploadLargeFileAutoMultipart: 12 MiB (> 8 MiB threshold) uploads via
// transfermanager's automatic multipart; round-trips intact with correct Size.
func TestUploadLargeFileAutoMultipart(t *testing.T) {
	s, cleanup := newTestStorageCfg(t, func(c *Config) { c.MaxFileSize = 0 })
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "big/auto-mp")

	const size = 12 << 20 // 12 MiB
	body := bytes.Repeat([]byte("Z"), size)
	info, err := s.Object.Upload(ctx, "big/auto-mp", bytes.NewReader(body), int64(size), "application/octet-stream")
	if err != nil {
		t.Fatalf("Upload %dMiB: %v", size>>20, err)
	}
	if info.Size != int64(size) {
		t.Errorf("info.Size: got %d want %d", info.Size, size)
	}
	rc, dinfo, err := s.Object.Download(ctx, "big/auto-mp")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if len(got) != size {
		t.Fatalf("download len: got %d want %d", len(got), size)
	}
	if dinfo.Size != int64(size) {
		t.Errorf("dinfo.Size: got %d want %d", dinfo.Size, size)
	}
	if !bytes.Equal(got, body) {
		t.Error("content mismatch on big round-trip")
	}
}

// TestMultipartManyParts: explicit Create → 3×5MiB + 1 tail → ListParts(4) →
// Complete; assembled size verified via Stat.
func TestMultipartManyParts(t *testing.T) {
	s, cleanup := newTestStorageCfg(t, func(c *Config) { c.MaxFileSize = 0 })
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "mp/many")

	uploadID, err := s.Multipart.Create(ctx, "mp/many", "application/octet-stream")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	part := bytes.Repeat([]byte("P"), 5<<20) // 5 MiB (S3 minimum non-last part)
	parts := make([]storage.CompletedPart, 0, 4)
	for i := 1; i <= 3; i++ {
		etag, err := s.Multipart.UploadPart(ctx, "mp/many", uploadID, int32(i), bytes.NewReader(part), int64(len(part)))
		if err != nil {
			t.Fatalf("UploadPart %d: %v", i, err)
		}
		parts = append(parts, storage.CompletedPart{PartNumber: int32(i), ETag: etag})
	}
	etag, err := s.Multipart.UploadPart(ctx, "mp/many", uploadID, 4, bytes.NewReader([]byte("tail")), 4)
	if err != nil {
		t.Fatalf("UploadPart tail: %v", err)
	}
	parts = append(parts, storage.CompletedPart{PartNumber: 4, ETag: etag})

	lp, err := s.Multipart.ListParts(ctx, "mp/many", uploadID)
	if err != nil {
		t.Fatalf("ListParts: %v", err)
	}
	if len(lp) != 4 {
		t.Fatalf("ListParts: got %d want 4", len(lp))
	}
	if _, err := s.Multipart.Complete(ctx, "mp/many", uploadID, parts); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	st, err := s.Object.Stat(ctx, "mp/many")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	want := int64(3*(5<<20) + 4)
	if st.Size != want {
		t.Errorf("assembled size: got %d want %d", st.Size, want)
	}
}

// TestListPaginationManyObjects: 25 objects under a prefix, paged at 10 — all
// retrieved via continuation cursor.
func TestListPaginationManyObjects(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	keys := make([]string, 25)
	for i := range keys {
		keys[i] = fmt.Sprintf("page/obj-%02d", i)
	}
	resetObjects(t, s, keys...)
	for _, k := range keys {
		if _, err := s.Object.Upload(ctx, k, bytes.NewReader([]byte("x")), 1, "text/plain"); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	cursor := ""
	pages := 0
	for {
		page, err := s.List.List(ctx, "page/", cursor, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		pages++
		for _, o := range page.Objects {
			seen[o.Key] = true
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
	}
	if len(seen) != 25 {
		t.Errorf("paginated count: got %d want 25", len(seen))
	}
}

// TestPresignBigFileRoundTrip: 6 MiB upload+download via presigned URLs.
func TestPresignBigFileRoundTrip(t *testing.T) {
	s, cleanup := newTestStorageCfg(t, func(c *Config) { c.MaxFileSize = 0 })
	defer cleanup()
	ctx := context.Background()
	resetObjects(t, s, "presign/big")

	const size = 6 << 20
	body := bytes.Repeat([]byte("Q"), size)
	upURL, err := s.Presign.PresignUpload(ctx, "presign/big", int64(size), time.Minute)
	if err != nil {
		t.Fatalf("PresignUpload: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upURL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT presigned: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("PUT status: %d", resp.StatusCode)
	}
	downURL, err := s.Presign.PresignDownload(ctx, "presign/big", time.Minute)
	if err != nil {
		t.Fatalf("PresignDownload: %v", err)
	}
	resp, err = http.DefaultClient.Get(downURL)
	if err != nil {
		t.Fatalf("GET presigned: %v", err)
	}
	got, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if len(got) != size {
		t.Errorf("big presign download len: got %d want %d", len(got), size)
	}
}

// TestForViewMultiTenantWithData: two tenant buckets via For(); each sees only
// its own object (same key, different content), and Bucket.List sees both.
func TestForViewMultiTenantWithData(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	const ta, tb = "tenant-a-data", "tenant-b-data"
	for _, bk := range []string{ta, tb} {
		_ = s.Bucket.Delete(ctx, bk)
		if err := s.Bucket.Create(ctx, bk); err != nil {
			t.Fatalf("Create %s: %v", bk, err)
		}
	}
	va, vb := s.For(ta), s.For(tb)
	t.Cleanup(func() {
		_ = va.Object.Delete(ctx, "x")
		_ = vb.Object.Delete(ctx, "x")
		_ = s.Bucket.Delete(ctx, ta)
		_ = s.Bucket.Delete(ctx, tb)
	})
	if _, err := va.Object.Upload(ctx, "x", bytes.NewReader([]byte("A")), 1, "text/plain"); err != nil {
		t.Fatal(err)
	}
	if _, err := vb.Object.Upload(ctx, "x", bytes.NewReader([]byte("B")), 1, "text/plain"); err != nil {
		t.Fatal(err)
	}
	// each tenant's list sees exactly its own object
	la, err := va.List.List(ctx, "", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(la.Objects) != 1 || la.Objects[0].Key != "x" {
		t.Errorf("tenant A list: %+v", la.Objects)
	}
	// content differs per tenant (same key, isolated buckets)
	rca, _, err := va.Object.Download(ctx, "x")
	if err != nil {
		t.Fatal(err)
	}
	gotA, _ := io.ReadAll(rca)
	_ = rca.Close()
	rcb, _, err := vb.Object.Download(ctx, "x")
	if err != nil {
		t.Fatal(err)
	}
	gotB, _ := io.ReadAll(rcb)
	_ = rcb.Close()
	if string(gotA) != "A" || string(gotB) != "B" {
		t.Errorf("tenant content isolation: A=%q B=%q", gotA, gotB)
	}
	// bucket List includes both tenants
	bs, err := s.Bucket.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, bk := range bs {
		have[bk.Name] = true
	}
	if !have[ta] || !have[tb] {
		t.Errorf("bucket List missing tenants: %v", have)
	}
}
