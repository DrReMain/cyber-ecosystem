package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type AuthzStamp struct {
	ent.Schema
}

func (AuthzStamp) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Default("policy").Immutable().MaxLen(64).
			Comment("fixed singleton key; the row self-materializes on the first bump"),
		field.Int64("version").Default(0).
			Comment("monotonic engine truth; bumped inside the change transaction by the local bump mixin"),
	}
}

func (AuthzStamp) Edges() []ent.Edge {
	return nil
}

func (AuthzStamp) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.CreatedUpdatedMixin{},
	}
}
