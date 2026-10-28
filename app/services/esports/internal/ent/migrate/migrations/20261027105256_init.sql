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
COMMENT ON COLUMN "public"."ability_score"."score" IS 'the ability contribution this attempt scored — written by the scoring engine, never the client';
-- Set comment to column: "percentile" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."percentile" IS 'same-age percentile once the cohort math lands; null until then';
-- Create "activation_code" table
CREATE TABLE "public"."activation_code" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "report_type" character varying NOT NULL,
  "quantity" bigint NOT NULL DEFAULT 1,
  "used_count" bigint NOT NULL DEFAULT 0,
  "expires_at" timestamptz NOT NULL,
  "created_by" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "activationcode_code" to table: "activation_code"
CREATE UNIQUE INDEX "activationcode_code" ON "public"."activation_code" ("code");
-- Create index "activationcode_created_at" to table: "activation_code"
CREATE INDEX "activationcode_created_at" ON "public"."activation_code" ("created_at");
-- Create index "activationcode_report_type_created_at" to table: "activation_code"
CREATE INDEX "activationcode_report_type_created_at" ON "public"."activation_code" ("report_type", "created_at");
-- Create index "activationcode_updated_at" to table: "activation_code"
CREATE INDEX "activationcode_updated_at" ON "public"."activation_code" ("updated_at");
-- Set comment to column: "tenant_id" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "code" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."code" IS 'operator-handout code string (prefix + secure random); stored clear because support re-copies it';
-- Set comment to column: "report_type" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."report_type" IS 'the report family this code opens — the pricing ruling retired single-record codes';
-- Set comment to column: "quantity" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."quantity" IS 'redeems allowed; generation mints single-use codes, the field keeps the redeem guard general';
-- Set comment to column: "used_count" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."used_count" IS 'redeems taken; the conditional UPDATE increments under the quantity guard';
-- Set comment to column: "expires_at" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."expires_at" IS 'code validity deadline; a redeem after it refuses even with uses left';
-- Set comment to column: "created_by" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."created_by" IS 'admin identity that minted the batch (audit join)';
-- Create "entitlement" table
CREATE TABLE "public"."entitlement" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "scope_kind" character varying NOT NULL,
  "scope_ref" character varying NULL,
  "expires_at" timestamptz NULL,
  "source" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "entitlement_created_at" to table: "entitlement"
CREATE INDEX "entitlement_created_at" ON "public"."entitlement" ("created_at");
-- Create index "entitlement_tenant_id_user_id_scope_kind_scope_ref" to table: "entitlement"
CREATE UNIQUE INDEX "entitlement_tenant_id_user_id_scope_kind_scope_ref" ON "public"."entitlement" ("tenant_id", "user_id", "scope_kind", "scope_ref");
-- Create index "entitlement_updated_at" to table: "entitlement"
CREATE INDEX "entitlement_updated_at" ON "public"."entitlement" ("updated_at");
-- Set comment to column: "tenant_id" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "user_id" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."user_id" IS 'holding member identity (system user id), logical reference — same key law as profile';
-- Set comment to column: "scope_kind" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_kind" IS 'grant shape: trial points at one record permanently, membership opens a window, report_type opens a report family via a type code';
-- Set comment to column: "scope_ref" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_ref" IS 'trial: the unlocked test record id; report_type: the family name (dedup guard); membership: null — renewal rows stack instead of dedup';
-- Set comment to column: "expires_at" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."expires_at" IS 'membership window end (paid_at + period days); null on trial rows — permanence is the absence of a deadline';
-- Set comment to column: "source" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."source" IS 'granting artifact id — the order number on membership rows, the code on report-type rows';
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
-- Create "pay_order" table
CREATE TABLE "public"."pay_order" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "order_no" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "sku_code" character varying NOT NULL,
  "scope_kind" character varying NOT NULL,
  "scope_ref" character varying NULL,
  "provider" character varying NOT NULL,
  "channel" character varying NOT NULL,
  "amount_fen" bigint NOT NULL,
  "state" character varying NOT NULL DEFAULT 'pending',
  "credential_kind" character varying NOT NULL,
  "credential_value" character varying NOT NULL,
  "transaction_id" character varying NULL,
  "expires_at" timestamptz NOT NULL,
  "paid_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "payorder_created_at" to table: "pay_order"
