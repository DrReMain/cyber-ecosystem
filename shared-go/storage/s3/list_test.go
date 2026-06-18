package s3

import (
	"bytes"
	"context"
	"testing"
)

func TestListPagination(t *testing.T) {
	s, cleanup := newTestStorage(t)
	defer cleanup()
	ctx := context.Background()
	keys := []string{"list/a", "list/b", "list/c", "other/x"}
	resetObjects(t, s, keys...)

	for _, k := range keys {
		if _, err := s.Object.Upload(ctx, k, bytes.NewReader([]byte("v")), 1, "text/plain"); err != nil {
			t.Fatal(err)
		}
	}

	// prefix filter + small page
	page, err := s.List.List(ctx, "list/", "", 2)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Objects) != 2 {
		t.Fatalf("page1 size: got %d want 2", len(page.Objects))
	}
	if page.NextCursor == "" {
		t.Fatal("want NextCursor for remaining items")
	}
	page2, err := s.List.List(ctx, "list/", page.NextCursor, 100)
	if err != nil {
		t.Fatalf("List page2: %v", err)
	}
	if len(page2.Objects) != 1 {
		t.Fatalf("page2 size: got %d want 1", len(page2.Objects))
	}
	if page2.NextCursor != "" {
		t.Fatalf("page2 should be last, got cursor %q", page2.NextCursor)
	}
}
