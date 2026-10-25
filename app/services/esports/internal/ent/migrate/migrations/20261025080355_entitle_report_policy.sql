-- Create "report_policy" table
CREATE TABLE "public"."report_policy" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
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
-- Set comment to column: "tenant_id" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "report_type" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."report_type" IS 'the family this row governs — the upsert key, one row per family';
-- Set comment to column: "policy" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."policy" IS 'who may enter: everyone, active members, the tenant''s own (awaiting the tenant plane), or only explicit grants';
-- Set comment to column: "updated_by" on table: "report_policy"
COMMENT ON COLUMN "public"."report_policy"."updated_by" IS 'admin identity that last set the policy (audit join)';
