package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"

	singletonV1 "cyber-ecosystem/apps/singleton/gen/go/v1"
)

var (
	ErrDepartmentHasChildren      = singletonV1.ErrorErrorReasonDepartmentHasChildren("")
	ErrDepartmentMoveToSelf       = singletonV1.ErrorErrorReasonDepartmentMoveToSelf("")
	ErrDepartmentMoveToDescendant = singletonV1.ErrorErrorReasonDepartmentMoveToDescendant("")
)

// Model ---------------------------------------------------------------------------------------------------------------

type Department struct {
	ID        *string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Name      *string
	Code      *string
	ParentID  *string
	Path      *string
	Status    *uint8
	Sort      *string
}

type DepartmentQueryIn struct {
	*common.PageRequest
	OrderBy  []*utils.OrderBy
	Name     *string
	Code     *string
	ParentID *string
}

type DepartmentQueryOut struct {
	*common.PageResponse
	List []*Department
}

// Port ----------------------------------------------------------------------------------------------------------------

type DepartmentRP interface {
	Create(ctx context.Context, dept *Department) error
	Update(ctx context.Context, fieldsMask []string, dept *Department) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Department, error)
	Query(ctx context.Context, in *DepartmentQueryIn) (*DepartmentQueryOut, error)
	Move(ctx context.Context, id string, targetParent *Department) error
	HasChildren(ctx context.Context, parentID string) (bool, error)
	GetDescendantDeptIDs(ctx context.Context, parentIDs []string) ([]string, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type DepartmentUC struct {
	UC
	deptRP DepartmentRP
}

func NewDepartmentUC(logger log.Logger, tm Transaction, deptRP DepartmentRP) *DepartmentUC {
	return &DepartmentUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_department")),
			tm:  tm,
		},
		deptRP: deptRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *DepartmentUC) Create(ctx context.Context, dept *Department) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.deptRP.Create(ctx, dept)
	})
}

func (uc *DepartmentUC) Update(ctx context.Context, fieldsMask []string, dept *Department) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.deptRP.Update(ctx, fieldsMask, dept)
	})
}

func (uc *DepartmentUC) Delete(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		hasChildren, err := uc.deptRP.HasChildren(ctx, id)
		if err != nil {
			return err
		}
		if hasChildren {
			return ErrDepartmentHasChildren
		}
		return uc.deptRP.Delete(ctx, id)
	})
}

func (uc *DepartmentUC) Get(ctx context.Context, id string) (*Department, error) {
	return uc.deptRP.Get(ctx, id)
}

func (uc *DepartmentUC) Query(ctx context.Context, in *DepartmentQueryIn) (*DepartmentQueryOut, error) {
	return uc.deptRP.Query(ctx, in)
}

func (uc *DepartmentUC) Move(ctx context.Context, id string, targetParentID *string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		dept, err := uc.deptRP.Get(ctx, id)
		if err != nil {
			return err
		}

		if targetParentID == nil {
			return uc.deptRP.Move(ctx, id, nil)
		}

		if *targetParentID == id {
			return ErrDepartmentMoveToSelf
		}

		targetParent, err := uc.deptRP.Get(ctx, *targetParentID)
		if err != nil {
			return err
		}

		currentPath := *dept.Path
		targetPath := *targetParent.Path
		if len(targetPath) >= len(currentPath) && targetPath[:len(currentPath)] == currentPath {
			return ErrDepartmentMoveToDescendant
		}

		return uc.deptRP.Move(ctx, id, targetParent)
	})
}
