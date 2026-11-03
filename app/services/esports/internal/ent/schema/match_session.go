package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/esports/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
)

type MatchSession struct {
	ent.Schema
}

func (MatchSession) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("owning member identity (system user id), logical reference — same key law as profile and test_record"),
		field.String("genre").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("the session's type-axis value; frozen at creation — changing type is a new session, never a mutation"),
		field.Enum("difficulty").Values("experience", "standard").Default("experience").
			Comment("mode axis: every attempt in the session scores on this mode at the submit seam; every session is born experience — standard is reserved for a purchasable entry"),
		field.Int64("seed").Default(0).
			Comment("the draw's input entropy (FNV over user|day|genre|salt), stored for audit reproduction only — behavior always replays the locked steps"),
		field.JSON("steps", []types.MatchStep{}).
			Comment("the locked arrangement in match_config render_order; frozen at creation, resumes replay it verbatim"),
		field.Int("retest_budget").Default(3).
			Comment("attempts allowed per step inside this session — first-ever session 3, later sessions 1; frozen at birth, each stored attempt spends one valid or not"),
		field.Enum("status").Values("active", "switched", "expired").Default("active").
			Comment("active resumes at any age — the lineup is locked in the row; switched = an explicit re-draw closed it; expired = the retired seven-day window close (legacy rows only, no longer produced)"),
	}
}

func (MatchSession) Edges() []ent.Edge {
	return nil
}

func (MatchSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.SoftDeleteMixin{},
	}
}

func (MatchSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("user_id", "created_at"),
		index.Fields("user_id").Unique().
			Annotations(entsql.IndexWhere("status = 'active' AND deleted_at IS NULL")),
	}
}

func (MatchSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "match_session"},
	}
}
