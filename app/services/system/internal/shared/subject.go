package shared

import (
	"context"
	"fmt"

	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// SubjectUserID extracts the caller's user id; a subjectless context is
// unauthenticated.
func SubjectUserID(ctx context.Context) (string, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return "", errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	return subject.UserID, nil
}
