package platform

import (
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/ent"
)

type entErrorChecker struct{}

func (entErrorChecker) IsNotFound(err error) bool        { return ent.IsNotFound(err) }
func (entErrorChecker) IsValidationError(err error) bool { return ent.IsValidationError(err) }
func (entErrorChecker) IsNotSingular(err error) bool     { return ent.IsNotSingular(err) }
func (entErrorChecker) IsNotLoaded(err error) bool       { return ent.IsNotLoaded(err) }
func (entErrorChecker) IsConstraintError(err error) bool { return ent.IsConstraintError(err) }

var defaultEntError = &entutil.DefaultError{
	NotFound:    errorspb.ErrorInfraErrorDbNotFound(""),
	Validation:  errorspb.ErrorInfraErrorDbValidation(""),
	NotSingular: errorspb.ErrorInfraErrorDbNotSingular(""),
	NotLoaded:   errorspb.ErrorInfraErrorDbNotLoaded(""),
	Constraint:  errorspb.ErrorInfraErrorDbConstraint(""),
	Internal:    errorspb.ErrorInfraErrorDbInternal(""),
}

func NewEntErrorHandler() EntErrorHandler {
	return func(err error) error {
		return entutil.HandleEntError(err, entErrorChecker{}, defaultEntError)
	}
}
