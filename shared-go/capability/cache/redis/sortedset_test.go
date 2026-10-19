package redis

import (
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"

	"cyber-ecosystem/shared-go/capability/cache"
)

func TestSortedSetConversion(t *testing.T) {
	in := []cache.Member{
		{Score: 1.5, Member: "a"},
		{Score: -0.25, Member: "b"},
	}
	z := toZ(in)
	if len(z) != 2 || z[0].Score != 1.5 || z[0].Member.(string) != "a" {
		t.Errorf("toZ = %+v", z)
	}
	out := fromZ(z)
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip = %+v, want %+v", out, in)
	}
	// Non-string members (go-redis types them as string for the read commands
	// used here, but stay panic-safe) decode to zero Members, not a panic.
	if got := fromZ([]redis.Z{{Score: 1, Member: 42}}); len(got) != 1 || got[0].Member != "" {
		t.Errorf("non-string member = %+v, want zero Member", got)
	}
}
