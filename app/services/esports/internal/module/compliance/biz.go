package compliance

import (
	"context"
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	// The shipped defaults an absent row answers. The schema's column
	// defaults materialize the same pair the moment the singleton row is
	// first written — two homes, one frozen pair: 18 is the product's
	// line, and changing it is this knob's runtime job, never a code edit.
	DefaultBirthCollect = true
	DefaultPayAgeLine   = 18
)

// DO ------------------------------------------------------------------------------------------------------------------

type ComplianceConfig struct {
	BirthCollect bool
	PayAgeLine   int32
	// nil = no override: each face renders its own localized default copy,
	// because the shipped copy is locale-dependent text.
	MinorPrompt        *string
	MissingBirthPrompt *string
}

// Port ---------------------------------------------------------------------------------------------------------------

type ComplianceRP interface {
	// Each setter updates only its own column of the singleton row and
	// answers the whole settled config.
	SetBirthCollection(ctx context.Context, collect bool, updatedBy string) (*ComplianceConfig, error)
	SetCompliancePrompts(ctx context.Context, minor, missing *string, updatedBy string) (*ComplianceConfig, error)
	SetPayAgeLine(ctx context.Context, age int32, updatedBy string) (*ComplianceConfig, error)
	Get(ctx context.Context) (*ComplianceConfig, error)
}

// UC -----------------------------------------------------------------------------------------------------------------

type ComplianceUC struct {
	shared.UC
	complianceRP ComplianceRP
}

func NewComplianceUC(logger *slog.Logger, tm shared.Transaction, complianceRP ComplianceRP) *ComplianceUC {
	return &ComplianceUC{
		UC:           shared.NewUC(logger.With("module", "module/compliance"), tm),
		complianceRP: complianceRP,
	}
}

// Method -------------------------------------------------------------------------------------------------------------

func (uc *ComplianceUC) SetBirthCollection(ctx context.Context, collect bool) (*ComplianceConfig, error) {
	updatedBy, err := actorID(ctx)
	if err != nil {
		return nil, err
	}
	return uc.complianceRP.SetBirthCollection(ctx, collect, updatedBy)
}

func (uc *ComplianceUC) SetCompliancePrompts(ctx context.Context, minor, missing *string) (*ComplianceConfig, error) {
	// Nil carries "leave that prompt's stored value untouched" — the face
	// sends only the one it is editing.
	updatedBy, err := actorID(ctx)
	if err != nil {
		return nil, err
	}
	return uc.complianceRP.SetCompliancePrompts(ctx, minor, missing, updatedBy)
}

func (uc *ComplianceUC) SetPayAgeLine(ctx context.Context, age int32) (*ComplianceConfig, error) {
	// The same bounds the wire enforces — a direct UC caller (e2e, a
	// future scheduler) deserves the refusal too, not a crooked line.
	if age < 1 || age > 120 {
		return nil, errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(fmt.Errorf("pay age line %d outside 1..120", age))
	}
	updatedBy, err := actorID(ctx)
	if err != nil {
		return nil, err
	}
	return uc.complianceRP.SetPayAgeLine(ctx, age, updatedBy)
}

func (uc *ComplianceUC) Get(ctx context.Context) (*ComplianceConfig, error) {
	return uc.complianceRP.Get(ctx)
}

// Private ------------------------------------------------------------------------------------------------------------

func actorID(ctx context.Context) (string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("compliance config writes require an admin session"))
	}
	return subject.UserID, nil
}
