package dept

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-kratos/kratos/v3/errors"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Dept struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	TenantID  string
	Name      *string
	ParentID  *string
	Remark    *string
	UserCount int64
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

type DeptMember struct {
	UserID  string
	Email   string
	Enabled bool
}

type MemberListIn struct {
	DeptID string
	*commonpb.PageRequest
}

type MemberListOut struct {
	*commonpb.PageResponse
	List []*DeptMember
}

// Port ----------------------------------------------------------------------------------------------------------------

type DeptRP interface {
	Create(ctx context.Context, d *Dept) (*Dept, error)
	Update(ctx context.Context, fieldsMask []string, d *Dept) (*Dept, error)
	Delete(ctx context.Context, id string) (string, error)
	List(ctx context.Context, in *DeptListIn) (*DeptListOut, error)
	Get(ctx context.Context, id string) (*Dept, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	HasUsers(ctx context.Context, id string) (bool, error)
	IsAncestor(ctx context.Context, ancestorID, descendantID string) (bool, error)
	RemoveMember(ctx context.Context, deptID, userID string) (bool, error)
	ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type DeptUC struct {
	shared.UC
	deptRP DeptRP
}

func NewDeptUC(logger *slog.Logger, tm shared.Transaction, deptRP DeptRP) *DeptUC {
	return &DeptUC{
		UC:     shared.NewUC(logger.With("module", "module/dept"), tm),
		deptRP: deptRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *DeptUC) Create(ctx context.Context, d *Dept) (out *Dept, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		if e := uc.ensureParent(ctx, d.ParentID); e != nil {
			return e
		}
		out, err = uc.deptRP.Create(ctx, d)
		return err
	})
	return
}

func (uc *DeptUC) Update(ctx context.Context, fieldsMask []string, d *Dept) (out *Dept, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		// reparenting: only when a new parent is set (d.ParentID != nil). Clearing to root
		// (d.ParentID == nil) has no cycle risk; the repo clears it only when fieldsMask asks.
		if d.ParentID != nil {
			if *d.ParentID == d.ID {
				return systempb.ErrorSystemDeptSelfParent("")
			}
			isAnc, e := uc.deptRP.IsAncestor(ctx, d.ID, *d.ParentID)
			if e != nil {
				return e
			}
			if isAnc {
				return systempb.ErrorSystemDeptCycle("")
			}
			if e := uc.ensureParent(ctx, d.ParentID); e != nil {
				return e
			}
		}
		out, err = uc.deptRP.Update(ctx, fieldsMask, d)
		return err
	})
	return
}

func (uc *DeptUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.Tm.InTx(ctx, func(ctx context.Context) error {
		has, e := uc.deptRP.HasChildren(ctx, id)
		if e != nil {
			return e
		}
		if has {
			return systempb.ErrorSystemDeptHasChildren("")
		}
		hasUsers, e := uc.deptRP.HasUsers(ctx, id)
		if e != nil {
			return e
		}
		if hasUsers {
			return systempb.ErrorSystemDeptHasUsers("")
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

func (uc *DeptUC) RemoveMember(ctx context.Context, deptID, userID string) error {
	if err := uc.ensureDept(ctx, deptID); err != nil {
		return err
	}
	var removed bool
	err := uc.Tm.InTx(ctx, func(ctx context.Context) error {
		var e error
		removed, e = uc.deptRP.RemoveMember(ctx, deptID, userID)
		return e
	})
	if err != nil {
		return err
	}
	if !removed {
		// Absence covers moved-elsewhere, deleted and out-of-scope — one
		// fuzzy not-found, same shape as the missing-row write path.
		return systempb.ErrorSystemDeptMemberNotFound("")
	}
	return nil
}

func (uc *DeptUC) ListMembers(ctx context.Context, in *MemberListIn) (*MemberListOut, error) {
	if err := uc.ensureDept(ctx, in.DeptID); err != nil {
		return nil, err
	}
	return uc.deptRP.ListMembers(ctx, in)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *DeptUC) ensureDept(ctx context.Context, id string) error {
	if _, err := uc.deptRP.Get(ctx, id); err != nil {
		if errors.IsNotFound(err) {
			return systempb.ErrorSystemDeptNotFound("")
		}
		return err
	}
	return nil
}

func (uc *DeptUC) ensureParent(ctx context.Context, parentID *string) error {
	if parentID == nil {
		return nil
	}
	_, err := uc.deptRP.Get(ctx, *parentID)
	return err
}
