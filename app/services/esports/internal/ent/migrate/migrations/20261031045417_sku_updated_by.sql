-- Modify "sku" table
ALTER TABLE "public"."sku" ADD COLUMN "updated_by" character varying NULL;
-- Set comment to column: "updated_by" on table: "sku"
COMMENT ON COLUMN "public"."sku"."updated_by" IS 'operator id of the last catalog write (audit join); null on seed rows';
