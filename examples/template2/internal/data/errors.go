package data

import (
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent"

	template2V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template2/v1"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"
)

type entErrorChecker struct{}

func (c *entErrorChecker) IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
func (c *entErrorChecker) IsValidationError(err error) bool {
	return ent.IsValidationError(err)
}
func (c *entErrorChecker) IsNotSingular(err error) bool {
	return ent.IsNotSingular(err)
}
func (c *entErrorChecker) IsNotLoaded(err error) bool {
	return ent.IsNotLoaded(err)
}
func (c *entErrorChecker) IsConstraintError(err error) bool {
	return ent.IsConstraintError(err)
}

func HandleError(err error) error {
	return entutil.HandleEntError(err, &entErrorChecker{},
		entutil.DefaultErrorReasons{
			NotFound:    template2V1.ErrorReason_ERROR_REASON_ENT_NOT_FOUND.String(),
			Validation:  template2V1.ErrorReason_ERROR_REASON_ENT_VALIDATION.String(),
			NotSingular: template2V1.ErrorReason_ERROR_REASON_ENT_NOT_SINGULAR.String(),
			NotLoaded:   template2V1.ErrorReason_ERROR_REASON_ENT_NOT_LOADED.String(),
			Constraint:  template2V1.ErrorReason_ERROR_REASON_ENT_CONSTRAINT.String(),
		},
		entutil.GetDefaultMessages(),
	)
}
