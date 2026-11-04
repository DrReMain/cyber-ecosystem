package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Announcement is the operator's broadcast draft and its one-way
// lifecycle: draft → published → offline, never back. Publishing is
// the delivery moment — a message-kind row is copied into every
// targeted member's letterbox in the same transaction the status
// flips, so "published" and "delivered" cannot disagree; offline
// retires a banner and stops nothing already sent. The member set is
// resolved at publish time: a snapshot, like a letter mailed on paper,
// not a living filter. Operator-plane row, so no tenant column and no
// tenant filtering — the same global shape as agreement_version.
type Announcement struct {
	ent.Schema
}

func (Announcement) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("kind").Values("banner", "message").Immutable().
			Comment("banner = live on the home surface; message = delivered into letterboxes at publish"),
		field.Enum("scope").Values("all", "membership").Immutable().
			Comment("who receives a message-kind row at publish: every roster member, or live membership holders"),
		field.String("title").NotEmpty().MaxLen(64).
			Comment("the headline both surfaces render"),
		field.String("content").NotEmpty().MaxLen(5000).
			Comment("the body; editable while draft, frozen in every delivered copy at publish"),
		field.Enum("status").Values("draft", "published", "offline").Default("draft").
			Comment("the one-way lifecycle: draft→published→offline, no revival"),
		field.Time("published_at").Optional().Nillable().
			Comment("absent until published — the delivery moment"),
		field.Time("offline_at").Optional().Nillable().
			Comment("absent until offlined"),
		field.String("published_by").MaxLen(20).Optional().Nillable().
			Comment("the operator who published; drafts carry none"),
	}
}

func (Announcement) Edges() []ent.Edge {
	return nil
}

func (Announcement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Announcement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kind", "status", "published_at"),
	}
}

func (Announcement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "announcement"},
	}
}
