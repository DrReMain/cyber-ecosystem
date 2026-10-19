-- Create "game" table
CREATE TABLE "public"."game" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "category" character varying NULL,
  "icon_url" character varying NULL,
  "description" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "game_code" to table: "game"
CREATE UNIQUE INDEX "game_code" ON "public"."game" ("code") WHERE (deleted_at IS NULL);
-- Create index "game_created_at" to table: "game"
CREATE INDEX "game_created_at" ON "public"."game" ("created_at");
-- Create index "game_tenant_id_created_at" to table: "game"
CREATE INDEX "game_tenant_id_created_at" ON "public"."game" ("tenant_id", "created_at");
-- Create index "game_updated_at" to table: "game"
CREATE INDEX "game_updated_at" ON "public"."game" ("updated_at");
-- Set comment to column: "tenant_id" on table: "game"
COMMENT ON COLUMN "public"."game"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "game"
COMMENT ON COLUMN "public"."game"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "code" on table: "game"
COMMENT ON COLUMN "public"."game"."code" IS 'immutable catalog code join key (e.g. lol); born once, referenced forever';
-- Set comment to column: "name" on table: "game"
COMMENT ON COLUMN "public"."game"."name" IS 'display title, legacy experience data';
-- Set comment to column: "category" on table: "game"
COMMENT ON COLUMN "public"."game"."category" IS 'vocabulary fps|moba|... — open word set, admin-managed; new categories classify in without a contract change';
-- Set comment to column: "icon_url" on table: "game"
COMMENT ON COLUMN "public"."game"."icon_url" IS 'icon URL; hydrated from file service at render';
-- Set comment to column: "description" on table: "game"
COMMENT ON COLUMN "public"."game"."description" IS 'short catalog blurb';
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
  "age_band" character varying NULL,
  "avatar_url" character varying NULL,
  "guardian_consented_at" timestamptz NULL,
  "games" jsonb NULL,
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
-- Set comment to column: "age_band" on table: "profile"
COMMENT ON COLUMN "public"."profile"."age_band" IS 'derived at write from birth year/month; 8-band word set under12|12-14|15-17|18-22|23-25|26-29|30-40|over40; null while birth unknown';
-- Set comment to column: "avatar_url" on table: "profile"
COMMENT ON COLUMN "public"."profile"."avatar_url" IS 'avatar URL; wechat prefill hydrates identity-side, storage lives here';
-- Set comment to column: "guardian_consented_at" on table: "profile"
COMMENT ON COLUMN "public"."profile"."guardian_consented_at" IS 'under-13 hard-gate stamp, server-written only; null = not consented';
-- Set comment to column: "games" on table: "profile"
COMMENT ON COLUMN "public"."profile"."games" IS 'replace-set preference entries (catalog ref or custom name, exactly one set per entry)';
