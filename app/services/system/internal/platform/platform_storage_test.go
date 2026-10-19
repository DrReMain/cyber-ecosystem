package platform

import (
	"context"
	"errors"
	"testing"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"
)

// The routing fakes: the scope factory records every bucket asked for, the
// bucket manager records its provisioning calls.
type routingBucket struct {
	created []string
	fail    error
}

func (b *routingBucket) Create(_ context.Context, bucket string) error {
	if b.fail != nil {
		return b.fail
	}
	b.created = append(b.created, bucket)
	return nil
}
func (b *routingBucket) Exists(context.Context, string) (bool, error) { return false, nil }
func (b *routingBucket) Delete(context.Context, string) error         { return nil }
func (b *routingBucket) List(context.Context) ([]storage.BucketInfo, error) {
	return nil, nil
}

func routingStorage(b *routingBucket) (*TenantStorage, *[]string) {
	var scoped []string
	s := storage.NewStorage(storage.View{}, b, func(name string) *storage.View {
		scoped = append(scoped, name)
		return &storage.View{}
	}, storage.Limits{})
	return &TenantStorage{Storage: s, defaultBucket: "core"}, &scoped
}

// The platform tenant rides the configured default bucket; every other
// domain rides the suffixed bucket, and the explicit switch wins over the
// subject — the mixin's precedence, so rows and buckets route one axis.
func TestForTenantRouting(t *testing.T) {
	b := &routingBucket{}
	ts, scoped := routingStorage(b)

	if got := ts.ForTenant(context.Background()); got != &ts.View {
		t.Fatal("no tenant: want the default bucket view")
	}
	if got := ts.ForTenant(security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "default"})); got != &ts.View {
		t.Fatal("platform seat: want the default bucket view")
	}

	tenantCtx := security.WithTenant(context.Background(), "acme")
	if got := ts.ForTenant(tenantCtx); got == nil || len(*scoped) != 1 || (*scoped)[0] != "core-acme" {
		t.Fatalf("explicit domain: scoped = %v, want core-acme", *scoped)
	}
	subjectCtx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "acme"})
	if got := ts.ForTenant(subjectCtx); got == nil || len(*scoped) != 2 || (*scoped)[1] != "core-acme" {
		t.Fatalf("subject domain: scoped = %v, want core-acme twice", *scoped)
	}
	both := security.WithTenant(subjectCtx, "acme")
	if got := ts.ForTenant(both); got == nil || len(*scoped) != 3 || (*scoped)[2] != "core-acme" {
		t.Fatalf("explicit over subject: scoped = %v, want core-acme", *scoped)
	}
	// Provisioning rides the first route per bucket, once per process.
	if len(b.created) != 1 || b.created[0] != "core-acme" {
		t.Fatalf("provisioning = %v, want core-acme once", b.created)
	}
}

// A failed provisioning un-marks the bucket so the next route retries.
func TestForTenantProvisioningRetriesAfterFailure(t *testing.T) {
	b := &routingBucket{fail: errors.New("provisioning denied")}
	ts, scoped := routingStorage(b)

	_ = ts.ForTenant(security.WithTenant(context.Background(), "acme"))
	if len(*scoped) != 1 {
		t.Fatalf("first route: scoped = %v", *scoped)
	}
	b.fail = nil
	_ = ts.ForTenant(security.WithTenant(context.Background(), "acme"))
	if len(*scoped) != 2 {
		t.Fatalf("retry route: scoped = %v, want a second attempt", *scoped)
	}
}
