-- Create "dept" table
CREATE TABLE "public"."dept" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "parent_id" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "dept_created_at" to table: "dept"
CREATE INDEX "dept_created_at" ON "public"."dept" ("created_at");
-- Create index "dept_tenant_id_parent_id_name" to table: "dept"
CREATE UNIQUE INDEX "dept_tenant_id_parent_id_name" ON "public"."dept" ("tenant_id", "parent_id", "name") WHERE (deleted_at IS NULL);
-- Create index "dept_updated_at" to table: "dept"
CREATE INDEX "dept_updated_at" ON "public"."dept" ("updated_at");
-- Create "item" table
CREATE TABLE "public"."item" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "sort" character varying NOT NULL COLLATE "C",
  "name" character varying NOT NULL,
  "description" character varying NOT NULL DEFAULT '',
  "status" character varying NOT NULL DEFAULT 'active',
  PRIMARY KEY ("id")
);
-- Create index "item_created_at" to table: "item"
CREATE INDEX "item_created_at" ON "public"."item" ("created_at");
-- Create index "item_name" to table: "item"
CREATE UNIQUE INDEX "item_name" ON "public"."item" ("name") WHERE (deleted_at IS NULL);
-- Create index "item_sort" to table: "item"
CREATE UNIQUE INDEX "item_sort" ON "public"."item" ("sort") WHERE (deleted_at IS NULL);
-- Create index "item_updated_at" to table: "item"
CREATE INDEX "item_updated_at" ON "public"."item" ("updated_at");
-- Set comment to column: "sort" on table: "item"
COMMENT ON COLUMN "public"."item"."sort" IS 'fractional index for ordering';
-- Set comment to column: "status" on table: "item"
COMMENT ON COLUMN "public"."item"."status" IS 'active/inactive';
-- Create "user" table
CREATE TABLE "public"."user" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" character varying NOT NULL,
  "email" character varying NOT NULL,
  "password_hash" character varying NOT NULL,
  "dept_id" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "user_created_at" to table: "user"
CREATE INDEX "user_created_at" ON "public"."user" ("created_at");
-- Create index "user_tenant_id_email" to table: "user"
CREATE UNIQUE INDEX "user_tenant_id_email" ON "public"."user" ("tenant_id", "email") WHERE (deleted_at IS NULL);
-- Create index "user_updated_at" to table: "user"
CREATE INDEX "user_updated_at" ON "public"."user" ("updated_at");
