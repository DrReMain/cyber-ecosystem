package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// AgreementVersion is one published version of one legal document. The
// compliance ledger: exactly one current version per document is the
// invariant every consent records against, maintained inside a single
// transaction by the owning module. Global by the same ruling as
// report_policy — one operator plane governs every tenant, so no tenant
// column and no tenant filtering. No soft delete: a version is either
// live or hard-deleted, and consent rows survive the deletion by
// carrying the version string they landed on.
type AgreementVersion struct {
	ent.Schema
}

func (AgreementVersion) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("doc").Values("user", "privacy").Immutable().
			Comment("the legal document this version publishes"),
		field.String("no").NotEmpty().MaxLen(20).Immutable().
			Comment("operator-quoted version string (v1.0.0…); unique per document, format-free"),
		field.String("title").NotEmpty().MaxRuneLen(100).Immutable().
			Comment("display title derived from the document type at create; historical rows keep what they were published under"),
		field.Text("body").NotEmpty().Immutable().
			Comment("the published legal text — the compliance archive an operator quotes in a dispute"),
		field.Time("effective_at").Immutable().
			Comment("the date this version started binding, shown on the legal page's version line"),
		field.Bool("is_current").Default(false).
			Comment("true = the version every new consent lands on; exactly one per document, flipped in one transaction"),
	}
}

func (AgreementVersion) Edges() []ent.Edge {
	return nil
}

func (AgreementVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (AgreementVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("doc", "no").Unique(),
	}
}

func (AgreementVersion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "agreement_version"},
	}
}
