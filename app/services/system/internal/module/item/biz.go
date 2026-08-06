package item

import (
	"context"
	"log/slog"
	"time"

	"github.com/looplab/fsm"

	"cyber-ecosystem/shared-go/utils"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Item struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        *string
	Description *string
	Status      *string
}

type ItemListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	ID      *string
	Name    *string
	Status  *string
}

type ItemListOut struct {
	*commonpb.PageResponse
	List []*Item
}

// Port ----------------------------------------------------------------------------------------------------------------

type ItemRP interface {
	Create(ctx context.Context, a *Item) (*Item, error)
	Update(ctx context.Context, fieldsMask []string, a *Item) (*Item, error)
	UpdateStatus(ctx context.Context, id, status string) (*Item, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *ItemListIn) (*ItemListOut, error)
	Get(ctx context.Context, id string) (*Item, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*Item, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type ItemUC struct {
	shared.UC
	itemRP ItemRP
}

func NewItemUC(logger *slog.Logger, tm shared.Transaction, itemRP ItemRP) *ItemUC {
	return &ItemUC{
		UC:     shared.NewUC(logger.With("module", "module/item"), tm),
		itemRP: itemRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ItemUC) Create(ctx context.Context, a *Item) (out *Item, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.itemRP.Create(ctx, a)
		return err
	})
	return
}

func (uc *ItemUC) Update(ctx context.Context, fieldsMask []string, a *Item) (out *Item, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.itemRP.Update(ctx, fieldsMask, a)
		return err
	})
	return
}

func (uc *ItemUC) UpdateStatus(ctx context.Context, id, status string) (out *Item, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		a, e := uc.itemRP.Get(ctx, id)
		if e != nil {
			return e
		}
		if e = a.TransitionTo(ctx, status); e != nil {
			return e
		}
		out, e = uc.itemRP.UpdateStatus(ctx, id, status)
		return e
	})
	return
}

func (uc *ItemUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.itemRP.Delete(ctx, id)
		return err
	})
	return
}

func (uc *ItemUC) List(ctx context.Context, in *ItemListIn) (*ItemListOut, error) {
	return uc.itemRP.List(ctx, in)
}

func (uc *ItemUC) Get(ctx context.Context, id string) (*Item, error) {
	return uc.itemRP.Get(ctx, id)
}

func (uc *ItemUC) Sort(ctx context.Context, id string, prevID, nextID *string) (out *Item, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.itemRP.Sort(ctx, id, prevID, nextID)
		return err
	})
	return
}

// Private -------------------------------------------------------------------------------------------------------------

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

func newItemFSM(current string, a *Item) *fsm.FSM {
	return fsm.NewFSM(
		current,
		[]fsm.EventDesc{
			{Name: StatusInactive, Src: []string{StatusActive}, Dst: StatusInactive},
			{Name: StatusActive, Src: []string{StatusInactive}, Dst: StatusActive},
		},
		map[string]fsm.Callback{
			"after_" + StatusInactive: func(_ context.Context, _ *fsm.Event) { a.Status = utils.Ptr(StatusInactive) },
			"after_" + StatusActive:   func(_ context.Context, _ *fsm.Event) { a.Status = utils.Ptr(StatusActive) },
		},
	)
}

func (a *Item) TransitionTo(ctx context.Context, target string) error {
	current := utils.Deref(a.Status, StatusActive)
	if current == target {
		return nil
	}
	f := newItemFSM(current, a)
	if err := f.Event(ctx, target); err != nil {
		return systempb.ErrorErrorReasonInvalidStatusTransition("").WithCause(err)
	}
	return nil
}
