-- Set comment to column: "score" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."score" IS 'the ability contribution this attempt scored — written by the scoring engine, never the client';
-- Set comment to column: "percentile" on table: "ability_score"
COMMENT ON COLUMN "public"."ability_score"."percentile" IS 'same-age percentile once the cohort math lands; null until then';
-- Set comment to column: "name" on table: "test"
COMMENT ON COLUMN "public"."test"."name" IS 'display title; seeds carry the placeholder code until the metric sheet calibrates';
-- Set comment to column: "ability" on table: "test"
COMMENT ON COLUMN "public"."test"."ability" IS 'primary ability, six-word vocabulary reactionSpeed|aimingAccuracy|trackingStability|skillReleaseAccuracy|resourceAllocation|decisionEfficiency; attribution is metric-sheet data pending calibration';
-- Set comment to column: "attempt_no" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."attempt_no" IS 'attempt ordinal inside the match session; uniqueness is the match-session law, not a DB constraint yet';
-- Set comment to column: "session_id" on table: "test_record"
COMMENT ON COLUMN "public"."test_record"."session_id" IS 'match session reference — null until the match-session aggregate lands, then backfilled';
