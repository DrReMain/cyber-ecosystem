-- Set comment to column: "source" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."source" IS 'granting artifact id — the order number on membership rows, the code on report-type rows';
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
COMMENT ON COLUMN "public"."report_snapshot"."tier" IS 'the pinned tier badge, judged once at unlock time — the ruling the roadmap froze: an unlocked report never moves with the rubric';
-- Set comment to column: "composite_percentile" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."composite_percentile" IS 'the pinned composite (equal-weight mean, missing abilities never zero) — null when no measured ability carried a percentile at pin time';
-- Set comment to column: "abilities" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."abilities" IS 'the pinned radar board: one entry per measured ability (ability, game_code, score, percentile, confidence, sample_count, peer_average) — the exact shape the report face returns';
-- Set comment to column: "cohort_basis" on table: "report_snapshot"
COMMENT ON COLUMN "public"."report_snapshot"."cohort_basis" IS 'the cohort rubric this snapshot was computed under (same-age-same-sex exact, or band-approximate when only the band is known) — the audit stamp that makes drift detectable';
