package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// AgreementConsent is one compliance-evidence row: this user agreed to
// that document version at that time on that login path. Append-only by
// ruling — a login writes one row per checked document and nothing ever
// rewrites or deletes a row, because the whole point is answering a
// minor-recharge or refund dispute with what actually happened. The
// version number is denormalized on purpose: deleting the version row
// later must not delete the evidence. Same global shape as
// agreement_version — no tenant column, no soft delete.
type AgreementConsent struct {
	ent.Schema
}

func (AgreementConsent) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20).Immutable().
			Comment("the consenting member identity"),
		field.Enum("doc").Values("user", "privacy").Immutable().
			Comment("the document the checkbox bound to"),
		field.String("version_no").NotEmpty().MaxLen(20).Immutable().
			Comment("the version string at consent time, denormalized so the evidence outlives the version row"),
		field.Enum("login_method").Values("sms_code", "password", "wechat").Immutable().
			Comment("the UI path the consent travelled on, declared by the client (the system owns the real auth facts)"),
	}
}

func (AgreementConsent) Edges() []ent.Edge {
	return nil
}

func (AgreementConsent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (AgreementConsent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("doc", "version_no"),
	}
}

func (AgreementConsent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "agreement_consent"},
	}
}
