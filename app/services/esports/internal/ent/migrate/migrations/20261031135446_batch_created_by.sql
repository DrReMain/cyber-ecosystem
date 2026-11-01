-- Modify "activation_batch" table
ALTER TABLE "public"."activation_batch" ADD COLUMN "created_by" character varying NOT NULL;
-- Set comment to column: "created_by" on table: "activation_batch"
COMMENT ON COLUMN "public"."activation_batch"."created_by" IS 'admin identity that minted the batch (audit join)';
