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

func (p *smsProvider) IssueCode(ctx context.Context, row *conf.Auth_Provider, phone string) (string, error) {
	canon := normalizePhone(phone)
	if canon == "" {
		return "", errorspb.ErrorGeneralErrorValidationFailed("").WithCause(
			fmt.Errorf("sms: not a mainland mobile: %s", maskPhone(phone)))
	}
	res, err := p.cache.RateLimiter.Allow(ctx, smsSendPrefix+canon, 1, smsSendWindow)
	if err != nil {
		return "", p.handleCache(err)
	}
	if res != nil && !res.Allowed {
		return "", systempb.ErrorSystemSmsRateLimited("").WithCause(
			fmt.Errorf("send window: %s", maskPhone(canon)))
	}
	n, err := p.cache.Counter.Incr(ctx, smsDailyPrefix+canon, 1)
	if err != nil {
		return "", p.handleCache(err)
	}
	if n == 1 {
		if err := p.cache.Counter.Expire(ctx, smsDailyPrefix+canon, smsDayWindow); err != nil {
			return "", p.handleCache(err)
		}
	}
	if n > smsSendPerDay {
		return "", systempb.ErrorSystemSmsRateLimited("").WithCause(
			fmt.Errorf("daily cap: %s", maskPhone(canon)))
	}
	code, err := randomDigits(smsCodeDigits)
	if err != nil {
		return "", errorspb.ErrorGeneralErrorInternal("").WithCause(err)
	}
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
		return "", p.handleSms(err)
	}
	// The state is minted only after a delivered send: a caller never
	// holds a state whose code never traveled.
	state, err := mintIdentityState(ctx, p.tokenRP, row.GetKey())
	if err != nil {
		return "", err
	}
	p.log.Info("sms code issued", "phone", maskPhone(canon), "provider", row.GetKey())
	return state, nil
}

func (p *smsProvider) Exchange(ctx context.Context, row *conf.Auth_Provider, payload identityPayload) (*exchangedIdentity, error) {
	// A request narrows into the sms view by assertion — its oneof grows
	// into the GetSms method as it admits the sms member
	// (LoginByIdentity carries it; BindCredential does not yet).
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
	if len(phone) < 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
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
