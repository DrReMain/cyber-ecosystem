-- Drop index "user_email" from table: "user"
DROP INDEX "public"."user_email";
-- Modify "user" table
ALTER TABLE "public"."user" ADD COLUMN "tenant_id" character varying NOT NULL;
-- Create index "user_tenant_id_email" to table: "user"
CREATE UNIQUE INDEX "user_tenant_id_email" ON "public"."user" ("tenant_id", "email") WHERE (deleted_at IS NULL);
