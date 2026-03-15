package entutil

import "github.com/go-kratos/kratos/v2/errors"

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

var defaultMessages = DefaultErrorMessages{
	NotFound:    "未找到相关数据",
	Validation:  "数据校验失败",
	NotSingular: "数据不唯一",
	NotLoaded:   "数据未加载",
	Constraint:  "内容已存在，请勿重复提交",
}

func GetDefaultMessages() DefaultErrorMessages {
	return defaultMessages
}

func HandleEntErrorWithMessages(err error, checker EntErrorChecker, messages DefaultErrorMessages) error {
	switch {
	case checker.IsNotFound(err):
		return errors.NotFound("ENT_NOT_FOUND_ERROR", messages.NotFound).WithCause(err)
	case checker.IsValidationError(err):
		return errors.BadRequest("ENT_VALIDATION_ERROR", messages.Validation).WithCause(err)
	case checker.IsNotSingular(err):
		return errors.BadRequest("ENT_NOT_SINGULAR_ERROR", messages.NotSingular).WithCause(err)
	case checker.IsNotLoaded(err):
		return errors.InternalServer("ENT_NOT_LOADED_ERROR", messages.NotLoaded).WithCause(err)
	case checker.IsConstraintError(err):
		return errors.Conflict("ENT_CONSTRAINT_ERROR", messages.Constraint).WithCause(err)
	default:
		return err
	}
}
