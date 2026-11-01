-- Modify "activation_code" table
ALTER TABLE "public"."activation_code" ADD COLUMN "batch_id" character varying NULL;
-- Create index "activationcode_batch_id" to table: "activation_code"
CREATE INDEX "activationcode_batch_id" ON "public"."activation_code" ("batch_id");
-- Set comment to column: "batch_id" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."batch_id" IS 'the minting batch; null only on rows predating the batch model';
-- Create "activation_batch" table
CREATE TABLE "public"."activation_batch" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "no" character varying NOT NULL,
  "report_type" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "disabled_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "activationbatch_created_at" to table: "activation_batch"
CREATE INDEX "activationbatch_created_at" ON "public"."activation_batch" ("created_at");
-- Create index "activationbatch_no" to table: "activation_batch"
CREATE UNIQUE INDEX "activationbatch_no" ON "public"."activation_batch" ("no");
-- Create index "activationbatch_report_type_created_at" to table: "activation_batch"
CREATE INDEX "activationbatch_report_type_created_at" ON "public"."activation_batch" ("report_type", "created_at");
-- Create index "activationbatch_updated_at" to table: "activation_batch"
CREATE INDEX "activationbatch_updated_at" ON "public"."activation_batch" ("updated_at");
-- Set comment to column: "tenant_id" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "no" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."no" IS 'human-shape batch number (B-YYMM-XXXX) — the string support quotes';
-- Set comment to column: "report_type" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."report_type" IS 'the report family every code in the batch opens';
-- Set comment to column: "expires_at" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."expires_at" IS 'code validity deadline shared by the whole batch';
-- Set comment to column: "disabled_at" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."disabled_at" IS 'null = live batch; set = the moment the operator disabled it — unredeemed codes void on the spot, landed grants stay landed';
