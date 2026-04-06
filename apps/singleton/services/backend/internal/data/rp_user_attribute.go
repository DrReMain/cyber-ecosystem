package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	"cyber-ecosystem/apps/singleton/services/backend/internal/data/ent"
	entuserattribute "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/userattribute"
)

type userAttributeRP struct {
	RP
}

func NewUserAttributeRP(logger log.Logger, store *Store) biz.UserAttributeRP {
	return &userAttributeRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_user_attribute")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *userAttributeRP) Set(ctx context.Context, attr *biz.UserAttribute) error {
	client := rp.store.GetClient(ctx)
	// Try to find existing attribute
	existing, err := client.UserAttribute.Query().
		Where(
			entuserattribute.UserIDEQ(*attr.UserID),
			entuserattribute.KeyEQ(*attr.Key),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create new
			if _, err := client.UserAttribute.Create().
				SetUserID(*attr.UserID).
				SetKey(*attr.Key).
				SetValue(*attr.Value).
				Save(ctx); err != nil {
				return HandleError(err)
			}
			return nil
		}
		return HandleError(err)
	}
	// Update existing
	if _, err := client.UserAttribute.UpdateOneID(existing.ID).
		SetValue(*attr.Value).
		Save(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *userAttributeRP) Remove(ctx context.Context, userID, key string) error {
	if _, err := rp.store.GetClient(ctx).UserAttribute.Delete().
		Where(
			entuserattribute.UserIDEQ(userID),
			entuserattribute.KeyEQ(key),
		).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *userAttributeRP) Query(ctx context.Context, userID string) ([]*biz.UserAttribute, error) {
	attrs, err := rp.store.GetClient(ctx).UserAttribute.Query().
		Where(entuserattribute.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return utils.SliceMap(attrs, mapUserAttribute), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapUserAttribute(a *ent.UserAttribute) *biz.UserAttribute {
	return &biz.UserAttribute{
		ID:        &a.ID,
		CreatedAt: &a.CreatedAt,
		UpdatedAt: &a.UpdatedAt,
		UserID:    &a.UserID,
		Key:       &a.Key,
		Value:     &a.Value,
	}
}
