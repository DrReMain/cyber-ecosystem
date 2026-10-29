package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type AccountMerge struct {
	ent.Schema
}

func (AccountMerge) Fields() []ent.Field {
	return []ent.Field{
		field.String("poor_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("deactivated account id; unique — an account can be absorbed at most once"),
		field.String("rich_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("surviving account id that took over the credential"),
		field.String("provider").NotEmpty().MaxRuneLen(64).Immutable().
			Comment("provider config key whose exchange proved control"),
		field.String("kind").NotEmpty().MaxRuneLen(16).Immutable().
			Comment("credential kind that moved (credential.kind)"),
		field.String("identifier").NotEmpty().MaxRuneLen(191).Immutable().
			Comment("credential identifier that moved (credential.identifier)"),
	}
}

func (AccountMerge) Edges() []ent.Edge {
	return nil
}

// Identity-plane ledger: the merge is a global-account fact, one row per
// absorption — no tenant, no soft delete, no updates.
func (AccountMerge) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (AccountMerge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("poor_id").Unique(),
		index.Fields("rich_id"),
	}
}

func (AccountMerge) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "account_merges"},
	}
}
