-- Create "account_merges" table
CREATE TABLE "public"."account_merges" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "poor_id" character varying NOT NULL,
  "rich_id" character varying NOT NULL,
  "provider" character varying NOT NULL,
  "kind" character varying NOT NULL,
  "identifier" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "accountmerge_created_at" to table: "account_merges"
CREATE INDEX "accountmerge_created_at" ON "public"."account_merges" ("created_at");
-- Create index "accountmerge_poor_id" to table: "account_merges"
CREATE UNIQUE INDEX "accountmerge_poor_id" ON "public"."account_merges" ("poor_id");
-- Create index "accountmerge_rich_id" to table: "account_merges"
CREATE INDEX "accountmerge_rich_id" ON "public"."account_merges" ("rich_id");
-- Create index "accountmerge_updated_at" to table: "account_merges"
CREATE INDEX "accountmerge_updated_at" ON "public"."account_merges" ("updated_at");
-- Set comment to column: "poor_id" on table: "account_merges"
COMMENT ON COLUMN "public"."account_merges"."poor_id" IS 'deactivated account id; unique — an account can be absorbed at most once';
-- Set comment to column: "rich_id" on table: "account_merges"
COMMENT ON COLUMN "public"."account_merges"."rich_id" IS 'surviving account id that took over the credential';
-- Set comment to column: "provider" on table: "account_merges"
COMMENT ON COLUMN "public"."account_merges"."provider" IS 'provider config key whose exchange proved control';
-- Set comment to column: "kind" on table: "account_merges"
COMMENT ON COLUMN "public"."account_merges"."kind" IS 'credential kind that moved (credential.kind)';
-- Set comment to column: "identifier" on table: "account_merges"
COMMENT ON COLUMN "public"."account_merges"."identifier" IS 'credential identifier that moved (credential.identifier)';
