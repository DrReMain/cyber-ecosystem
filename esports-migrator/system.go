package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/xid"
)

const (
	// Seat vocabulary for migrated C-end members: the default tenant and
	// the member population the wechat/sms providers auto-register into.
	migrTenant     = "default"
	migrPopulation = "member"

	kindSMS    = "sms"       // identifier = phone number
	kindWechat = "wechat_mp" // identifier = wechat openid
	entityUser = "user"
)

func insertSystemUser(ctx context.Context, tx *sql.Tx, p *userPlan) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO "user" (id, created_at, updated_at, deleted_at, account, password_hash, dept_id, avatar, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, NULL, $7)
		ON CONFLICT (id) DO NOTHING`,
		p.srcID(), p.src.CreatedAt.UTC(), p.src.UpdatedAt.UTC(), p.deletedAt(), p.account, p.passwordHash, p.enabled())
	if err != nil {
		return false, fmt.Errorf("insert user %s: %w", p.srcID(), err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// insertSystemCredential binds one external identifier. Soft-deleted
// identities get no credentials: the credential table has no soft delete —
// writing one would occupy its unique key and block a future real
// registration on the same phone or openid.
func insertSystemCredential(ctx context.Context, tx *sql.Tx, kind, identifier, userID string, createdAt time.Time) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO "credential" (id, created_at, updated_at, kind, identifier, user_id)
		VALUES ($1, $2, $2, $3, $4, $5)
		ON CONFLICT (kind, identifier) DO NOTHING`,
		xid.New().String(), createdAt.UTC(), kind, identifier, userID)
	if err != nil {
		return false, fmt.Errorf("insert credential %s/%s: %w", kind, identifier, err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func insertSystemMembership(ctx context.Context, tx *sql.Tx, userID string, createdAt time.Time) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO "membership" (id, created_at, updated_at, user_id, tenant_id, population)
		VALUES ($1, $2, $2, $3, $4, $5)
		ON CONFLICT (user_id, tenant_id, population) DO NOTHING`,
		xid.New().String(), createdAt.UTC(), userID, migrTenant, migrPopulation)
	if err != nil {
		return false, fmt.Errorf("insert membership %s: %w", userID, err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
