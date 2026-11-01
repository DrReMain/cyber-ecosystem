-- Create "compliance_config" table
CREATE TABLE "public"."compliance_config" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "slot" character varying NOT NULL,
  "birth_collect" boolean NOT NULL DEFAULT true,
  "pay_age_line" bigint NOT NULL DEFAULT 18,
  "minor_prompt" character varying NOT NULL DEFAULT '',
  "missing_birth_prompt" character varying NOT NULL DEFAULT '',
  "updated_by" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "complianceconfig_created_at" to table: "compliance_config"
CREATE INDEX "complianceconfig_created_at" ON "public"."compliance_config" ("created_at");
-- Create index "complianceconfig_slot" to table: "compliance_config"
CREATE UNIQUE INDEX "complianceconfig_slot" ON "public"."compliance_config" ("slot");
-- Create index "complianceconfig_updated_at" to table: "compliance_config"
CREATE INDEX "complianceconfig_updated_at" ON "public"."compliance_config" ("updated_at");
-- Set comment to column: "slot" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."slot" IS 'the upsert pin — this table''s one and only row';
-- Set comment to column: "birth_collect" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."birth_collect" IS 'collect the member''s birth date at registration; off = age band only (the pay wall then guides the backfill)';
-- Set comment to column: "pay_age_line" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."pay_age_line" IS 'members younger than this cannot pay; age reckons on the maximum boundary — born Jan 1 means one year older only on Jan 2';
-- Set comment to column: "minor_prompt" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."minor_prompt" IS 'pay-wall copy for an under-age member; empty = each face renders its own localized default';
-- Set comment to column: "missing_birth_prompt" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."missing_birth_prompt" IS 'pay-wall copy when the birth date is missing; empty = localized default';
-- Set comment to column: "updated_by" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."updated_by" IS 'admin identity that last set any knob (audit join)';
