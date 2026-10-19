package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/kratos/security"
	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"
	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/agent/internal/conf"
	"cyber-ecosystem/app/services/agent/internal/module/agentconfig"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

const (
	maxHistoryTurns = 48     // replayed-context turn cap: coherent multi-turn follow-ups without unbounded request size
	maxHistoryChars = 24_000 // ≈6k replayed tokens (4 bytes ≈ 1 token heuristic); the S4 harness swaps this for error-adaptive shrinking
)

type modelsResp struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// DO ------------------------------------------------------------------------------------------------------------------

type Turn struct {
	Role    string
	Content string
}

type Delta struct {
	Content   string
	Reasoning string
}

type ChatSession struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	OwnerID   string
	Title     string
}

type SessionMessage struct {
	ID        string
	CreatedAt time.Time
	Role      string
	Content   string
	Reasoning string
	Model     string
	Finish    string
}

type PreparedSession struct {
	ID      string
	Title   string
	History []*Turn
	IsNew   bool // true = id pre-allocated; the session row is inserted only when the exchange persists
}

type ExchangeIn struct {
	Session            *PreparedSession
	Model              string
	UserContent        string
	AssistantContent   string
	AssistantReasoning string
	Finish             string
}

type ListSessionsOut struct {
	*commonpb.PageResponse
	List []*ChatSession
}

type ListMessagesOut struct {
	*commonpb.PageResponse
	List []*SessionMessage
}

// Port ------------------------------------------------------------------------------------------------------------------

type ChatRP interface {
	CreateWithExchange(ctx context.Context, sess *ChatSession, user, assistant *SessionMessage) error
	AppendExchange(ctx context.Context, sessionID string, user, assistant *SessionMessage) error
	TouchSession(ctx context.Context, ownerID, sessionID string) error
	UpdateSessionTitle(ctx context.Context, ownerID, sessionID, title string) (bool, error)
	DeleteSessionByID(ctx context.Context, ownerID, sessionID string) error
	DeleteMessagesBySession(ctx context.Context, sessionID string) error
	ListSessions(ctx context.Context, ownerID string, page *commonpb.PageRequest) (*ListSessionsOut, error)
	ListMessages(ctx context.Context, sessionID string, page *commonpb.PageRequest) (*ListMessagesOut, error)
	FindSessionByID(ctx context.Context, ownerID, sessionID string) (*ChatSession, error)
	LatestTurns(ctx context.Context, sessionID string, limit int) ([]*Turn, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ChatUC struct {
	shared.UC
	masterKey string

	agentConfigRP agentconfig.AgentConfigRP
	chatRP        ChatRP
}

func NewChatUC(logger *slog.Logger, tm shared.Transaction, c *conf.Crypto, agentConfigRP agentconfig.AgentConfigRP, chatRP ChatRP) (*ChatUC, error) {
	if c == nil || c.GetMasterKey() == "" {
		return nil, fmt.Errorf("crypto.master_key is required")
	}
	return &ChatUC{
		UC:        shared.NewUC(logger.With("module", "module/chat"), tm),
		masterKey: c.GetMasterKey(),

		agentConfigRP: agentConfigRP,
		chatRP:        chatRP,
	}, nil
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ChatUC) CreateOrLoad(ctx context.Context, sessionID, content string) (*PreparedSession, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}
	prep := &PreparedSession{}
	if sessionID == "" {
		// Pre-allocate the id and defer the INSERT to AppendExchange: an
		// upstream failure then leaves the id existing nowhere, so the table
		// never holds an empty session and no rollback is needed.
		prep.ID = xid.New().String()
		prep.Title = clipTitle(content)
		prep.IsNew = true
	} else {
		// nil, nil = absent or unowned — the me-face never leaks which.
		sess, ferr := uc.chatRP.FindSessionByID(ctx, userID, sessionID)
		if ferr != nil {
			return nil, ferr
		}
		if sess == nil {
			return nil, errorspb.ErrorGeneralErrorNotFound("")
		}
		prep.ID = sess.ID
		prep.Title = sess.Title
		// Sending is what the rail ranks by: bump updated_at at exchange
		// start, not when the stream settles.
		if terr := uc.chatRP.TouchSession(ctx, userID, sessionID); terr != nil {
			return nil, terr
		}
	}
	// Newest-first walk under the char budget (the count is already capped
	// by LatestTurns), then reverse to conversation order.
	latest, err := uc.chatRP.LatestTurns(ctx, prep.ID, maxHistoryTurns)
	if err != nil {
		return nil, err
	}
	total := 0
	window := make([]*Turn, 0, len(latest))
	for _, t := range latest {
		if len(window) >= maxHistoryTurns || total+len(t.Content) > maxHistoryChars {
			break
		}
		window = append(window, t)
		total += len(t.Content)
	}
	slices.Reverse(window)
	prep.History = window
	return prep, nil
}

func (uc *ChatUC) AppendExchange(ctx context.Context, in *ExchangeIn) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return kauthz.ErrMissingSubject
	}
	user := &SessionMessage{Role: "user", Content: in.UserContent}
	assistant := &SessionMessage{
		Role:      "assistant",
		Content:   in.AssistantContent,
		Reasoning: in.AssistantReasoning,
		Model:     in.Model,
		Finish:    in.Finish,
	}
	return uc.Tm.InTx(ctx, func(tctx context.Context) error {
		if in.Session.IsNew {
			return uc.chatRP.CreateWithExchange(tctx, &ChatSession{
				ID:      in.Session.ID,
				OwnerID: subject.UserID,
				Title:   in.Session.Title,
			}, user, assistant)
		}
		return uc.chatRP.AppendExchange(tctx, in.Session.ID, user, assistant)
	})
}

