package agentconfig

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	"cyber-ecosystem/app/services/agent/internal/conf"
)

// passthroughTm runs the unit inline — the upsert's create-or-rewrite
// branch is the law under test, not the store's atomicity.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeAgentConfigRP answers the sealed-credential store with in-memory
// rows keyed by user and records every write; fakeUserRP answers the
// hydration read with a fixed brief set.
type fakeAgentConfigRP struct {
	AgentConfigRP
	rows map[string]*AgentConfig

	created    []*AgentConfig
	createdKey []string
	updated    []*AgentConfig
	updatedKey []*string
	deleted    []string
	listOut    *ListOut
}

func (f *fakeAgentConfigRP) FindSealedByUserID(_ context.Context, userID string) (*AgentConfig, string, error) {
	if mc, ok := f.rows[userID]; ok {
		return mc, "sealed-old", nil
	}
	return nil, "", nil
}

func (f *fakeAgentConfigRP) Create(_ context.Context, mc *AgentConfig, apiKeySealed string) (*AgentConfig, error) {
	f.created = append(f.created, mc)
	f.createdKey = append(f.createdKey, apiKeySealed)
	mc.ID = "mc-new"
	f.rows[mc.UserID] = mc
	return mc, nil
}

func (f *fakeAgentConfigRP) Update(_ context.Context, mc *AgentConfig, apiKeySealed *string) error {
	f.updated = append(f.updated, mc)
	f.updatedKey = append(f.updatedKey, apiKeySealed)
	f.rows[mc.UserID] = mc
	return nil
}

func (f *fakeAgentConfigRP) DeleteByUserID(_ context.Context, userID string) error {
	f.deleted = append(f.deleted, userID)
	delete(f.rows, userID)
	return nil
}

func (f *fakeAgentConfigRP) List(context.Context, *ListIn) (*ListOut, error) {
	return f.listOut, nil
}

type fakeUserRP struct {
	UserRP
	briefs   map[string]*UserBrief
	hydrated [][]string
}

func (f *fakeUserRP) Hydrate(_ context.Context, ids []string) map[string]*UserBrief {
	f.hydrated = append(f.hydrated, ids)
	return f.briefs
}

func newAgentConfigUC(rp *fakeAgentConfigRP, urp *fakeUserRP) *AgentConfigUC {
	uc, err := NewAgentConfigUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{},
		&conf.Crypto{MasterKey: "0123456789abcdef"},
		rp, urp,
	)
	if err != nil {
		panic(err)
	}
	return uc
}

func cfgSubject(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID, TenantID: "t1"})
}

// The master-key floor mirrors the chat consumer's own: HKDF derivation
// cannot add entropy, so a below-floor key silently weakens every sealed
// API key — construction refuses it.
func TestNewAgentConfigUCMasterKeyFloor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, key := range []string{"", "0123456789abcde"} { // empty; 15 bytes, one short
		if _, err := NewAgentConfigUC(logger, nil, &conf.Crypto{MasterKey: key}, nil, nil); err == nil {
			t.Fatalf("master key %q must fail construction", key)
		}
	}
	if _, err := NewAgentConfigUC(logger, nil, nil, nil, nil); err == nil {
		t.Fatal("nil crypto config must fail construction")
	}
}

// The upsert's absent half: one live row per user, so absence creates —
// and a supplied API key lands sealed, never as the caller's plaintext.
func TestUpdateCreatesWhenAbsent(t *testing.T) {
	rp := &fakeAgentConfigRP{rows: map[string]*AgentConfig{}}
	uc := newAgentConfigUC(rp, &fakeUserRP{})

	if err := uc.Update(cfgSubject("u1"), &UpdateIn{BaseURL: "https://api.example", APIKey: strPtr("sk-secret")}); err != nil {
		t.Fatalf("update create: %v", err)
	}
	if len(rp.created) != 1 || rp.created[0].UserID != "u1" || rp.created[0].BaseURL != "https://api.example" {
		t.Fatalf("created: %+v", rp.created)
	}
	if sealed := rp.createdKey[0]; sealed == "" || sealed == "sk-secret" || strings.Contains(sealed, "sk-secret") {
		t.Fatalf("created key must be a non-empty seal of the plaintext, got %q", sealed)
	}
}

