package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Model ---------------------------------------------------------------------------------------------------------------

// UserAttribute represents a key-value attribute on a user, used for ABAC rules.
type UserAttribute struct {
	ID        *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	UserID    *string
	Key       *string
	Value     *string
}

// Port ----------------------------------------------------------------------------------------------------------------

// UserAttributeRP abstracts user attribute persistence.
type UserAttributeRP interface {
	Set(ctx context.Context, attr *UserAttribute) error
	Remove(ctx context.Context, userID, key string) error
	Query(ctx context.Context, userID string) ([]*UserAttribute, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

// UserAttributeUC handles user attribute CRUD with scope cache invalidation.
type UserAttributeUC struct {
	UC
	userAttrRP  UserAttributeRP
	invalidator ScopeCacheInvalidator
}

// NewUserAttributeUC creates a new UserAttributeUC.
func NewUserAttributeUC(
	logger log.Logger,
	tm Transaction,
	userAttrRP UserAttributeRP,
	invalidator ScopeCacheInvalidator,
) *UserAttributeUC {
	return &UserAttributeUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_user_attribute")),
			tm:  tm,
		},
		userAttrRP:  userAttrRP,
		invalidator: invalidator,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

// Set sets a key-value attribute on a user and invalidates their scope cache.
func (uc *UserAttributeUC) Set(ctx context.Context, attr *UserAttribute) error {
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		if err := uc.userAttrRP.Set(txCtx, attr); err != nil {
			return err
		}
		if err := uc.invalidator.InvalidateUser(txCtx, *attr.UserID); err != nil {
			uc.log.Warnf("failed to invalidate scope cache for user %s: %v", *attr.UserID, err)
		}
		return nil
	})
}

// Remove removes a user attribute and invalidates their scope cache.
func (uc *UserAttributeUC) Remove(ctx context.Context, userID, key string) error {
	return uc.tm.InTx(ctx, func(txCtx context.Context) error {
		if err := uc.userAttrRP.Remove(txCtx, userID, key); err != nil {
			return err
		}
		if err := uc.invalidator.InvalidateUser(txCtx, userID); err != nil {
			uc.log.Warnf("failed to invalidate scope cache for user %s: %v", userID, err)
		}
		return nil
	})
}

// Query returns all attributes for a user.
func (uc *UserAttributeUC) Query(ctx context.Context, userID string) ([]*UserAttribute, error) {
	return uc.userAttrRP.Query(ctx, userID)
}
