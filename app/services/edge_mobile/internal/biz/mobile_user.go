package biz

import (
	"time"

	"cyber-ecosystem/shared-go/helper"

	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
)

type MobileUser struct {
	ID           string
	Nickname     *string
	Avatar       *string
	Phone        *string
	Password     *string
	PasswordHash *string
	Status       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type MobileUserListIn struct {
	*commonv1.PageRequest
	Phone   *string
	Status  *string
	OrderBy []*helper.OrderBy
}

type MobileUserListOut struct {
	*commonv1.PageResponse
	List []*MobileUser
}
