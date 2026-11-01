-- Modify "activation_code" table
ALTER TABLE "public"."activation_code" ADD COLUMN "redeemed_by" character varying NULL, ADD COLUMN "redeemed_at" timestamptz NULL;
-- Set comment to column: "redeemed_by" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."redeemed_by" IS 'the member the single use went to, set by the same conditional UPDATE that takes it — the code row is the redeem''s own record, the family grant it produced may later be ratcheted by a newer code and lose this trail';
-- Set comment to column: "redeemed_at" on table: "activation_code"
COMMENT ON COLUMN "public"."activation_code"."redeemed_at" IS 'when the use was taken; written with redeemed_by in the one arbitration';
