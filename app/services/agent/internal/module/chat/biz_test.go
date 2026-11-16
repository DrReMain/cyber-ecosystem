package chat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/agent/internal/conf"
)

// The master-key floor for the sealing key's consumer: HKDF derivation
// cannot add entropy, so a below-floor key silently weakens every sealed
// API key; construction must refuse it. The floor matches the other
// consumer (agentconfig) so one key passes both constructors or neither.
func TestNewChatUCMasterKeyFloor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctor := func(key string) error {
		_, err := NewChatUC(logger, nil, &conf.Crypto{MasterKey: key}, nil, nil)
		return err
	}
	for _, key := range []string{"", "0123456789abcde"} { // empty; 15 bytes, one short
		if err := ctor(key); err == nil {
			t.Fatalf("master key %q must fail construction", key)
		}
	}
	if _, err := NewChatUC(nil, nil, nil, nil, nil); err == nil {
		t.Fatal("nil crypto config must fail construction")
	}
	if err := ctor("0123456789abcdef"); err != nil { // 16 bytes, the floor
		t.Fatalf("16-byte master key must construct, got: %v", err)
	}
}

// passthroughTm runs the unit inline — the delete cascade's ordering is
// the law under test, not the store's atomicity.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeChatRP answers the session store's thin faces and records every
// call in order; the streaming faces are not this batch's to test.
type fakeChatRP struct {
	ChatRP

	renamed    bool
	titleCalls []struct{ ownerID, sessionID, title string }

	session *ChatSession

	order           []string
	deletedMessages []string
	deletedSessions []string

	sessionsOut *ListSessionsOut
	listedOwner string
	messagesOut *ListMessagesOut
}

func (f *fakeChatRP) UpdateSessionTitle(_ context.Context, ownerID, sessionID, title string) (bool, error) {
	f.titleCalls = append(f.titleCalls, struct{ ownerID, sessionID, title string }{ownerID, sessionID, title})
	return f.renamed, nil
}

func (f *fakeChatRP) FindSessionByID(_ context.Context, ownerID, sessionID string) (*ChatSession, error) {
	if f.session != nil && f.session.OwnerID == ownerID && f.session.ID == sessionID {
		return f.session, nil
	}
	return nil, nil
}

func (f *fakeChatRP) DeleteMessagesBySession(_ context.Context, sessionID string) error {
	f.order = append(f.order, "messages")
	f.deletedMessages = append(f.deletedMessages, sessionID)
	return nil
}

func (f *fakeChatRP) DeleteSessionByID(_ context.Context, ownerID, sessionID string) error {
	f.order = append(f.order, "session")
	f.deletedSessions = append(f.deletedSessions, sessionID)
	return nil
}

func (f *fakeChatRP) ListSessions(_ context.Context, ownerID string, _ *commonpb.PageRequest) (*ListSessionsOut, error) {
	f.listedOwner = ownerID
	return f.sessionsOut, nil
}

func (f *fakeChatRP) ListMessages(_ context.Context, _ string, _ *commonpb.PageRequest) (*ListMessagesOut, error) {
	return f.messagesOut, nil
}

func newChatUC(rp *fakeChatRP) *ChatUC {
	uc, err := NewChatUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{},
		&conf.Crypto{MasterKey: "0123456789abcdef"},
		nil, rp,
	)
	if err != nil {
		panic(err)
	}
	return uc
}

func chatCtx(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID, TenantID: "t1"})
}

// Rename forwards with the caller's id and answers the repo's no-such-row
// as the uniform not-found.
func TestRenameSessionDelegatesAndNotFound(t *testing.T) {
	rp := &fakeChatRP{renamed: true}
	uc := newChatUC(rp)

	if err := uc.RenameSession(chatCtx("u1"), "s1", "new title"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if len(rp.titleCalls) != 1 || rp.titleCalls[0].ownerID != "u1" || rp.titleCalls[0].sessionID != "s1" || rp.titleCalls[0].title != "new title" {
		t.Fatalf("rename ask: %+v", rp.titleCalls)
	}

	rp.renamed = false
	if err := uc.RenameSession(chatCtx("u1"), "s1", "x"); !errorspb.IsGeneralErrorNotFound(err) {
		t.Fatalf("rename miss: want NOT_FOUND, got %v", err)
	}
	if err := uc.RenameSession(context.Background(), "s1", "x"); !errors.Is(err, kauthz.ErrMissingSubject) {
		t.Fatalf("subjectless rename: want MISSING_SUBJECT, got %v", err)
	}
}

// Delete is a cascade inside one transaction — messages first, then the
// session — and a foreign or missing session answers not-found before any
// write.
func TestDeleteSessionCascadesInOrder(t *testing.T) {
	rp := &fakeChatRP{session: &ChatSession{ID: "s1", OwnerID: "u1"}}
	uc := newChatUC(rp)

	if err := uc.DeleteSession(chatCtx("u1"), "s1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(rp.order) != 2 || rp.order[0] != "messages" || rp.order[1] != "session" {
		t.Fatalf("cascade order: %v", rp.order)
	}

	if err := uc.DeleteSession(chatCtx("u2"), "s1"); !errorspb.IsGeneralErrorNotFound(err) {
		t.Fatalf("foreign session delete: want NOT_FOUND, got %v", err)
	}
	if len(rp.order) != 2 {
		t.Fatalf("foreign delete must not write: %v", rp.order)
	}
}

// Both list faces narrow by the caller's id; the message list refuses a
// session the caller does not own before reading a row.
func TestListFacesNarrowByCaller(t *testing.T) {
	want := &ListSessionsOut{}
	rp := &fakeChatRP{sessionsOut: want, session: &ChatSession{ID: "s1", OwnerID: "u1"}, messagesOut: &ListMessagesOut{}}
	uc := newChatUC(rp)

	out, err := uc.ListSessions(chatCtx("u1"), &commonpb.PageRequest{})
	if err != nil || out != want {
		t.Fatalf("list sessions: out=%v err=%v", out, err)
	}
	if rp.listedOwner != "u1" {
		t.Fatalf("list sessions owner: %s", rp.listedOwner)
	}

	if _, err = uc.ListMessages(chatCtx("u1"), "s1", &commonpb.PageRequest{}); err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if _, err = uc.ListMessages(chatCtx("u2"), "s1", &commonpb.PageRequest{}); !errorspb.IsGeneralErrorNotFound(err) {
		t.Fatalf("foreign messages: want NOT_FOUND, got %v", err)
	}
}
