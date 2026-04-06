package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// Model ---------------------------------------------------------------------------------------------------------------

type WorkReport struct {
	ID           *string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Title        *string
	Content      *string
	Type         *string
	DepartmentID *string
	AccessLevel  *int
	Region       *string
	CreatedBy    *string
	Status       *string
}

type WorkReportQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	Type    *string
	Status  *string
}

type WorkReportQueryOut struct {
	*common.PageResponse
	List []*WorkReport
}

// Port ----------------------------------------------------------------------------------------------------------------

type WorkReportRP interface {
	Create(ctx context.Context, report *WorkReport) (*WorkReport, error)
	Update(ctx context.Context, report *WorkReport, fieldsMask []string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*WorkReport, error)
	Query(ctx context.Context, in *WorkReportQueryIn) (*WorkReportQueryOut, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type WorkReportUC struct {
	UC
	workReportRP WorkReportRP
}

func NewWorkReportUC(logger log.Logger, tm Transaction, workReportRP WorkReportRP) *WorkReportUC {
	return &WorkReportUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_work_report")),
			tm:  tm,
		},
		workReportRP: workReportRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *WorkReportUC) Create(ctx context.Context, report *WorkReport) (*WorkReport, error) {
	return uc.workReportRP.Create(ctx, report)
}

func (uc *WorkReportUC) Update(ctx context.Context, report *WorkReport, fieldsMask []string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.workReportRP.Update(ctx, report, fieldsMask)
	})
}

func (uc *WorkReportUC) Delete(ctx context.Context, id string) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.workReportRP.Delete(ctx, id)
	})
}

func (uc *WorkReportUC) Get(ctx context.Context, id string) (*WorkReport, error) {
	return uc.workReportRP.Get(ctx, id)
}

func (uc *WorkReportUC) Query(ctx context.Context, in *WorkReportQueryIn) (*WorkReportQueryOut, error) {
	return uc.workReportRP.Query(ctx, in)
}
