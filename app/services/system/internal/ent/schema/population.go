package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
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
		field.String("plane").NotEmpty().MaxRuneLen(16).
			Comment("compile-time vocabulary: backend (organizational grants) | member (engine-denied consumer seats)"),
	}
}

func (Population) Edges() []ent.Edge {
	return nil
}

// No SoftDelete: a deleted row frees the code unique.
func (Population) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.AuthzStampMixin{},
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
