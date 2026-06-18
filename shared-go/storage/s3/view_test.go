package s3

import (
	"bytes"
	"context"
	"io"
	"testing"
)

// TestForViewBucketIsolation proves For(bucket) scopes ops to that bucket: an
// object written via the tenant view is invisible to the default bucket's list,
// and vice versa.
func TestForViewBucketIsolation(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	const tenant = "tenant-view-test"

	_ = s.Bucket.Delete(ctx, tenant)
	if err := s.Bucket.Create(ctx, tenant); err != nil {
		t.Fatalf("Create tenant bucket: %v", err)
	}
	v := s.For(tenant)
	if v == nil {
		t.Fatal("For returned nil")
	}
	t.Cleanup(func() {
		_ = v.Object.Delete(ctx, "tenant/iso")
		_ = s.Object.Delete(ctx, "default/iso")
		_ = s.Bucket.Delete(ctx, tenant)
	})

	if _, err := v.Object.Upload(ctx, "tenant/iso", bytes.NewReader([]byte("in-tenant")), 9, "text/plain"); err != nil {
		t.Fatalf("tenant Upload: %v", err)
	}
	if _, err := s.Object.Upload(ctx, "default/iso", bytes.NewReader([]byte("in-default")), 10, "text/plain"); err != nil {
		t.Fatalf("default Upload: %v", err)
	}

	// default bucket must NOT see the tenant key
	def, err := s.List.List(ctx, "tenant/", "", 100)
	if err != nil {
		t.Fatalf("default List: %v", err)
	}
	if len(def.Objects) != 0 {
		t.Errorf("default bucket leaked tenant keys: got %d", len(def.Objects))
	}

	// tenant view sees only its own object
	ten, err := v.List.List(ctx, "", "", 100)
	if err != nil {
		t.Fatalf("tenant List: %v", err)
	}
	if len(ten.Objects) != 1 || ten.Objects[0].Key != "tenant/iso" {
		t.Errorf("tenant bucket list: got %+v, want only tenant/iso", ten.Objects)
	}

	// round-trip via the tenant view
	rc, _, err := v.Object.Download(ctx, "tenant/iso")
	if err != nil {
		t.Fatalf("tenant Download: %v", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(got) != "in-tenant" {
		t.Errorf("tenant download body: got %q want in-tenant", got)
	}
}