// The upsert's present half rewrites the row; a nil API key means keep
// the stored ciphertext — the repo receives a nil seal, not an empty one.
func TestUpdateRewritesPresentAndKeepsKey(t *testing.T) {
	rp := &fakeAgentConfigRP{rows: map[string]*AgentConfig{
		"u1": {ID: "mc-1", UserID: "u1", BaseURL: "https://old.example"},
	}}
	uc := newAgentConfigUC(rp, &fakeUserRP{})

	if err := uc.Update(cfgSubject("u1"), &UpdateIn{BaseURL: "https://new.example"}); err != nil {
		t.Fatalf("update rewrite: %v", err)
	}
	if len(rp.updated) != 1 || rp.updated[0].BaseURL != "https://new.example" || rp.updatedKey[0] != nil {
		t.Fatalf("rewrite: %+v key=%v", rp.updated, rp.updatedKey)
	}
}

// Delete is idempotent by design — upsert makes absence a normal state —
// and both it and the read faces refuse without a subject.
func TestDeleteIdempotentAndGated(t *testing.T) {
	rp := &fakeAgentConfigRP{rows: map[string]*AgentConfig{"u1": {UserID: "u1"}}}
	uc := newAgentConfigUC(rp, &fakeUserRP{})

	if err := uc.Delete(cfgSubject("u1")); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.deleted) != 1 || rp.deleted[0] != "u1" {
		t.Fatalf("deleted: %v", rp.deleted)
	}
	if err := uc.Delete(context.Background()); err == nil {
		t.Fatal("subjectless delete must refuse")
	}
}

// Get answers the caller's own row — nil, nil when absent (the handler
// maps that to an empty response), never another user's.
func TestGetReturnsOwnRowOnly(t *testing.T) {
	rp := &fakeAgentConfigRP{rows: map[string]*AgentConfig{
		"u2": {UserID: "u2", BaseURL: "https://other.example"},
	}}
	uc := newAgentConfigUC(rp, &fakeUserRP{})

	mc, err := uc.Get(cfgSubject("u1"))
	if err != nil || mc != nil {
		t.Fatalf("absent get: mc=%v err=%v", mc, err)
	}
	rp.rows["u1"] = &AgentConfig{UserID: "u1", BaseURL: "https://mine.example"}
	mc, err = uc.Get(cfgSubject("u1"))
	if err != nil || mc == nil || mc.UserID != "u1" {
		t.Fatalf("own get: mc=%v err=%v", mc, err)
	}
	if _, err = uc.Get(context.Background()); !errors.Is(err, kauthz.ErrMissingSubject) {
		t.Fatalf("subjectless get: want MISSING_SUBJECT, got %v", err)
	}
}

// The admin roster's hydration: one brief read per request over the
// deduplicated user set — a denied or missing profile renders as absent
// fields, never a failed row.
func TestListHydratesBriefsBatchedAndDeduplicated(t *testing.T) {
	rp := &fakeAgentConfigRP{listOut: &ListOut{List: []*ListItem{
		{Config: &AgentConfig{UserID: "u1"}},
		{Config: &AgentConfig{UserID: "u2"}},
		{Config: &AgentConfig{UserID: "u1"}},
	}}}
	urp := &fakeUserRP{briefs: map[string]*UserBrief{
		"u1": {ID: "u1"}, "u2": {ID: "u2"},
	}}
	uc := newAgentConfigUC(rp, urp)

	out, err := uc.List(cfgSubject("admin"), &ListIn{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(urp.hydrated) != 1 || len(urp.hydrated[0]) != 2 {
		t.Fatalf("hydrate calls: %v", urp.hydrated)
	}
	for _, it := range out.List {
		if it.User == nil || it.User.ID != it.Config.UserID {
			t.Fatalf("brief join: %+v", it)
		}
	}
	if _, err = uc.List(context.Background(), &ListIn{}); err == nil {
		t.Fatal("subjectless list must refuse")
	}
}

func strPtr(s string) *string { return &s }
