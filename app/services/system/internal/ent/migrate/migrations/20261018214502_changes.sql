-- Modify "user" table
-- The identity key is renamed, not reborn: every existing account IS its
-- email string, so logins keep working through the rename.
ALTER TABLE "public"."user" RENAME COLUMN "email" TO "account";
-- Create index "user_account" to table: "user"
DROP INDEX IF EXISTS "public"."user_email";
CREATE UNIQUE INDEX "user_account" ON "public"."user" ("account") WHERE (deleted_at IS NULL);
-- Set comment to column: "account" on table: "user"
COMMENT ON COLUMN "public"."user"."account" IS 'global identity handle; email and phone are bindable credentials, never the identity key';
