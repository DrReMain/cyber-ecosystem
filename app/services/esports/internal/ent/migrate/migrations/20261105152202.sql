-- Modify "pay_order" table
ALTER TABLE "public"."pay_order" ADD COLUMN "refund_no" character varying NULL, ADD COLUMN "refund_id" character varying NULL, ADD COLUMN "refunded_fen" bigint NULL, ADD COLUMN "refund_reason" character varying NULL, ADD COLUMN "refund_state" character varying NULL, ADD COLUMN "refunded_at" timestamptz NULL;
-- Set comment to column: "refund_no" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refund_no" IS 'out_refund_no of the order''s one full refund (''R''+order_no); null = never refunded — the refund is a derived fact on a paid order, not a state';
-- Set comment to column: "refund_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refund_id" IS 'channel refund id backfilled from the refund answer — the support/portal join key';
-- Set comment to column: "refunded_fen" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refunded_fen" IS 'refunded amount in fen; the full-refund tool writes the order amount, partial refunds are not a product shape';
-- Set comment to column: "refund_reason" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refund_reason" IS 'the operator''s recorded reason for the offline exception — the audit trail the FAQ''s manual handling assumes';
-- Set comment to column: "refund_state" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refund_state" IS 'channel refund lifecycle; abnormal is terminal-for-now (the channel''s manual-intervention bucket), processing converges through the refund query';
-- Set comment to column: "refunded_at" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."refunded_at" IS 'channel-side refund success time; null until the refund succeeds';
