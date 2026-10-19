package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"

	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
)

type File struct {
	ent.Schema
}

func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().MaxRuneLen(512).
			Comment("storage object key; server-generated prefix+xid, never client-chosen"),
		field.String("name").NotEmpty().MaxRuneLen(255).
			Comment("display name; decoupled from the stored key"),
		field.String("content_type").Default("").MaxRuneLen(128).
			Comment("declared at upload, backfilled from HEAD at confirm; storage wins"),
		field.Int64("size").Default(0).
			Comment("declared at create, replaced by the measured size at confirm"),
		field.String("source").Default("client_upload").MaxRuneLen(32).
			Comment("client_upload | server_generated"),
		field.String("status").Default("uploading").MaxRuneLen(16).
			Comment("uploading | processing | confirmed | failed; aborted uploads hard-delete the row"),
		field.String("upload_id").Default("").MaxRuneLen(128).
			Comment("S3 multipart session id; empty on the single-PUT path"),
		field.String("etag").Default("").MaxRuneLen(64).
			Comment("object ETag captured at confirm; integrity/dedup seam"),
		field.String("owner_id").Default("").MaxRuneLen(20).
			Comment("creating user; datascope self dimension"),
	}
}

func (File) Edges() []ent.Edge {
	return nil
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
		local_mixins.TenantMixin{},
		local_mixins.DatascopeMixin{UserField: "owner_id"},
		local_mixins.SoftDeleteMixin{},
	}
}

func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "created_at"),
		index.Fields("key").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}

func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "file"},
	}
}
