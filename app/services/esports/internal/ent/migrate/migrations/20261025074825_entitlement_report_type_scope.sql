-- Set comment to column: "scope_kind" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_kind" IS 'grant shape: trial points at one record permanently, membership opens a window, report_type opens a report family via a type code';
-- Set comment to column: "scope_ref" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_ref" IS 'trial: the unlocked test record id; report_type: the family name (dedup guard); membership: null — renewal rows stack instead of dedup';
