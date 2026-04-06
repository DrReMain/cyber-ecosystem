package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/condition"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
	"cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
)

type sessionValidator struct {
	log       *log.Helper
	sessionRP SessionRP
}

func NewSessionValidator(logger log.Logger, sessionRP SessionRP) security.SessionValidator {
	return &sessionValidator{
		log:       log.NewHelper(log.With(logger, "module", "biz/adapter")),
		sessionRP: sessionRP,
	}
}

func (v *sessionValidator) ValidateSession(ctx context.Context, sessionID string) error {
	revoked, err := v.sessionRP.IsSessionRevoked(ctx, sessionID)
	if err != nil {
		return auth.ErrSessionUnavailable.WithCause(err)
	}
	if revoked {
		return auth.ErrSessionRevoked
	}
	return nil
}

type authorizer struct {
	log      *log.Helper
	policyUC *PolicyUC
}

func NewAuthorizer(logger log.Logger, policyUC *PolicyUC) security.Authorizer {
	return &authorizer{
		log:      log.NewHelper(log.With(logger, "module", "biz/adapter")),
		policyUC: policyUC,
	}
}

func (a *authorizer) Authorize(ctx context.Context, subject, operation string) error {
	allowed, err := a.policyUC.Enforce(ctx, subject, operation)
	if err != nil {
		return singletonV1.ErrorErrorReasonUnspecified("").WithCause(err)
	}
	if !allowed {
		return singletonV1.ErrorErrorReasonForbidden("")
	}
	return nil
}

type conditionChecker struct {
	log        *log.Helper
	condUC     *ConditionUC
	userAttrRP UserAttributeRP
}

func NewConditionChecker(logger log.Logger, condUC *ConditionUC, userAttrRP UserAttributeRP) security.ConditionChecker {
	return &conditionChecker{
		log:        log.NewHelper(log.With(logger, "module", "biz/adapter")),
		condUC:     condUC,
		userAttrRP: userAttrRP,
	}
}

func (c *conditionChecker) CheckConditions(ctx context.Context, subject, operation string) error {
	if clientIP, ok := security.ClientIPFromContext(ctx); ok && clientIP != "" {
		ctx = condition.WithClientIP(ctx, clientIP)
	}

	attrs, err := c.userAttrRP.Query(ctx, subject)
	if err != nil {
		return singletonV1.ErrorErrorReasonUnspecified("").WithCause(err)
	}
	attrMap := make(map[string]string, len(attrs))
	for _, a := range attrs {
		if a.Key != nil && a.Value != nil {
			attrMap[*a.Key] = *a.Value
		}
	}
	ctx = condition.WithUserAttributes(ctx, attrMap)

	allowed, err := c.condUC.CheckConditions(ctx, subject, operation)
	if err != nil {
		return singletonV1.ErrorErrorReasonUnspecified("").WithCause(err)
	}
	if !allowed {
		return singletonV1.ErrorErrorReasonForbidden("access condition denied")
	}
	return nil
}

func NewScopeResolver(dataScopeUC *DataScopeUC) datascope.ScopeResolveFunc {
	return dataScopeUC.ResolveScope
}
