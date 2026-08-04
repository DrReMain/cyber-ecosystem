package biz

import (
	"context"
	"log/slog"
	"time"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Dept struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantID  string
	Name      *string
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
		out, err = uc.deptRP.Create(ctx, d)
		return err
	})
	return
}

func (uc *DeptUC) Update(ctx context.Context, fieldsMask []string, d *Dept) (out *Dept, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		out, err = uc.deptRP.Update(ctx, fieldsMask, d)
		return err
	})
	return
}

func (uc *DeptUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
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
