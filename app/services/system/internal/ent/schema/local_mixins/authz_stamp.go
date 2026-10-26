package local_mixins

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/mixin"

	gen "cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/authzstamp"
)

const (
	StampRowKey        = "policy" // the singleton row every bump upserts; readers use the same key
	userFieldEnabled   = "enabled"
	userFieldDeletedAt = "deleted_at"
)

// The engine's version truth is a single row bumped inside the change
// transaction itself: the bump is a property of the tables feeding the
// snapshot, not a duty of every writer — seed jobs, admin paths, and
// future code all cover automatically. The upsert rides the mutation's
// own client, which is transaction-bound when the caller wrote inside
// InTx; a single-statement writer leaves a millisecond window that the
// periodic reconcile and the forced-rebuild cap heal. A soft delete may
// re-enter the pipeline and bump twice — harmless: the stamp only needs
// to move, never to move by exactly one.

type AuthzStampMixin struct {
	mixin.Schema
}

func (AuthzStampMixin) Fields() []ent.Field {
	return nil
}

func (AuthzStampMixin) Hooks() []gen.Hook {
	return []gen.Hook{
		func(next gen.Mutator) gen.Mutator {
			return gen.MutateFunc(func(ctx context.Context, m gen.Mutation) (gen.Value, error) {
				if !stampRelevant(m) {
					return next.Mutate(ctx, m)
				}
				v, err := next.Mutate(ctx, m)
				if err != nil {
					return v, err
				}
				mc, ok := m.(interface{ Client() *gen.Client })
				if !ok {
					return v, fmt.Errorf("authz stamp bump: mutation without a client: %T", m)
				}
				if berr := mc.Client().AuthzStamp.Create().
					SetID(StampRowKey).
					SetVersion(1).
					OnConflict(
						sql.ConflictColumns(authzstamp.FieldID),
					).
					Update(func(u *gen.AuthzStampUpsert) {
						u.AddVersion(1)
						u.UpdateUpdatedAt()
					}).
					Exec(ctx); berr != nil {
					return v, fmt.Errorf("authz stamp bump: %w", berr)
				}
				return v, nil
			})
		},
	}
}

func stampRelevant(m gen.Mutation) bool {
	if m.Type() != gen.TypeUser {
		return true
	}
	// User rows change for many non-authz reasons (password, profile);
	// the snapshot only moves when a user crosses the disabled set or
	// disappears, so the bump fires on the enabled field or a delete —
	// never on profile-only edits, which would storm every replica.
	switch m.Op() {
	case gen.OpDelete, gen.OpDeleteOne:
		return true
	}
	// Field (not FieldCleared — that is the SET-NULL path): a value
	// assigned to enabled, or the soft-delete marker being set, moves the
	// snapshot; every other user edit leaves the stamp alone.
	if _, ok := m.Field(userFieldEnabled); ok {
		return true
	}
	if _, ok := m.Field(userFieldDeletedAt); ok {
		return true
	}
	// A restore clears the marker — the SET-NULL path: the user re-enters
	// the snapshot, so the un-delete moves the stamp too.
	return m.FieldCleared(userFieldDeletedAt)
}
