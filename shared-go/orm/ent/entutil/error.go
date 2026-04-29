package entutil

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
)

type EntErrorChecker interface {
	IsNotFound(err error) bool
	IsValidationError(err error) bool
	IsNotSingular(err error) bool
	IsNotLoaded(err error) bool
	IsConstraintError(err error) bool
}

type DefaultError struct {
	NotFound    *errors.Error
	Validation  *errors.Error
	NotSingular *errors.Error
	NotLoaded   *errors.Error
	Constraint  *errors.Error
	Internal    *errors.Error
}

func HandleEntError(err error, checker EntErrorChecker, errs *DefaultError) error {
	if err := validateReasons(errs); err != nil {
		return fmt.Errorf("ent error reason mapping invalid: %w", err)
	}
	switch {
	case checker.IsNotFound(err):
		return errs.NotFound.WithCause(err)
	case checker.IsValidationError(err):
		return errs.Validation.WithCause(err)
	case checker.IsNotSingular(err):
		return errs.NotSingular.WithCause(err)
	case checker.IsNotLoaded(err):
		return errs.NotLoaded.WithCause(err)
	case checker.IsConstraintError(err):
		return errs.Constraint.WithCause(err)
	default:
		if errs.Internal != nil {
			return errs.Internal.WithCause(err)
		}
		return err
	}
}

func validateReasons(errs *DefaultError) error {
	if errs.NotFound == nil {
		return fmt.Errorf("NotFound is nil")
	}
	if errs.Validation == nil {
		return fmt.Errorf("validation is nil")
	}
	if errs.NotSingular == nil {
		return fmt.Errorf("NotSingular is nil")
	}
	if errs.NotLoaded == nil {
		return fmt.Errorf("NotLoaded is nil")
	}
	if errs.Constraint == nil {
		return fmt.Errorf("constraint is nil")
	}
	return nil
}