CREATE INDEX "payorder_created_at" ON "public"."pay_order" ("created_at");
-- Create index "payorder_order_no" to table: "pay_order"
CREATE UNIQUE INDEX "payorder_order_no" ON "public"."pay_order" ("order_no");
-- Create index "payorder_state_expires_at" to table: "pay_order"
CREATE INDEX "payorder_state_expires_at" ON "public"."pay_order" ("state", "expires_at");
-- Create index "payorder_tenant_id_user_id_created_at" to table: "pay_order"
CREATE INDEX "payorder_tenant_id_user_id_created_at" ON "public"."pay_order" ("tenant_id", "user_id", "created_at");
-- Create index "payorder_tenant_id_user_id_scope_kind_scope_ref" to table: "pay_order"
CREATE UNIQUE INDEX "payorder_tenant_id_user_id_scope_kind_scope_ref" ON "public"."pay_order" ("tenant_id", "user_id", "scope_kind", "scope_ref") WHERE ((state)::text = 'pending'::text);
-- Create index "payorder_updated_at" to table: "pay_order"
CREATE INDEX "payorder_updated_at" ON "public"."pay_order" ("updated_at");
-- Set comment to column: "tenant_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "order_no" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."order_no" IS 'business order number (''P''+ULID) — the channel idempotency key (out_trade_no) and the audit join key into bills';
-- Set comment to column: "user_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."user_id" IS 'buying member identity (system user id), logical reference — same key law as profile';
-- Set comment to column: "sku_code" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."sku_code" IS 'sku catalog code at purchase time (snapshot; later sku edits never rewrite history)';
-- Set comment to column: "scope_kind" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."scope_kind" IS 'grant shape: trial unlocks one test record (scope_ref), period opens the membership window';
-- Set comment to column: "scope_ref" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."scope_ref" IS 'trial: the test record id being unlocked; period: null — the window shape comes from the sku';
-- Set comment to column: "provider" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."provider" IS 'payment provider name (wechat, mock) — the key the platform''s container map routes query/close by';
-- Set comment to column: "channel" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."channel" IS 'access-context channel value (provider vocabulary, e.g. wechat_jsapi); fixes how the credential is consumed';
-- Set comment to column: "amount_fen" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."amount_fen" IS 'agreed amount in fen, frozen at creation; settlement verifies the channel-reported amount against it';
-- Set comment to column: "state" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."state" IS 'channel-state projection; paid is terminal — refunds are derived facts, not states';
-- Set comment to column: "credential_kind" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."credential_kind" IS 'how the page consumes the credential (invoke|jump_url|qr_code), snapshotted with the value';
-- Set comment to column: "credential_value" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."credential_value" IS 'channel credential at creation; same-scope reuse replays it with zero channel round trips';
-- Set comment to column: "transaction_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."transaction_id" IS 'channel transaction id, backfilled at settlement — the bill-reconciliation join key';
-- Set comment to column: "expires_at" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."expires_at" IS 'payment deadline carried from creation (channel default two hours); expiry closes pending orders';
-- Set comment to column: "paid_at" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."paid_at" IS 'channel-side payment time; the entitlement window starts here for membership';
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
  "sex" character varying NULL,
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
-- Set comment to column: "sex" on table: "profile"
COMMENT ON COLUMN "public"."profile"."sex" IS 'three states by the sex ruling — undetermined is the honest not-said; the report cohort reads it (same-age-same-sex, undetermined benches against the same-age pool)';
-- Set comment to column: "age_band" on table: "profile"
COMMENT ON COLUMN "public"."profile"."age_band" IS 'derived at write from birth year/month; 8-band word set under12|12-14|15-17|18-22|23-25|26-29|30-40|over40; null while birth unknown';
-- Set comment to column: "avatar_url" on table: "profile"
COMMENT ON COLUMN "public"."profile"."avatar_url" IS 'avatar URL; wechat prefill hydrates identity-side, storage lives here';
-- Set comment to column: "guardian_consented_at" on table: "profile"
COMMENT ON COLUMN "public"."profile"."guardian_consented_at" IS 'under-13 hard-gate stamp, server-written only; null = not consented';
-- Set comment to column: "games" on table: "profile"
COMMENT ON COLUMN "public"."profile"."games" IS 'replace-set preference entries (catalog ref or custom name, exactly one set per entry)';
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
-- Create "report_policy" table
CREATE TABLE "public"."report_policy" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "report_type" character varying NOT NULL,
  "policy" character varying NOT NULL,
  "updated_by" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "reportpolicy_created_at" to table: "report_policy"
