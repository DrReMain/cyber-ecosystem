-- Create "user" table
CREATE TABLE "public"."user" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "email" character varying NOT NULL,
  "password_hash" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "user_created_at" to table: "user"
CREATE INDEX "user_created_at" ON "public"."user" ("created_at");
-- Create index "user_email" to table: "user"
CREATE UNIQUE INDEX "user_email" ON "public"."user" ("email") WHERE (deleted_at IS NULL);
-- Create index "user_updated_at" to table: "user"
CREATE INDEX "user_updated_at" ON "public"."user" ("updated_at");
