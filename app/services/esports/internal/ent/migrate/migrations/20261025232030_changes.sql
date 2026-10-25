-- Modify "profile" table
ALTER TABLE "public"."profile" ADD COLUMN "sex" character varying NULL;
-- Set comment to column: "sex" on table: "profile"
COMMENT ON COLUMN "public"."profile"."sex" IS 'three states by the sex ruling — undetermined is the honest not-said; the report cohort reads it (same-age-same-sex, undetermined benches against the same-age pool)';
