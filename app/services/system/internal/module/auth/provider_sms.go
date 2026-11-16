package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/capability/cache"
	"cyber-ecosystem/shared-go/capability/sms"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const smsKind = "sms"

const (
	smsCodePrefix = "auth:web:sms-code:" // + canonical phone; the value is the live code
	smsCodeTTL    = 10 * time.Minute
	smsCodeDigits = 6
	// One send per window and a daily cap per phone — the whole brute-force
	// budget derives from here: every exchange burns its code (GetDel) and
	// its state (one-shot), so a guess costs a send.
	smsSendWindow  = 60 * time.Second
	smsSendPerDay  = 10
	smsDayWindow   = 24 * time.Hour
	smsSendPrefix  = "auth:sms:send-window:"
	smsDailyPrefix = "auth:sms:send-daily:"
	// The per-source pair is orthogonal to the per-phone pair: phone keys
	// protect one number's holder; source keys protect the send budget from
	// number-walking. A NAT-shared exit folds a whole office into one
	// source — the failure direction is over-restriction, never exposure.
	smsSrcWindow    = time.Second
	smsSrcPerDay    = 100
	smsSrcWinPrefix = "auth:sms:src-window:"
	smsSrcDayPrefix = "auth:sms:src-daily:"
)

var mainlandMobile = regexp.MustCompile(`^1[3-9][0-9]{9}$`)

type smsProvider struct {
	log         *slog.Logger
	sender      sms.Sender
	handleSms   platform.SmsErrorHandler
	handleCache platform.CacheErrorHandler
	cache       *cache.Cache
	tokenRP     TokenRP
}

func NewSmsProvider(
	logger *slog.Logger,
	sender sms.Sender,
	handleSms platform.SmsErrorHandler,
	handleCache platform.CacheErrorHandler,
	c *cache.Cache,
	tokenRP TokenRP,
) *smsProvider {
	return &smsProvider{
		log:         logger.With("module", "module/auth", "provider", "sms"),
		sender:      sender,
		handleSms:   handleSms,
		handleCache: handleCache,
		cache:       c,
		tokenRP:     tokenRP,
	}
}

func (p *smsProvider) sendBudget(ctx context.Context, winKey string, window time.Duration, dayKey string, perDay int, who string) error {
	// One dimension of the issuance budget (per-phone, or per-source
	// against number-walking): one send per window plus a daily cap. The
	// daily counter's TTL is re-asserted whenever it reads as unset, so a
	// lost or failed Expire at the counter's birth cannot make the budget
	// immortal — the heal lands on a later send at the latest.
	res, err := p.cache.RateLimiter.Allow(ctx, winKey, 1, window)
	if err != nil {
		return p.handleCache(err)
	}
	if res != nil && !res.Allowed {
		return systempb.ErrorSystemSmsRateLimited("").WithCause(
			fmt.Errorf("send window: %s", who))
	}
	n, err := p.cache.Counter.Incr(ctx, dayKey, 1)
	if err != nil {
		return p.handleCache(err)
	}
	ttl, terr := p.cache.KV.GetTTL(ctx, dayKey)
	if terr != nil {
		return p.handleCache(terr)
	}
	if ttl == 0 {
		if err := p.cache.Counter.Expire(ctx, dayKey, smsDayWindow); err != nil {
			return p.handleCache(err)
		}
	}
	if n > int64(perDay) {
		return systempb.ErrorSystemSmsRateLimited("").WithCause(
			fmt.Errorf("daily cap: %s", who))
	}
	return nil
}

