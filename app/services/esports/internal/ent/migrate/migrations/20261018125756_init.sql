-- Create "profile" table
CREATE TABLE "public"."profile" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "user_id" character varying NOT NULL,
  "nickname" character varying NULL,
  "birth_year" bigint NULL,
  "birth_month" bigint NULL,
  "avatar_url" character varying NULL,
  "guardian_consented_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "profile_created_at" to table: "profile"
CREATE INDEX "profile_created_at" ON "public"."profile" ("created_at");
-- Create index "profile_tenant_id_created_at" to table: "profile"
CREATE INDEX "profile_tenant_id_created_at" ON "public"."profile" ("tenant_id", "created_at");
-- Create index "profile_updated_at" to table: "profile"
CREATE INDEX "profile_updated_at" ON "public"."profile" ("updated_at");
-- Create index "profile_user_id" to table: "profile"
CREATE UNIQUE INDEX "profile_user_id" ON "public"."profile" ("user_id") WHERE (deleted_at IS NULL);
-- Set comment to column: "tenant_id" on table: "profile"
COMMENT ON COLUMN "public"."profile"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "profile"
COMMENT ON COLUMN "public"."profile"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "user_id" on table: "profile"
COMMENT ON COLUMN "public"."profile"."user_id" IS 'owning member identity (system user id); one live row per member';
-- Set comment to column: "nickname" on table: "profile"
COMMENT ON COLUMN "public"."profile"."nickname" IS 'display name; 2-12 chars once set (boundary-validated, not a DB concern)';
-- Set comment to column: "birth_year" on table: "profile"
COMMENT ON COLUMN "public"."profile"."birth_year" IS 'birth year; year+month is the collected granularity, a day is fabricated precision';
-- Set comment to column: "birth_month" on table: "profile"
COMMENT ON COLUMN "public"."profile"."birth_month" IS 'birth month 1-12';
-- Set comment to column: "avatar_url" on table: "profile"
COMMENT ON COLUMN "public"."profile"."avatar_url" IS 'avatar URL; wechat prefill hydrates identity-side, storage lives here';
-- Set comment to column: "guardian_consented_at" on table: "profile"
COMMENT ON COLUMN "public"."profile"."guardian_consented_at" IS 'under-13 hard-gate stamp, server-written only; null = not consented';
