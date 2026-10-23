-- Create "match_config" table
CREATE TABLE "public"."match_config" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "genre" character varying NOT NULL,
  "ability" character varying NOT NULL,
  "render_order" bigint NOT NULL DEFAULT 1,
  PRIMARY KEY ("id")
);
-- Create index "matchconfig_created_at" to table: "match_config"
CREATE INDEX "matchconfig_created_at" ON "public"."match_config" ("created_at");
-- Create index "matchconfig_genre_ability" to table: "match_config"
CREATE UNIQUE INDEX "matchconfig_genre_ability" ON "public"."match_config" ("genre", "ability") WHERE (deleted_at IS NULL);
-- Create index "matchconfig_genre_render_order" to table: "match_config"
CREATE INDEX "matchconfig_genre_render_order" ON "public"."match_config" ("genre", "render_order");
-- Create index "matchconfig_tenant_id_created_at" to table: "match_config"
CREATE INDEX "matchconfig_tenant_id_created_at" ON "public"."match_config" ("tenant_id", "created_at");
-- Create index "matchconfig_updated_at" to table: "match_config"
CREATE INDEX "matchconfig_updated_at" ON "public"."match_config" ("updated_at");
-- Set comment to column: "tenant_id" on table: "match_config"
COMMENT ON COLUMN "public"."match_config"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "match_config"
COMMENT ON COLUMN "public"."match_config"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "genre" on table: "match_config"
COMMENT ON COLUMN "public"."match_config"."genre" IS 'type-axis vocabulary fps|moba|... — the scenario axis (channel batches) joins the same column later; open word set, admin-managed, same law as test.category';
-- Set comment to column: "ability" on table: "match_config"
COMMENT ON COLUMN "public"."match_config"."ability" IS 'six-word vocabulary reactionSpeed|aimingAccuracy|... — the ability this slot draws a test for';
-- Set comment to column: "render_order" on table: "match_config"
COMMENT ON COLUMN "public"."match_config"."render_order" IS 'orchestration position inside the genre; steps emit in this order, so retuning reorders lineups without a contract change';
-- Create "match_session" table
CREATE TABLE "public"."match_session" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "user_id" character varying NOT NULL,
  "genre" character varying NOT NULL,
  "difficulty" character varying NOT NULL DEFAULT 'experience',
  "seed" bigint NOT NULL DEFAULT 0,
  "steps" jsonb NOT NULL,
  "retest_budget" bigint NOT NULL DEFAULT 3,
  "status" character varying NOT NULL DEFAULT 'active',
  PRIMARY KEY ("id")
);
-- Create index "matchsession_created_at" to table: "match_session"
CREATE INDEX "matchsession_created_at" ON "public"."match_session" ("created_at");
-- Create index "matchsession_tenant_id_created_at" to table: "match_session"
CREATE INDEX "matchsession_tenant_id_created_at" ON "public"."match_session" ("tenant_id", "created_at");
-- Create index "matchsession_updated_at" to table: "match_session"
CREATE INDEX "matchsession_updated_at" ON "public"."match_session" ("updated_at");
-- Create index "matchsession_user_id" to table: "match_session"
CREATE UNIQUE INDEX "matchsession_user_id" ON "public"."match_session" ("user_id") WHERE (((status)::text = 'active'::text) AND (deleted_at IS NULL));
-- Create index "matchsession_user_id_created_at" to table: "match_session"
CREATE INDEX "matchsession_user_id_created_at" ON "public"."match_session" ("user_id", "created_at");
-- Set comment to column: "tenant_id" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "user_id" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."user_id" IS 'owning member identity (system user id), logical reference — same key law as profile and test_record';
-- Set comment to column: "genre" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."genre" IS 'the session''s type-axis value; frozen at creation — changing type is a new session, never a mutation';
-- Set comment to column: "difficulty" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."difficulty" IS 'mode axis: every attempt in the session scores on this mode at the submit seam; every session is born experience — standard is reserved for a purchasable entry';
-- Set comment to column: "seed" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."seed" IS 'the draw''s input entropy (FNV over user|day|genre|salt), stored for audit reproduction only — behavior always replays the locked steps';
-- Set comment to column: "steps" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."steps" IS 'the locked arrangement in match_config render_order; frozen at creation, resumes replay it verbatim';
-- Set comment to column: "retest_budget" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."retest_budget" IS 'attempts allowed per step inside this session — first-ever session 3, later sessions 1; frozen at birth, each stored attempt spends one valid or not';
-- Set comment to column: "status" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."status" IS 'active resumes inside the seven-day window; switched = explicit re-draw closed it, expired = window passed (closed lazily on read)';
