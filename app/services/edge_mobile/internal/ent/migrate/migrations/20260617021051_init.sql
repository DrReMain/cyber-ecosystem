-- Create "mobile_user" table
CREATE TABLE "public"."mobile_user" (
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
-- Create index "mobileuser_created_at" to table: "mobile_user"
CREATE INDEX "mobileuser_created_at" ON "public"."mobile_user" ("created_at");
-- Create index "mobileuser_phone" to table: "mobile_user"
CREATE UNIQUE INDEX "mobileuser_phone" ON "public"."mobile_user" ("phone") WHERE (deleted_at IS NULL);
-- Create index "mobileuser_sort" to table: "mobile_user"
CREATE UNIQUE INDEX "mobileuser_sort" ON "public"."mobile_user" ("sort") WHERE (deleted_at IS NULL);
-- Create index "mobileuser_updated_at" to table: "mobile_user"
CREATE INDEX "mobileuser_updated_at" ON "public"."mobile_user" ("updated_at");
-- Set comment to column: "sort" on table: "mobile_user"
COMMENT ON COLUMN "public"."mobile_user"."sort" IS 'fractional index for ordering';
-- Set comment to column: "phone" on table: "mobile_user"
COMMENT ON COLUMN "public"."mobile_user"."phone" IS 'login account';
-- Set comment to column: "status" on table: "mobile_user"
COMMENT ON COLUMN "public"."mobile_user"."status" IS 'enabled|disabled|restricted';
