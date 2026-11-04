-- Modify "compliance_config" table
ALTER TABLE "public"."compliance_config" ADD COLUMN "renewal_remind_days" bigint NOT NULL DEFAULT 7;
-- Set comment to column: "pay_age_line" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."pay_age_line" IS 'members younger than this cannot pay; birth is year+month granularity, so the line clears only from the month after the birthday month — the conservative reading';
-- Set comment to column: "renewal_remind_days" on table: "compliance_config"
COMMENT ON COLUMN "public"."compliance_config"."renewal_remind_days" IS 'days before membership expiry the renewal reminder shows; the shipped default answers when unset';
