package biz

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/utils"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Port --------------------------------------------------------------------------------------------------------

type MobileUserRP interface {
	Create(ctx context.Context, u *MobileUser) (*MobileUser, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	Get(ctx context.Context, id string) (*MobileUser, error)
	List(ctx context.Context, in *MobileUserListIn) (*MobileUserListOut, error)
	Update(ctx context.Context, fieldsMask []string, u *MobileUser) (*MobileUser, error)
	UpdateStatus(ctx context.Context, id, status string) (*MobileUser, error)
	Delete(ctx context.Context, id string) (string, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*MobileUser, error)
}

// UC --------------------------------------------------------------------------------------------------------

type MobileUserUC struct {
	UC
	mobileUserRP MobileUserRP
}

func NewMobileUserUC(logger *slog.Logger, tm Transaction, mobileUserRP MobileUserRP) *MobileUserUC {
	return &MobileUserUC{
		UC: UC{
			log: logger.With("module", "biz/mobile_user_uc"),
			tm:  tm,
		},
		mobileUserRP: mobileUserRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------

func (uc *MobileUserUC) Create(ctx context.Context, u *MobileUser) (out *MobileUser, err error) {
	if u.Password != nil {
		hash, hErr := utils.Hash(*u.Password)
		if hErr != nil {
			return nil, errorspb.ErrorGeneralErrorInternal("")
		}
		u.PasswordHash = &hash
	}
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		exists, e := uc.mobileUserRP.ExistsByPhone(ctx, *u.Phone)
		if e != nil {
			return e
		}
		if exists {
			return mobilepb.ErrorErrorReasonMobileUserAlreadyExists("")
		}
		out, err = uc.mobileUserRP.Create(ctx, u)
		return err
	})
	return
}

func (uc *MobileUserUC) Get(ctx context.Context, id string) (*MobileUser, error) {
	return uc.mobileUserRP.Get(ctx, id)
}

func (uc *MobileUserUC) List(ctx context.Context, in *MobileUserListIn) (*MobileUserListOut, error) {
	return uc.mobileUserRP.List(ctx, in)
}

func (uc *MobileUserUC) Update(ctx context.Context, fieldsMask []string, u *MobileUser) (out *MobileUser, err error) {
	if u.Password != nil {
		hash, hErr := utils.Hash(*u.Password)
		if hErr != nil {
			return nil, errorspb.ErrorGeneralErrorInternal("")
		}
		u.PasswordHash = &hash
	}
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.mobileUserRP.Update(ctx, fieldsMask, u)
		return err
	})
	return
}

func (uc *MobileUserUC) UpdateStatus(ctx context.Context, id, status string) (out *MobileUser, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		u, e := uc.mobileUserRP.Get(ctx, id)
		if e != nil {
			return e
		}
		if e = u.TransitionTo(ctx, status); e != nil {
			return e
		}
		out, e = uc.mobileUserRP.UpdateStatus(ctx, id, status)
		return e
	})
	return
}

func (uc *MobileUserUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.mobileUserRP.Delete(ctx, id)
		return err
	})
	return
}

func (uc *MobileUserUC) Sort(ctx context.Context, id string, prevID, nextID *string) (out *MobileUser, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.mobileUserRP.Sort(ctx, id, prevID, nextID)
		return err
	})
	return
}
