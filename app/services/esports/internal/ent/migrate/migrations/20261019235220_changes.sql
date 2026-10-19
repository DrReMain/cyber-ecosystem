-- Create "ability_score" table
CREATE TABLE "public"."ability_score" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "record_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "ability" character varying NOT NULL,
  "score" double precision NOT NULL,
  "percentile" double precision NULL,
  PRIMARY KEY ("id")
);
-- Create index "abilityscore_created_at" to table: "ability_score"
CREATE INDEX "abilityscore_created_at" ON "public"."ability_score" ("created_at");
-- Create index "abilityscore_record_id_ability" to table: "ability_score"
CREATE UNIQUE INDEX "abilityscore_record_id_ability" ON "public"."ability_score" ("record_id", "ability") WHERE (deleted_at IS NULL);
-- Create index "abilityscore_tenant_id_created_at" to table: "ability_score"
CREATE INDEX "abilityscore_tenant_id_created_at" ON "public"."ability_score" ("tenant_id", "created_at");
-- Create index "abilityscore_updated_at" to table: "ability_score"
CREATE INDEX "abilityscore_updated_at" ON "public"."ability_score" ("updated_at");
-- Create index "abilityscore_user_id_ability_created_at" to table: "ability_score"
CREATE INDEX "abilityscore_user_id_ability_created_at" ON "public"."ability_score" ("user_id", "ability", "created_at");
-- Set comment to column: "tenant_id" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "record_id" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."record_id" IS 'owning attempt (join into test_record.id) — plain id, not an edge: soft-delete cascade and future cross-service aggregation both favor the loose join';
-- Set comment to column: "user_id" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."user_id" IS 'denormalized owner so cross-attempt aggregation (historical best, same-age percentile) never joins through test_record';
-- Set comment to column: "ability" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."ability" IS 'six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency — one row per ability an attempt feeds';
-- Set comment to column: "score" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."score" IS 'the ability contribution this attempt scored — written by the E3 scoring engine, never the client';
-- Set comment to column: "percentile" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."percentile" IS 'same-age percentile once the E5 cohort math lands; null until then';
-- Create "question_bank" table
CREATE TABLE "public"."question_bank" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "test_code" character varying NOT NULL,
  "scenario_code" character varying NOT NULL,
  "payload" jsonb NULL,
  "answer" jsonb NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "questionbank_created_at" to table: "question_bank"
CREATE INDEX "questionbank_created_at" ON "public"."question_bank" ("created_at");
-- Create index "questionbank_tenant_id_created_at" to table: "question_bank"
CREATE INDEX "questionbank_tenant_id_created_at" ON "public"."question_bank" ("tenant_id", "created_at");
-- Create index "questionbank_test_code_scenario_code" to table: "question_bank"
CREATE UNIQUE INDEX "questionbank_test_code_scenario_code" ON "public"."question_bank" ("test_code", "scenario_code") WHERE (deleted_at IS NULL);
-- Create index "questionbank_updated_at" to table: "question_bank"
CREATE INDEX "questionbank_updated_at" ON "public"."question_bank" ("updated_at");
-- Set comment to column: "tenant_id" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "test_code" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."test_code" IS 'owning test (join into test.code) — the scenario answer key lives with the test it grades';
-- Set comment to column: "scenario_code" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."scenario_code" IS 'scenario join key — the identifier attempts echo back in their decision facts';
-- Set comment to column: "payload" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."payload" IS 'question face — scenario content for rendering; shape is test-owned content, not contract';
-- Set comment to column: "answer" on table: "question_bank"
COMMENT ON COLUMN "public"."question_bank"."answer" IS 'answer key — correctness of submitted choices is judged against this server-side, never asserted by the client';
-- Create "test" table
CREATE TABLE "public"."test" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "category" character varying NULL,
  "ability" character varying NULL,
  "description" character varying NULL,
  "unit_score_label" character varying NULL,
  "duration_seconds" bigint NOT NULL DEFAULT 30,
  "introduction" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "test_code" to table: "test"
CREATE UNIQUE INDEX "test_code" ON "public"."test" ("code") WHERE (deleted_at IS NULL);
-- Create index "test_created_at" to table: "test"
CREATE INDEX "test_created_at" ON "public"."test" ("created_at");
-- Create index "test_tenant_id_created_at" to table: "test"
CREATE INDEX "test_tenant_id_created_at" ON "public"."test" ("tenant_id", "created_at");
-- Create index "test_updated_at" to table: "test"
CREATE INDEX "test_updated_at" ON "public"."test" ("updated_at");
-- Set comment to column: "tenant_id" on table: "test"
COMMENT ON COLUMN "public"."test"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "test"
COMMENT ON COLUMN "public"."test"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "code" on table: "test"
COMMENT ON COLUMN "public"."test"."code" IS 'immutable catalog code join key (e.g. red_dot); engines, seeds, and cross-domain references key off the code, same law as game codes';
-- Set comment to column: "name" on table: "test"
COMMENT ON COLUMN "public"."test"."name" IS 'display title; seeds carry the placeholder code until the E3 metric sheet calibrates';
-- Set comment to column: "category" on table: "test"
COMMENT ON COLUMN "public"."test"."category" IS 'vocabulary fps|moba|universal — universal feeds every genre pool; open word set, admin-managed, classification is content not contract';
-- Set comment to column: "ability" on table: "test"
COMMENT ON COLUMN "public"."test"."ability" IS 'primary ability, six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency; attribution is metric-sheet data, calibrated with E3';
-- Set comment to column: "description" on table: "test"
COMMENT ON COLUMN "public"."test"."description" IS 'short catalog blurb';
-- Set comment to column: "unit_score_label" on table: "test"
COMMENT ON COLUMN "public"."test"."unit_score_label" IS 'display form of the unit score (rendering copy only; the numeric itself never crosses the contract)';
-- Set comment to column: "duration_seconds" on table: "test"
COMMENT ON COLUMN "public"."test"."duration_seconds" IS 'canonical test length — the single source TestConfig mirrors so engine boot is one call; 30 is the placeholder until the metric sheet calibrates';
-- Set comment to column: "introduction" on table: "test"
COMMENT ON COLUMN "public"."test"."introduction" IS 'long-form how-to copy, per test; copy changes are content, never contract';
-- Create "test_config" table
CREATE TABLE "public"."test_config" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "test_code" character varying NOT NULL,
  "duration_seconds" bigint NOT NULL DEFAULT 30,
  "background_invalid_seconds" bigint NOT NULL DEFAULT 10,
  "engine_params" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create index "testconfig_created_at" to table: "test_config"
