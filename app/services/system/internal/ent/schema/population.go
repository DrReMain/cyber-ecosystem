package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Population struct {
	ent.Schema
}

func (Population) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().MaxRuneLen(20).Immutable().
			Match(regexp.MustCompile(codeJoinKeyPattern)).
			Comment("join key: membership rows carry this value in population"),
		field.String("name").NotEmpty().MaxRuneLen(128),
		field.String("archetype").NotEmpty().MaxRuneLen(16).
			Comment("compile-time vocabulary: engine (organizational grants) | patron (identity-is-permission)"),
	}
}

func (Population) Edges() []ent.Edge {
	return nil
}

// No SoftDelete: a deleted row must not hold the code unique hostage.
func (Population) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Population) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

func (Population) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "population"},
	}
}