func (p *smsProvider) IssueCode(ctx context.Context, row *conf.Auth_Provider, phone, source string) (string, error) {
	canon := normalizePhone(phone)
	if canon == "" {
		return "", errorspb.ErrorGeneralErrorValidationFailed("").WithCause(
			fmt.Errorf("sms: not a mainland mobile: %s", maskPhone(phone)))
	}
	// Source first, then phone: the source pair spends nothing on any one
	// number, so a walker exhausts its own source caps before denting any
	// phone budget. An empty source (in-process callers, tests) skips the
	// source dimension only — the phone budgets still hold.
	if source != "" {
		if err := p.sendBudget(ctx, smsSrcWinPrefix+source, smsSrcWindow, smsSrcDayPrefix+source, smsSrcPerDay, source); err != nil {
			return "", err
		}
	}
	if err := p.sendBudget(ctx, smsSendPrefix+canon, smsSendWindow, smsDailyPrefix+canon, smsSendPerDay, maskPhone(canon)); err != nil {
		return "", err
	}
	code, err := randomDigits(smsCodeDigits)
	if err != nil {
		return "", errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
	// The rotation precedes the dial: the store holds one live code per
	// phone, so merely attempting a resend kills the prior pair.
	if err := p.tokenRP.Set(ctx, smsCodePrefix+canon, []byte(code), smsCodeTTL); err != nil {
		return "", err
	}
	sc := row.GetSms()
	if _, err := p.sender.Send(ctx, &sms.Message{
		To:           canon,
		SignName:     sc.GetSignName(),
		TemplateCode: sc.GetTemplateCode(),
		Params:       map[string]string{"code": code},
	}); err != nil {
		// A refused channel also burns the fresh code — otherwise every
		// failed send (config typo, carrier throttle) left a redeemable
		// ghost behind while the send window kept burning: the caller
		// was told "failed", yet the server held a live credential no
		// phone ever saw.
		_ = p.tokenRP.Del(ctx, smsCodePrefix+canon)
		return "", p.handleSms(err)
	}
	// The state is minted only after a delivered send: a caller never
	// holds a state whose code never traveled.
	state, err := mintIdentityState(ctx, p.tokenRP, row.GetKey(), "")
	if err != nil {
		return "", err
	}
	p.log.Info("sms code issued", "phone", maskPhone(canon), "provider", row.GetKey())
	return state, nil
}

func (p *smsProvider) Exchange(ctx context.Context, row *conf.Auth_Provider, payload any) (*exchangedIdentity, error) {
	// A request narrows into the sms view by assertion — the identity
	// requests carrying the sms member (LoginByIdentity, BindCredential,
	// MergeAccounts) all satisfy it.
	sp, ok := payload.(smsPayload)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorValidationFailed("").WithCause(fmt.Errorf("sms payload required"))
	}
	canon := normalizePhone(sp.GetSms().GetPhone())
	if canon == "" {
		return nil, errorspb.ErrorGeneralErrorValidationFailed("").WithCause(
			fmt.Errorf("sms: not a mainland mobile: %s", maskPhone(sp.GetSms().GetPhone())))
	}
	// A miss (never sent, expired, or already burned) and a mismatch answer
	// identically — the caller learns nothing about which.
	stored, err := p.tokenRP.GetDel(ctx, smsCodePrefix+canon)
	if err != nil {
		if errorspb.IsInfraErrorCacheMiss(err) {
			return nil, systempb.ErrorSystemSmsCodeInvalid("").WithCause(
				fmt.Errorf("no live code: %s", maskPhone(canon)))
		}
		return nil, err
	}
	if string(stored) != sp.GetSms().GetCode() {
		return nil, systempb.ErrorSystemSmsCodeInvalid("").WithCause(
			fmt.Errorf("code mismatch: %s", maskPhone(canon)))
	}
	return &exchangedIdentity{Kind: smsKind, Identifier: canon}, nil
}

func normalizePhone(phone string) string {
	// The canonical form (bare mainland mobile) is the credential
	// identifier; send and exchange must derive the same string. The
	// market rule is this channel's own — the contract checks shape only.
	s := strings.TrimPrefix(phone, "+86")
	if mainlandMobile.MatchString(s) {
		return s
	}
	return ""
}

func maskPhone(phone string) string {
	// Thin alias: the masking law lives in utils, this name keeps the
	// package's own call sites reading naturally.
	return utils.MaskPhone(phone)
}

func randomDigits(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	for b.Len() < n {
		var buf [32]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("sms: entropy failure: %w", err)
		}
		for _, v := range buf {
			if v >= 250 { // 25 full decades in a byte; the tail biases
				continue
			}
			b.WriteByte('0' + v%10)
			if b.Len() == n {
				break
			}
		}
	}
	return b.String(), nil
}
