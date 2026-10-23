-- Create index "testrecord_session_id_test_code_attempt_no" to table: "test_record"
CREATE UNIQUE INDEX "testrecord_session_id_test_code_attempt_no" ON "public"."test_record" ("session_id", "test_code", "attempt_no") WHERE (session_id IS NOT NULL);
-- Set comment to column: "attempt_no" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."attempt_no" IS 'attempt ordinal inside the match session; the partial unique (session_id, test_code, attempt_no) makes the count-then-insert claim atomic under concurrent submits';
-- Set comment to column: "session_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."session_id" IS 'match session reference — null on practice attempts, set on session attempts';
