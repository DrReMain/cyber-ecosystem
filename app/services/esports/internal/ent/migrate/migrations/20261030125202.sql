-- Modify "entitlement" table
ALTER TABLE "public"."entitlement" ADD COLUMN "status" character varying NOT NULL DEFAULT 'active', ADD COLUMN "terminated_at" timestamptz NULL;
-- Set comment to column: "status" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."status" IS 'the only mutable fact on the ledger: the terminate action flips a live membership row to terminated; expiry itself never lands here — it stays a read-time fact of expires_at vs now, so no sweeper exists';
-- Set comment to column: "terminated_at" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."terminated_at" IS 'when the terminate action struck, set alongside status; null on rows still active';
