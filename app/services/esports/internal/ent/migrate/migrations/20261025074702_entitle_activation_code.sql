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
