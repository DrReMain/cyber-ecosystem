package mixins

import (
	"time"

	"github.com/rs/xid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
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
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (IDStringMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
		index.Fields("updated_at"),
	}
}
