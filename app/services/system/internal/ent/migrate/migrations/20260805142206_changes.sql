-- Create "dept" table
CREATE TABLE "public"."dept" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "dept_created_at" to table: "dept"
CREATE INDEX "dept_created_at" ON "public"."dept" ("created_at");
-- Create index "dept_tenant_id_name" to table: "dept"
CREATE UNIQUE INDEX "dept_tenant_id_name" ON "public"."dept" ("tenant_id", "name") WHERE (deleted_at IS NULL);
-- Create index "dept_updated_at" to table: "dept"
CREATE INDEX "dept_updated_at" ON "public"."dept" ("updated_at");