CREATE INDEX "reportpolicy_created_at" ON "public"."report_policy" ("created_at");
-- Create index "reportpolicy_report_type" to table: "report_policy"
CREATE UNIQUE INDEX "reportpolicy_report_type" ON "public"."report_policy" ("report_type");
-- Create index "reportpolicy_updated_at" to table: "report_policy"
CREATE INDEX "reportpolicy_updated_at" ON "public"."report_policy" ("updated_at");
-- Set comment to column: "report_type" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."report_type" IS 'the family this row governs — the upsert key, one row per family';
-- Set comment to column: "policy" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."policy" IS 'who may enter: everyone, active members, the tenant''s own (awaiting the tenant plane), or only explicit grants';
-- Set comment to column: "updated_by" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."updated_by" IS 'admin identity that last set the policy (audit join)';
-- Create "report_snapshot" table
CREATE TABLE "public"."report_snapshot" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "record_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "tier" character varying NOT NULL,
  "composite_percentile" double precision NULL,
  "abilities" jsonb NOT NULL,
  "cohort_basis" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "reportsnapshot_created_at" to table: "report_snapshot"
CREATE INDEX "reportsnapshot_created_at" ON "public"."report_snapshot" ("created_at");
-- Create index "reportsnapshot_tenant_id_record_id" to table: "report_snapshot"
CREATE UNIQUE INDEX "reportsnapshot_tenant_id_record_id" ON "public"."report_snapshot" ("tenant_id", "record_id") WHERE (deleted_at IS NULL);
-- Create index "reportsnapshot_tenant_id_user_id_created_at" to table: "report_snapshot"
CREATE INDEX "reportsnapshot_tenant_id_user_id_created_at" ON "public"."report_snapshot" ("tenant_id", "user_id", "created_at");
-- Create index "reportsnapshot_updated_at" to table: "report_snapshot"
CREATE INDEX "reportsnapshot_updated_at" ON "public"."report_snapshot" ("updated_at");
-- Set comment to column: "tenant_id" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "record_id" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."record_id" IS 'owning attempt (join into test_record.id) — plain id, not an edge: the loose join keeps soft-delete cascade and future aggregation out of this table''s way, the ability_score precedent';
-- Set comment to column: "user_id" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."user_id" IS 'denormalized owner, the ability_score precedent — cohort reads never join through test_record';
-- Set comment to column: "tier" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."tier" IS 'the pinned tier badge, judged once at unlock time — the ruling the roadmap froze: an unlocked report never moves with the rubric; unjudged pins the honest no-verdict (every cohort too thin for a percentile), never a made-up letter';
-- Set comment to column: "composite_percentile" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."composite_percentile" IS 'the pinned composite (equal-weight mean, missing abilities never zero) — null when no measured ability carried a percentile at pin time';
-- Set comment to column: "abilities" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."abilities" IS 'the pinned radar board: one entry per measured ability (ability, game_code, score, percentile, confidence, sample_count, peer_average) — the exact shape the report face returns';
-- Set comment to column: "cohort_basis" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."cohort_basis" IS 'the cohort rubric this snapshot was computed under (same-age-same-sex exact, or band-approximate when only the band is known) — the audit stamp that makes drift detectable';
-- Create "sku" table
CREATE TABLE "public"."sku" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "kind" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "price_fen" bigint NOT NULL,
  "list_price_fen" bigint NOT NULL,
  "period_days" bigint NULL,
  "on_shelf" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "sku_code" to table: "sku"