CREATE INDEX "testconfig_created_at" ON "public"."test_config" ("created_at");
-- Create index "testconfig_tenant_id_created_at" to table: "test_config"
CREATE INDEX "testconfig_tenant_id_created_at" ON "public"."test_config" ("tenant_id", "created_at");
-- Create index "testconfig_test_code" to table: "test_config"
CREATE UNIQUE INDEX "testconfig_test_code" ON "public"."test_config" ("test_code") WHERE (deleted_at IS NULL);
-- Create index "testconfig_updated_at" to table: "test_config"
CREATE INDEX "testconfig_updated_at" ON "public"."test_config" ("updated_at");
-- Set comment to column: "tenant_id" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "test_code" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."test_code" IS 'join key into test.code — same aggregate, one row per test, joined plain by code (the contract''s join-key law), not an edge: the pair is born together but read and tuned on different cadences';
-- Set comment to column: "duration_seconds" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."duration_seconds" IS 'mirror of test.duration_seconds — the engine boot bag is self-sufficient behind one config call';
-- Set comment to column: "background_invalid_seconds" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."background_invalid_seconds" IS 'client-side mechanics threshold: backgrounding longer than this invalidates the run; the engine only marks spans, validity is judged at submit';
-- Set comment to column: "engine_params" on table: "test_config"
COMMENT ON COLUMN "public"."test_config"."engine_params" IS 'engine tuning bag — spawn intervals, speeds, sizes; keys are content owned by the engine registry (metric spec sheet), the KV shape keeps tuning ops-only';
-- Create "test_record" table
CREATE TABLE "public"."test_record" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "test_code" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "attempt_no" bigint NOT NULL,
  "session_id" character varying NULL,
  "started_at" timestamptz NOT NULL,
  "ended_at" timestamptz NOT NULL,
  "backgrounds" jsonb NULL,
  "raw_detail" jsonb NOT NULL,
  "raw_score" double precision NULL,
  "authoritative_score" double precision NULL,
  "grade" character varying NULL,
  "valid" boolean NOT NULL DEFAULT true,
  "invalid_reason" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "testrecord_created_at" to table: "test_record"
CREATE INDEX "testrecord_created_at" ON "public"."test_record" ("created_at");
-- Create index "testrecord_tenant_id_created_at" to table: "test_record"
CREATE INDEX "testrecord_tenant_id_created_at" ON "public"."test_record" ("tenant_id", "created_at");
-- Create index "testrecord_test_code_created_at" to table: "test_record"
CREATE INDEX "testrecord_test_code_created_at" ON "public"."test_record" ("test_code", "created_at");
-- Create index "testrecord_updated_at" to table: "test_record"
CREATE INDEX "testrecord_updated_at" ON "public"."test_record" ("updated_at");
-- Create index "testrecord_user_id_test_code_created_at" to table: "test_record"
CREATE INDEX "testrecord_user_id_test_code_created_at" ON "public"."test_record" ("user_id", "test_code", "created_at");
-- Set comment to column: "tenant_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "test_code" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."test_code" IS 'the attempted test (join into test.code)';
-- Set comment to column: "user_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."user_id" IS 'owning member identity (system user id), logical reference — same key law as profile';
-- Set comment to column: "attempt_no" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."attempt_no" IS 'attempt ordinal inside the match session; uniqueness is the E4 session law, not a DB constraint yet';
-- Set comment to column: "session_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."session_id" IS 'match session reference — null until the E4 session aggregate lands, then backfilled';
-- Set comment to column: "started_at" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."started_at" IS 'attempt start (wall clock, server-stamped from the submitted anchor)';
-- Set comment to column: "ended_at" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."ended_at" IS 'attempt end';
-- Set comment to column: "backgrounds" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."backgrounds" IS 'engine-marked backgrounding spans; validity is judged against test_config.background_invalid_seconds';
-- Set comment to column: "raw_detail" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."raw_detail" IS 'the oneof fact stream verbatim (e.g. {''red_dot'': {''events'': [...]}}) — facts only, every aggregate is recomputed from it; the detail kind is the map''s single key';
-- Set comment to column: "raw_score" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."raw_score" IS 'linear uncapped track — null on invalid attempts (nothing scores)';
-- Set comment to column: "authoritative_score" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."authoritative_score" IS 'percentile-tiered authoritative track — null on invalid attempts';
-- Set comment to column: "grade" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."grade" IS 'percentile tier — top 5 / 15 / 40 percent / rest; enum-bounded at the ent layer so drift fails the write, not the report';
-- Set comment to column: "valid" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."valid" IS 'server-side validity verdict — background over budget or incomplete detail invalidates the attempt';
-- Set comment to column: "invalid_reason" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."invalid_reason" IS 'ops vocabulary code when valid=false (user copy resolves client-side)';
