-- Create "authz_stamps" table
CREATE TABLE "public"."authz_stamps" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "version" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "authzstamp_created_at" to table: "authz_stamps"
CREATE INDEX "authzstamp_created_at" ON "public"."authz_stamps" ("created_at");
-- Create index "authzstamp_updated_at" to table: "authz_stamps"
CREATE INDEX "authzstamp_updated_at" ON "public"."authz_stamps" ("updated_at");
