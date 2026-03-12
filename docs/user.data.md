```go
package data

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/kratos/system-service/internal/data/ent/user"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/iancoleman/strcase"
)

type userRP struct {
	data *Data
	log  *log.Helper
}

func NewUserRP(data *Data, logger log.Logger) biz.UserRP {
	return &userRP{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func fieldTransform(paths []string) []string {
	res := make([]string, 0, len(paths))
	for _, p := range paths {
		res = append(res, strcase.ToSnake(p))
	}
	return res
}

func (rp *userRP) Create(ctx context.Context, u *biz.User) (*biz.User, error) {
	row, err := rp.data.db.User.Create().
		SetUsername(u.Username).
		SetEmail(u.Email).
		SetAge(u.Age).
		SetPassword(u.Password).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return rp.mapToBiz(row), nil
}

func (rp *userRP) Update(ctx context.Context, u *biz.User, fields []string) (*biz.User, error) {
	updater := rp.data.db.User.UpdateOneID(u.ID)
	mask := make(map[string]struct{})
	for _, f := range fields {
		mask[strcase.ToSnake(f)] = struct{}{}
	}

	if _, ok := mask["username"]; ok {
		updater.SetUsername(u.Username)
	}
	if _, ok := mask["email"]; ok {
		updater.SetEmail(u.Email)
	}
	if _, ok := mask["age"]; ok {
		updater.SetAge(u.Age)
	}

	row, err := updater.Save(ctx)
	if err != nil {
		return nil, err
	}
	return rp.mapToBiz(row), nil
}

func (rp *userRP) DeleteBatch(ctx context.Context, ids []string) error {
	_, err := rp.data.db.User.Delete().Where(user.IDIn(ids...)).Exec(ctx)
	return err
}

func (rp *userRP) Delete(ctx context.Context, s string) error {
	return rp.data.db.User.DeleteOneID(s).Exec(ctx)
}

func (rp *userRP) Get(ctx context.Context, id string) (*biz.User, error) {
	row, err := rp.data.db.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return rp.mapToBiz(row), nil
}

func (rp *userRP) Query(ctx context.Context, opt *biz.UserQueryOption) ([]*biz.User, int64, error) {
	q := rp.data.db.User.Query()
	if opt.Email != "" {
		q.Where(user.EmailContains(opt.Email))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if len(opt.Fields) > 0 {
		q.Select(fieldTransform(opt.Fields)...)
	}

	rows, err := q.Offset(int((opt.PageNo - 1) * opt.PageSize)).
		Limit(int(opt.PageSize)).
		Order(ent.Desc(user.FieldCreatedAt)).
		All(ctx)

	return rp.mapListToBiz(rows), int64(total), err
}

func (rp *userRP) List(ctx context.Context, opt *biz.UserQueryOption) ([]*biz.User, error) {
	q := rp.data.db.User.Query()
	if len(opt.Fields) > 0 {
		q.Select(fieldTransform(opt.Fields)...)
	}
	rows, err := q.All(ctx)
	return rp.mapListToBiz(rows), err
}

func (rp *userRP) mapToBiz(x *ent.User) *biz.User {
	return &biz.User{
		ID:        x.ID,
		Username:  x.Username,
		Email:     *x.Email,
		Age:       *x.Age,
		CreatedAt: x.CreatedAt,
		UpdatedAt: x.UpdatedAt,
	}
}

func (rp *userRP) mapListToBiz(x []*ent.User) []*biz.User {
	res := make([]*biz.User, 0, len(x))
	for _, v := range x {
		res = append(res, rp.mapToBiz(v))
	}
	return res
}

```
