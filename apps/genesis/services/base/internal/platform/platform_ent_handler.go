package platform

import (
	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/apps/genesis/services/base/internal/ent"
)

type entErrorChecker struct{}

func (c *entErrorChecker) IsNotFound(err error) bool        { return ent.IsNotFound(err) }
func (c *entErrorChecker) IsValidationError(err error) bool { return ent.IsValidationError(err) }
func (c *entErrorChecker) IsNotSingular(err error) bool     { return ent.IsNotSingular(err) }
func (c *entErrorChecker) IsNotLoaded(err error) bool       { return ent.IsNotLoaded(err) }
func (c *entErrorChecker) IsConstraintError(err error) bool { return ent.IsConstraintError(err) }

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
		return entutil.HandleEntError(err, &entErrorChecker{}, defaultEntError)
	}
}
