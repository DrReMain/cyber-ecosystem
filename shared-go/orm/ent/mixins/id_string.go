package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/rs/xid"
)

type IDStringMixin struct {
	mixin.Schema
}

func (IDStringMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(20).
			DefaultFunc(func() string {
				return xid.New().String()
			}),
	}
}

func (IDStringMixin) Indexes() []ent.Index {
	return []ent.Index{}
}
