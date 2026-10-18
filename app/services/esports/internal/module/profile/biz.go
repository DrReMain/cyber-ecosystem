package profile

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Profile struct {
	ID string
}

// UC ------------------------------------------------------------------------------------------------------------------

type ProfileUC struct {
	shared.UC
}

func NewProfileUC(logger *slog.Logger, tm shared.Transaction) *ProfileUC {
	return &ProfileUC{UC: shared.NewUC(logger.With("module", "module/profile"), tm)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ProfileUC) Get(ctx context.Context) *Profile {
	subject, _ := security.SubjectFromCtx(ctx)
	if subject == nil {
		return &Profile{}
	}
	return &Profile{ID: subject.UserID}
}
