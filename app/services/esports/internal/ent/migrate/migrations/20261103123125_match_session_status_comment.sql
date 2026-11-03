-- Set comment to column: "status" on table: "match_session"
COMMENT ON COLUMN "public"."match_session"."status" IS 'active resumes at any age — the lineup is locked in the row; switched = an explicit re-draw closed it; expired = the retired seven-day window close (legacy rows only, no longer produced)';
