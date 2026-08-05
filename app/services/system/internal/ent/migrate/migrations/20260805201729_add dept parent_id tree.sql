-- Drop index "dept_tenant_id_name" from table: "dept"
DROP INDEX "public"."dept_tenant_id_name";
-- Modify "dept" table
ALTER TABLE "public"."dept" ADD COLUMN "parent_id" character varying NULL;
-- Create index "dept_tenant_id_parent_id_name" to table: "dept"
CREATE UNIQUE INDEX "dept_tenant_id_parent_id_name" ON "public"."dept" ("tenant_id", "parent_id", "name") WHERE (deleted_at IS NULL);
