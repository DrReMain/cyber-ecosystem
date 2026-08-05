package biz

import (
	"context"
	"errors"
	"log/slog"
	"time"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Dept struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantID  string
	Name      *string
	ParentID  *string
}

type DeptListIn struct {
	*commonpb.PageRequest
	OrderBy []string
	Name    *string
}

type DeptListOut struct {
	*commonpb.PageResponse
	List []*Dept
}

// Port ----------------------------------------------------------------------------------------------------------------

type DeptRP interface {
	Create(ctx context.Context, d *Dept) (*Dept, error)
	Update(ctx context.Context, fieldsMask []string, d *Dept) (*Dept, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *DeptListIn) (*DeptListOut, error)
	Get(ctx context.Context, id string) (*Dept, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	IsAncestor(ctx context.Context, ancestorID, descendantID string) (bool, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type DeptUC struct {
	UC
	deptRP DeptRP
}

func NewDeptUC(logger *slog.Logger, tm Transaction, deptRP DeptRP) *DeptUC {
	return &DeptUC{
		UC:     UC{log: logger.With("module", "biz/dept"), tm: tm},
		deptRP: deptRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *DeptUC) Create(ctx context.Context, d *Dept) (out *Dept, err error) {
	d.TenantID = defaultTenant
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		// parent existence + same-tenant (the latter enforced by intercept on Get)
		if d.ParentID != nil {
			if _, e := uc.deptRP.Get(ctx, *d.ParentID); e != nil {
				return e
			}
		}
		out, err = uc.deptRP.Create(ctx, d)
		return err
	})
	return
}

func (uc *DeptUC) Update(ctx context.Context, fieldsMask []string, d *Dept) (out *Dept, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		if d.ParentID != nil {
			if *d.ParentID == d.ID {
				return errorspb.ErrorGeneralErrorValidationFailed("").WithCause(errors.New("dept cannot be its own parent"))
			}
			isAnc, e := uc.deptRP.IsAncestor(ctx, d.ID, *d.ParentID)
			if e != nil {
				return e
			}
			if isAnc {
				return errorspb.ErrorGeneralErrorValidationFailed("").WithCause(errors.New("dept parent would form a cycle"))
			}
			if _, e := uc.deptRP.Get(ctx, *d.ParentID); e != nil {
				return e
			}
		}
		out, err = uc.deptRP.Update(ctx, fieldsMask, d)
		return err
	})
	return
}

func (uc *DeptUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		has, e := uc.deptRP.HasChildren(ctx, id)
		if e != nil {
			return e
		}
		if has {
			return errorspb.ErrorGeneralErrorValidationFailed("").WithCause(errors.New("dept has children, cannot be deleted"))
		}
		out, err = uc.deptRP.Delete(ctx, id)
		return err
	})
	return
}

func (uc *DeptUC) List(ctx context.Context, in *DeptListIn) (*DeptListOut, error) {
	return uc.deptRP.List(ctx, in)
}

func (uc *DeptUC) Get(ctx context.Context, id string) (*Dept, error) {
	return uc.deptRP.Get(ctx, id)
}