func (uc *ChatUC) RenameSession(ctx context.Context, sessionID, title string) error {
	userID, err := callerID(ctx)
	if err != nil {
		return err
	}
	renamed, err := uc.chatRP.UpdateSessionTitle(ctx, userID, sessionID, title)
	if err != nil {
		return err
	}
	if !renamed {
		return errorspb.ErrorGeneralErrorNotFound("")
	}
	return nil
}

func (uc *ChatUC) DeleteSession(ctx context.Context, sessionID string) error {
	userID, err := callerID(ctx)
	if err != nil {
		return err
	}
	sess, ferr := uc.chatRP.FindSessionByID(ctx, userID, sessionID)
	if ferr != nil {
		return ferr
	}
	if sess == nil {
		return errorspb.ErrorGeneralErrorNotFound("")
	}
	return uc.Tm.InTx(ctx, func(tctx context.Context) error {
		if derr := uc.chatRP.DeleteMessagesBySession(tctx, sessionID); derr != nil {
			return derr
		}
		return uc.chatRP.DeleteSessionByID(tctx, userID, sessionID)
	})
}

func (uc *ChatUC) ListModels(ctx context.Context) ([]string, error) {
	baseURL, apiKey, err := uc.resolveDial(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/models", nil)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("build models request: %w", err))
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, mapUpstreamErr(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, mapStatusErr(resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, mapUpstreamErr(err)
	}
	out, err := utils.Unmarshal[modelsResp](body)
	if err != nil {
		return nil, errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("decode models response: %w", err))
	}
	models := make([]string, 0, len(out.Data))
	for _, m := range out.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

func (uc *ChatUC) ListSessions(ctx context.Context, in *commonpb.PageRequest) (*ListSessionsOut, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}
	return uc.chatRP.ListSessions(ctx, userID, in)
}

func (uc *ChatUC) ListMessages(ctx context.Context, sessionID string, in *commonpb.PageRequest) (*ListMessagesOut, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}
	sess, ferr := uc.chatRP.FindSessionByID(ctx, userID, sessionID)
	if ferr != nil {
		return nil, ferr
	}
	if sess == nil {
		return nil, errorspb.ErrorGeneralErrorNotFound("")
	}
	return uc.chatRP.ListMessages(ctx, sessionID, in)
}

func (uc *ChatUC) Stream(ctx context.Context, model string, turns []*Turn, emit func(*Delta) error) (string, error) {
	baseURL, apiKey, err := uc.resolveDial(ctx)
	if err != nil {
		return "", err
	}
	// Per-request construction: base_url is per-user, so there is no shared
	// pool to reuse; stream lifetime is bounded by ctx, not by a client timeout.
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	})
	if err != nil {
		return "", mapUpstreamErr(err)
	}
	sr, err := cm.Stream(ctx, toSchema(turns))
	if err != nil {
		return "", mapUpstreamErr(err)
	}
	defer sr.Close()
	finish := "stop"
	for {
		msg, rerr := sr.Recv()
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			// A canceled ctx is the client hang-up path, not an endpoint failure.
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "", mapUpstreamErr(rerr)
		}
		if msg.Content != "" || msg.ReasoningContent != "" {
			if serr := emit(&Delta{Content: msg.Content, Reasoning: msg.ReasoningContent}); serr != nil {
				return "", serr
			}
		}
		if msg.ResponseMeta != nil && msg.ResponseMeta.FinishReason != "" {
			finish = msg.ResponseMeta.FinishReason
		}
	}
	return finish, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *ChatUC) resolveDial(ctx context.Context) (string, string, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return "", "", err
	}
	mc, sealed, err := uc.agentConfigRP.FindSealedByUserID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if mc == nil {
		return "", "", errorspb.ErrorGeneralErrorNotFound("")
	}
	if sealed == "" {
		return mc.BaseURL, "", nil
	}
	apiKey, err := utils.Open(sealed, uc.masterKey)
	if err != nil {
		return "", "", errorspb.ErrorGeneralErrorInternal("").WithCause(fmt.Errorf("open api key: %w", err))
	}
	return mc.BaseURL, apiKey, nil
}

func callerID(ctx context.Context) (string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return "", kauthz.ErrMissingSubject
	}
	return subject.UserID, nil
}

func clipTitle(content string) string {
	flat := strings.Join(strings.Fields(content), " ")
	runes := []rune(flat)
	if len(runes) > 20 {
		return string(runes[:20]) + "…"
	}
	return flat
}

func toSchema(turns []*Turn) []*schema.Message {
	return utils.SliceMap(turns, func(t *Turn) *schema.Message {
		return &schema.Message{Role: schema.RoleType(t.Role), Content: t.Content}
	})
}

func mapUpstreamErr(err error) error {
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return errorspb.ErrorInfraErrorNetworkTimeout("").WithCause(fmt.Errorf("llm endpoint: %w", err))
	}
	return errorspb.ErrorInfraErrorNetworkConnection("").WithCause(fmt.Errorf("llm endpoint: %w", err))
}

func mapStatusErr(status int) error {
	if status >= 400 && status < 500 {
		return errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("llm endpoint status: %d", status))
	}
	return errorspb.ErrorGeneralErrorUnavailable("").WithCause(fmt.Errorf("llm endpoint status: %d", status))
}
