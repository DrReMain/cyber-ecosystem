package biz

import (
	"context"
	"log/slog"
	"time"

	"github.com/looplab/fsm"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	corepb "cyber-ecosystem/gen/go/cyber/core/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// Entity --------------------------------------------------------------------------------------------------------

const (
	StatusEnabled    = "enabled"
	StatusDisabled   = "disabled"
	StatusRestricted = "restricted"
)

type User struct {
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

func (u *User) TransitionTo(ctx context.Context, target string) error {
	u.Status = utils.Ptr(utils.Deref(u.Status, StatusEnabled))
	f := newUserFSM(*u.Status, u)
	if err := f.Event(ctx, target); err != nil {
		return corepb.ErrorErrorReasonStatusInvalidTransition("").WithCause(err)
	}
	return nil
}

type UserListIn struct {
	*commonv1.PageRequest
	Phone   *string
	Status  *string
	OrderBy []*helper.OrderBy
}

type UserListOut struct {
	*commonv1.PageResponse
	List []*User
}

// Port --------------------------------------------------------------------------------------------------------

type UserRP interface {
	Create(ctx context.Context, u *User) (*User, error)
	Update(ctx context.Context, fieldsMask []string, u *User) (*User, error)
	UpdateStatus(ctx context.Context, id, status string) (*User, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *UserListIn) (*UserListOut, error)
	Get(ctx context.Context, id string) (*User, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*User, error)
}

// UC --------------------------------------------------------------------------------------------------------

type UserUC struct {
	UC
	userRP UserRP
}

func NewUserUC(logger *slog.Logger, tm Transaction, userRP UserRP) *UserUC {
	return &UserUC{
		UC:     UC{log: logger.With("module", "biz/user"), tm: tm},
		userRP: userRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------

func (uc *UserUC) Create(ctx context.Context, u *User) (out *User, err error) {
	if u.Password != nil {
		hash, hErr := utils.Hash(*u.Password)
		if hErr != nil {
			return nil, errorspb.ErrorGeneralErrorInternal("")
		}
		u.PasswordHash = &hash
	}
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		exists, e := uc.userRP.ExistsByPhone(ctx, *u.Phone)
		if e != nil {
			return e
		}
		if exists {
			return corepb.ErrorErrorReasonUserAlreadyExists("")
		}
		out, err = uc.userRP.Create(ctx, u)
		return err
	})
	return
}

func (uc *UserUC) Update(ctx context.Context, fieldsMask []string, u *User) (out *User, err error) {
	if u.Password != nil {
		hash, hErr := utils.Hash(*u.Password)
		if hErr != nil {
			return nil, errorspb.ErrorGeneralErrorInternal("")
		}
		u.PasswordHash = &hash
	}
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Update(ctx, fieldsMask, u)
		return err
	})
	return
}

func (uc *UserUC) UpdateStatus(ctx context.Context, id, status string) (out *User, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		u, e := uc.userRP.Get(ctx, id)
		if e != nil {
			return e
		}
		if e = u.TransitionTo(ctx, status); e != nil {
			return e
		}
		out, e = uc.userRP.UpdateStatus(ctx, id, status)
		return e
	})
	return
}

func (uc *UserUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Delete(ctx, id)
		return err
	})
	return
}

func (uc *UserUC) List(ctx context.Context, in *UserListIn) (*UserListOut, error) {
	return uc.userRP.List(ctx, in)
}

func (uc *UserUC) Get(ctx context.Context, id string) (*User, error) {
	return uc.userRP.Get(ctx, id)
}

func (uc *UserUC) Sort(ctx context.Context, id string, prevID, nextID *string) (out *User, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.userRP.Sort(ctx, id, prevID, nextID)
		return err
	})
	return
}

// Private --------------------------------------------------------------------------------------------------------

func newUserFSM(current string, u *User) *fsm.FSM {
	return fsm.NewFSM(
		current,
		[]fsm.EventDesc{
			{Name: StatusDisabled, Src: []string{StatusEnabled}, Dst: StatusDisabled},
			{Name: StatusRestricted, Src: []string{StatusEnabled}, Dst: StatusRestricted},
			{Name: StatusEnabled, Src: []string{StatusDisabled, StatusRestricted}, Dst: StatusEnabled},
		},
		map[string]fsm.Callback{
			"after_" + StatusDisabled:   func(_ context.Context, _ *fsm.Event) { *u.Status = StatusDisabled },
			"after_" + StatusRestricted: func(_ context.Context, _ *fsm.Event) { *u.Status = StatusRestricted },
			"after_" + StatusEnabled:    func(_ context.Context, _ *fsm.Event) { *u.Status = StatusEnabled },
		},
	)
}
