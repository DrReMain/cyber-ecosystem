-- Create "agreement_consent" table
CREATE TABLE "public"."agreement_consent" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" character varying NOT NULL,
  "doc" character varying NOT NULL,
  "version_no" character varying NOT NULL,
  "login_method" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "agreementconsent_created_at" to table: "agreement_consent"
CREATE INDEX "agreementconsent_created_at" ON "public"."agreement_consent" ("created_at");
-- Create index "agreementconsent_doc_version_no" to table: "agreement_consent"
CREATE INDEX "agreementconsent_doc_version_no" ON "public"."agreement_consent" ("doc", "version_no");
-- Create index "agreementconsent_updated_at" to table: "agreement_consent"
CREATE INDEX "agreementconsent_updated_at" ON "public"."agreement_consent" ("updated_at");
-- Create index "agreementconsent_user_id" to table: "agreement_consent"
CREATE INDEX "agreementconsent_user_id" ON "public"."agreement_consent" ("user_id");
-- Set comment to column: "user_id" on table: "agreement_consent"
COMMENT ON COLUMN "public"."agreement_consent"."user_id" IS 'the consenting member identity';
-- Set comment to column: "doc" on table: "agreement_consent"
COMMENT ON COLUMN "public"."agreement_consent"."doc" IS 'the document the checkbox bound to';
-- Set comment to column: "version_no" on table: "agreement_consent"
COMMENT ON COLUMN "public"."agreement_consent"."version_no" IS 'the version string at consent time, denormalized so the evidence outlives the version row';
-- Set comment to column: "login_method" on table: "agreement_consent"
COMMENT ON COLUMN "public"."agreement_consent"."login_method" IS 'the UI path the consent travelled on, declared by the client (the system owns the real auth facts)';
-- Create "agreement_version" table
CREATE TABLE "public"."agreement_version" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "doc" character varying NOT NULL,
  "no" character varying NOT NULL,
  "title" character varying NOT NULL,
  "body" text NOT NULL,
  "effective_at" timestamptz NOT NULL,
  "is_current" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "agreementversion_created_at" to table: "agreement_version"
CREATE INDEX "agreementversion_created_at" ON "public"."agreement_version" ("created_at");
-- Create index "agreementversion_doc_no" to table: "agreement_version"
CREATE UNIQUE INDEX "agreementversion_doc_no" ON "public"."agreement_version" ("doc", "no");
-- Create index "agreementversion_updated_at" to table: "agreement_version"
CREATE INDEX "agreementversion_updated_at" ON "public"."agreement_version" ("updated_at");
-- Set comment to column: "doc" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."doc" IS 'the legal document this version publishes';
-- Set comment to column: "no" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."no" IS 'operator-quoted version string (v1.0.0…); unique per document, format-free';
-- Set comment to column: "title" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."title" IS 'display title derived from the document type at create; historical rows keep what they were published under';
-- Set comment to column: "body" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."body" IS 'the published legal text — the compliance archive an operator quotes in a dispute';
-- Set comment to column: "effective_at" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."effective_at" IS 'the date this version started binding, shown on the legal page''s version line';
-- Set comment to column: "is_current" on table: "agreement_version"
COMMENT ON COLUMN "public"."agreement_version"."is_current" IS 'true = the version every new consent lands on; exactly one per document, flipped in one transaction';
