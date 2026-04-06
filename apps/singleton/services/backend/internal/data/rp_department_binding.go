package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/singleton/services/backend/internal/biz"
	entdepartment "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/department"
	entuser "cyber-ecosystem/apps/singleton/services/backend/internal/data/ent/user"
)

type departmentBindingRP struct {
	RP
}

func NewDepartmentBindingRP(logger log.Logger, store *Store) biz.DepartmentBindingRP {
	return &departmentBindingRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_department_binding")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *departmentBindingRP) SetUserDepartments(ctx context.Context, userID string, departmentIDs []string) error {
	if _, err := rp.store.GetClient(ctx).User.UpdateOneID(userID).
		ClearDepartments().
		AddDepartmentIDs(departmentIDs...).
		Save(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *departmentBindingRP) ListUserDepartments(ctx context.Context, userID string) ([]*biz.Department, error) {
	depts, err := rp.store.GetClient(ctx).User.Query().
		Where(entuser.ID(userID)).
		QueryDepartments().
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return utils.SliceMap(depts, mapDepartment), nil
}

func (rp *departmentBindingRP) ListDepartmentUsers(ctx context.Context, departmentID string) ([]string, error) {
	users, err := rp.store.GetClient(ctx).Department.Query().
		Where(entdepartment.ID(departmentID)).
		QueryUsers().
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids, nil
}
