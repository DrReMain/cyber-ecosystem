-- Create "users" table
CREATE TABLE "public"."users" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "sort" character varying NOT NULL COLLATE "C",
  "nickname" character varying NULL,
  "avatar" character varying NULL,
  "phone" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'enabled',
  "password_hash" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "user_created_at" to table: "users"
CREATE INDEX "user_created_at" ON "public"."users" ("created_at");
-- Create index "user_phone" to table: "users"
CREATE UNIQUE INDEX "user_phone" ON "public"."users" ("phone") WHERE (deleted_at IS NULL);
-- Create index "user_sort" to table: "users"
CREATE UNIQUE INDEX "user_sort" ON "public"."users" ("sort") WHERE (deleted_at IS NULL);
-- Create index "user_updated_at" to table: "users"
CREATE INDEX "user_updated_at" ON "public"."users" ("updated_at");
-- Set comment to column: "sort" on table: "users"
COMMENT ON COLUMN "public"."users"."sort" IS 'fractional index for ordering';
-- Set comment to column: "phone" on table: "users"
COMMENT ON COLUMN "public"."users"."phone" IS 'login account';
-- Set comment to column: "status" on table: "users"
COMMENT ON COLUMN "public"."users"."status" IS 'enabled|disabled|restricted';
