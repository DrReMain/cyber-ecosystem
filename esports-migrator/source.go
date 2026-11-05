package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// openSource dials the legacy MySQL store. The DSN is forced into
// parseTime with loc=Asia/Shanghai: legacy DATETIME columns are bare wall
// clocks in that zone, and the driver must read them as such before the
// writer converts to UTC for the timestamptz targets.
func openSource(dsn string) (*sql.DB, error) {
	if !strings.Contains(dsn, "loc=") {
		dsn += dsnParamSep(dsn) + "loc=Asia%2FShanghai"
	}
	if !strings.Contains(dsn, "parseTime=") {
		dsn += dsnParamSep(dsn) + "parseTime=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := ping(db, "source"); err != nil {
		return nil, err
	}
	return db, nil
}

func openPostgres(name, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s target: %w", name, err)
	}
	if err := ping(db, name); err != nil {
		return nil, err
	}
	return db, nil
}

func ping(db *sql.DB, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping %s: %w", name, err)
	}
	return nil
}

// dsnParamSep picks the separator for appending a parameter: a bare DSN
// (".../legacy") glued with "&" would fold the parameter into the database
// name.
func dsnParamSep(dsn string) string {
	if strings.Contains(dsn, "?") {
		return "&"
	}
	return "?"
}

// srcUser carries only the identity-phase columns of the legacy user row;
// later phases re-query their own column sets.
type srcUser struct {
	ID                int64
	Username          sql.NullString
	Password          sql.NullString
	PasswordAlgorithm string
	Phone             sql.NullString
	OpenID            sql.NullString
	WechatSex         sql.NullInt64
	WechatCountry     sql.NullString
	WechatProvince    sql.NullString
	WechatCity        sql.NullString
	Nickname          sql.NullString // prod copy: NULL despite init.sql's NOT NULL claim (9/285 rows)
	AgeRange          sql.NullString
	BirthDate         sql.NullTime
	AvatarURL         sql.NullString
	SchoolID          sql.NullInt64
	Status            int64
	IsDeleted         bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

const userColumns = `id, username, password, password_algorithm, phone, open_id,
	wechat_sex, wechat_country, wechat_province, wechat_city, nickname,
	age_range, birth_date, avatar_url, school_id, status, is_deleted,
	created_at, updated_at`

func fetchUsers(ctx context.Context, db *sql.DB) ([]*srcUser, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT "+userColumns+" FROM `user` ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query legacy user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*srcUser
	for rows.Next() {
		u := &srcUser{}
		var isDeleted int64
		err := rows.Scan(
			&u.ID, &u.Username, &u.Password, &u.PasswordAlgorithm, &u.Phone, &u.OpenID,
			&u.WechatSex, &u.WechatCountry, &u.WechatProvince, &u.WechatCity, &u.Nickname,
			&u.AgeRange, &u.BirthDate, &u.AvatarURL, &u.SchoolID, &u.Status, &isDeleted,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan legacy user %d: %w", u.ID, err)
		}
		u.IsDeleted = isDeleted != 0
		out = append(out, u)
	}
	return out, rows.Err()
}

// schoolRowCounts feeds the guard: the organization layer was dissolved on
// the premise that no production school data exists, so any non-empty
// school table halts the run for a human decision instead of guessing.
func schoolRowCounts(ctx context.Context, db *sql.DB) (school, schoolReport int64, err error) {
	if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `school`").Scan(&school); err != nil {
		return 0, 0, fmt.Errorf("count school: %w", err)
	}
	if err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `school_report`").Scan(&schoolReport); err != nil {
		return 0, 0, fmt.Errorf("count school_report: %w", err)
	}
	return school, schoolReport, nil
}
