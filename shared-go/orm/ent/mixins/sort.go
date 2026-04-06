package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"roci.dev/fracdex"
)

type SortMixin struct {
	mixin.Schema
}

func (SortMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("sort").
			DefaultFunc(func() string {
				key, _ := fracdex.KeyBetween("", "")
				return key
			}).
			Comment("fractional index for ordering"),
	}
}
