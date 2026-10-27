package population

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

// fakePopulationRP answers the registry slice: a passthrough list.
type fakePopulationRP struct {
	PopulationRP
	listOut *PopulationListOut
}

func (f *fakePopulationRP) List(context.Context, *PopulationListIn) (*PopulationListOut, error) {
	return f.listOut, nil
}

func newPopulationUC(rp *fakePopulationRP) *PopulationUC {
	return NewPopulationUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, rp,
	)
}

// List delegates — the registry carries no list-side law of its own.
func TestListDelegates(t *testing.T) {
	want := &PopulationListOut{List: []*Population{{Code: "user"}}}
	rp := &fakePopulationRP{listOut: want}
	got, err := newPopulationUC(rp).List(context.Background(), &PopulationListIn{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got != want {
		t.Fatalf("list: out = %v, want passthrough", got)
	}
}