CREATE UNIQUE INDEX "sku_code" ON "public"."sku" ("code") WHERE (deleted_at IS NULL);
-- Create index "sku_created_at" to table: "sku"
CREATE INDEX "sku_created_at" ON "public"."sku" ("created_at");
-- Create index "sku_updated_at" to table: "sku"
CREATE INDEX "sku_updated_at" ON "public"."sku" ("updated_at");
-- Set comment to column: "tenant_id" on table: "sku"
COMMENT ON COLUMN "public"."sku"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "sku"
COMMENT ON COLUMN "public"."sku"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "code" on table: "sku"
COMMENT ON COLUMN "public"."sku"."code" IS 'catalog code join key (e.g. member_month); orders snapshot it, the row ULID never crosses a boundary';
-- Set comment to column: "kind" on table: "sku"
COMMENT ON COLUMN "public"."sku"."kind" IS 'what the sku sells: one trial''s report permanently vs a membership window';
-- Set comment to column: "name" on table: "sku"
COMMENT ON COLUMN "public"."sku"."name" IS 'display title (pay sheet line); also the channel order description source';
-- Set comment to column: "description" on table: "sku"
COMMENT ON COLUMN "public"."sku"."description" IS 'pay-sheet blurb under the title';
-- Set comment to column: "price_fen" on table: "sku"
COMMENT ON COLUMN "public"."sku"."price_fen" IS 'selling price in fen (1/100 yuan) — money stays integer fen end to end';
-- Set comment to column: "list_price_fen" on table: "sku"
COMMENT ON COLUMN "public"."sku"."list_price_fen" IS 'list (pre-discount) price; equals price_fen while no discount is shown, carried so discounting stays a value change';
-- Set comment to column: "period_days" on table: "sku"
COMMENT ON COLUMN "public"."sku"."period_days" IS 'membership window in days (31: the 32nd day expires); null on trial skus';
-- Set comment to column: "on_shelf" on table: "sku"
COMMENT ON COLUMN "public"."sku"."on_shelf" IS 'member-plane visibility switch; off-shelf rows stay for order-history joins';
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
COMMENT ON COLUMN "public"."test"."name" IS 'display title; seeds carry the placeholder code until the metric sheet calibrates';
-- Set comment to column: "category" on table: "test"
COMMENT ON COLUMN "public"."test"."category" IS 'vocabulary fps|moba|universal — universal feeds every genre pool; open word set, admin-managed, classification is content not contract';
-- Set comment to column: "ability" on table: "test"
COMMENT ON COLUMN "public"."test"."ability" IS 'primary ability, six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency; attribution is metric-sheet data pending calibration';
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
-- Create index "testrecord_session_id_test_code_attempt_no" to table: "test_record"
CREATE UNIQUE INDEX "testrecord_session_id_test_code_attempt_no" ON "public"."test_record" ("session_id", "test_code", "attempt_no") WHERE (session_id IS NOT NULL);
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
COMMENT ON COLUMN "public"."test_record"."attempt_no" IS 'attempt ordinal inside the match session; the partial unique (session_id, test_code, attempt_no) makes the count-then-insert claim atomic under concurrent submits';
-- Set comment to column: "session_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."session_id" IS 'match session reference — null on practice attempts, set on session attempts';
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
