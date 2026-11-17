package e2e

// The chat streaming face, live over the dev stack: a real ent-backed
// session/message store, the real eino openai SSE consumer, and a fake
// openai wire endpoint — everything between the RPC face and the LLM is
// production code. The grpc transport shell is faked (a stream recorder);
// transport is the library's face, the chat chain is ours.
//
// Gated: ENT_E2E=1 with the dev stack (postgres) up runs it.

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"google.golang.org/grpc"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	agentpb "cyber-ecosystem/gen/go/cyber/agent/v1"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/ent"
	entagentconfig "cyber-ecosystem/app/services/agent/internal/ent/agentconfig"
	entchatmessage "cyber-ecosystem/app/services/agent/internal/ent/chatmessage"
	entchatsession "cyber-ecosystem/app/services/agent/internal/ent/chatsession"
	agentconfigmod "cyber-ecosystem/app/services/agent/internal/module/agentconfig"
	chatmod "cyber-ecosystem/app/services/agent/internal/module/chat"
	"cyber-ecosystem/app/services/agent/internal/platform"
)

const (
	csUser      = "e2e-chat-stream"
	csMasterKey = "e2e-master-key-0123456789abcdef" // 32 bytes, over the 16 floor
)

type fakeStream struct {
	grpc.ServerStream
	ctx    context.Context
	frames []*agentpb.ChatResponse
}

func (f *fakeStream) Send(m *agentpb.ChatResponse) error { f.frames = append(f.frames, m); return nil }
func (f *fakeStream) Context() context.Context           { return f.ctx }

func chatStreamClean(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.AgentConfig.Delete().Where(entagentconfig.UserID(csUser)).Exec(ctx); err != nil {
		t.Fatalf("clean agent configs: %v", err)
	}
	ids, err := db.ChatSession.Query().Where(entchatsession.OwnerID(csUser)).IDs(ctx)
	if err != nil {
		t.Fatalf("find owned sessions: %v", err)
	}
	if len(ids) > 0 {
		if _, err := db.ChatMessage.Delete().Where(entchatmessage.SessionIDIn(ids...)).Exec(ctx); err != nil {
			t.Fatalf("clean chat messages: %v", err)
		}
	}
	if _, err := db.ChatSession.Delete().Where(entchatsession.OwnerID(csUser)).Exec(ctx); err != nil {
		t.Fatalf("clean chat sessions: %v", err)
	}
}

// The fake openai wire endpoint: two content deltas then DONE — the minimum
// shape the eino consumer turns into an ordered frame sequence. It binds a
// non-loopback local address: the dial guard correctly refuses loopback, and
// private unicast is a first-class target, so the fake lives where the guard
// lets production dials go.
func nonLoopbackAddr(t *testing.T) string {
	t.Helper()
	ifaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("interfaces: %v", err)
	}
	for _, ifc := range ifaces {
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok && ipn.IP.To4() != nil && !ipn.IP.IsLoopback() {
				return ipn.IP.String()
			}
		}
	}
	t.Skip("no non-loopback local address to host the fake endpoint on")
	return ""
}

func fakeOpenAIEndpoint(t *testing.T) *httptest.Server {
	t.Helper()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fl := w.(http.Flusher)
		for _, chunk := range []string{
			`data: {"id":"e2e","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"e2e","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" streaming"},"finish_reason":null}]}`,
			`data: [DONE]`,
		} {
			if _, err := fmt.Fprintf(w, "%s\n\n", chunk); err != nil {
				return
			}
			fl.Flush()
		}
	})
	srv := httptest.NewUnstartedServer(handler)
	l, err := net.Listen("tcp", nonLoopbackAddr(t)+":0")
	if err != nil {
		t.Fatalf("listen on a non-loopback address: %v", err)
	}
	srv.Listener = l
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

func TestChatStreamLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev stack (postgres) up to run the chat-stream end-to-end")
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: twEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: twEnvOr("DB_PASS", "postgres"), DbName: twEnvOr("DB_NAME", "agent"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler())
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	chatStreamClean(t, db)
	t.Cleanup(func() {
		chatStreamClean(t, db)
		closeDB()
	})

	endpoint := fakeOpenAIEndpoint(t)
	sealed, err := utils.Seal("e2e-api-key", csMasterKey)
	if err != nil {
		t.Fatalf("seal api key: %v", err)
	}
	if err := db.AgentConfig.Create().
		SetUserID(csUser).
		SetBaseURL(endpoint.URL + "/v1").
		SetAPIKey(sealed).
		Exec(context.Background()); err != nil {
		t.Fatalf("seed agent config: %v", err)
	}

	agentConfigRP := agentconfigmod.NewAgentConfigRP(logger, pf)
	chatRP := chatmod.NewChatRP(logger, pf)
	uc, err := chatmod.NewChatUC(logger, pf, &conf.Crypto{MasterKey: csMasterKey}, agentConfigRP, chatRP)
	if err != nil {
		t.Fatalf("chat uc: %v", err)
	}
	svc := chatmod.NewChatService(logger, uc)

	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: csUser, TenantID: "default"})
	stream := &fakeStream{ctx: ctx}
	if err := svc.Chat(&agentpb.ChatRequest{Content: new("Say it"), Model: new("e2e-model")}, stream); err != nil {
		t.Fatalf("chat stream: %v", err)
	}

	// Frame order: the session id echoes first (new session), then the
	// content deltas in wire order.
	if len(stream.frames) < 2 {
		t.Fatalf("expected the id echo plus deltas, got %d frames", len(stream.frames))
	}
	sessionID := stream.frames[0].GetSessionId()
	if sessionID == "" {
		t.Fatal("the first frame of a new session must echo the session id")
	}
	var got strings.Builder
	for _, f := range stream.frames[1:] {
		got.WriteString(f.GetContentDelta())
	}
	if got.String() != "Hello streaming" {
		t.Fatalf("deltas must arrive in wire order, got %q", got.String())
	}

	// Persistence: the session row exists for this owner and carries both
	// turns — the user's ask and the assistant's settled text.
	sess, err := db.ChatSession.Query().Where(entchatsession.ID(sessionID)).Only(context.Background())
	if err != nil {
		t.Fatalf("session row: %v", err)
	}
	if sess.OwnerID != csUser {
		t.Fatalf("session owner: %q", sess.OwnerID)
	}
	turns, err := db.ChatMessage.Query().Where(entchatmessage.SessionID(sessionID)).All(context.Background())
	if err != nil {
		t.Fatalf("message rows: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected the user turn and the settled assistant turn, got %d", len(turns))
	}
}
