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

type DefaultErrorMessages struct {
	NotFound    string
	Validation  string
	NotSingular string
	NotLoaded   string
	Constraint  string
}

type DefaultErrorReasons struct {
	NotFound    string
	Validation  string
	NotSingular string
	NotLoaded   string
	Constraint  string
}

var defaultMessages = DefaultErrorMessages{
	NotFound:    "",
	Validation:  "",
	NotSingular: "",
	NotLoaded:   "",
	Constraint:  "",
}

func GetDefaultMessages() DefaultErrorMessages {
	return defaultMessages
}

func HandleEntError(err error, checker EntErrorChecker, reasons DefaultErrorReasons, messages DefaultErrorMessages) error {
	if err := validateReasons(reasons); err != nil {
		return fmt.Errorf("ent error reason mapping invalid: %w", err)
	}
	switch {
	case checker.IsNotFound(err):
		return errors.NotFound(reasons.NotFound, messages.NotFound).WithCause(err)
	case checker.IsValidationError(err):
		return errors.BadRequest(reasons.Validation, messages.Validation).WithCause(err)
	case checker.IsNotSingular(err):
		return errors.BadRequest(reasons.NotSingular, messages.NotSingular).WithCause(err)
	case checker.IsNotLoaded(err):
		return errors.InternalServer(reasons.NotLoaded, messages.NotLoaded).WithCause(err)
	case checker.IsConstraintError(err):
		return errors.Conflict(reasons.Constraint, messages.Constraint).WithCause(err)
	default:
		return err
	}
}

func validateReasons(reasons DefaultErrorReasons) error {
	if reasons.NotFound == "" {
		return fmt.Errorf("NotFound reason is empty")
	}
	if reasons.Validation == "" {
		return fmt.Errorf("Validation reason is empty")
	}
	if reasons.NotSingular == "" {
		return fmt.Errorf("NotSingular reason is empty")
	}
	if reasons.NotLoaded == "" {
		return fmt.Errorf("NotLoaded reason is empty")
	}
	if reasons.Constraint == "" {
		return fmt.Errorf("Constraint reason is empty")
	}
	return nil
}
