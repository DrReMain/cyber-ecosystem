-- Create "announcement" table
CREATE TABLE "public"."announcement" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "kind" character varying NOT NULL,
  "scope" character varying NOT NULL,
  "title" character varying NOT NULL,
  "content" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'draft',
  "published_at" timestamptz NULL,
  "offline_at" timestamptz NULL,
  "published_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "announcement_created_at" to table: "announcement"
CREATE INDEX "announcement_created_at" ON "public"."announcement" ("created_at");
-- Create index "announcement_kind_status_published_at" to table: "announcement"
CREATE INDEX "announcement_kind_status_published_at" ON "public"."announcement" ("kind", "status", "published_at");
-- Create index "announcement_updated_at" to table: "announcement"
CREATE INDEX "announcement_updated_at" ON "public"."announcement" ("updated_at");
-- Set comment to column: "kind" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."kind" IS 'banner = live on the home surface; message = delivered into letterboxes at publish';
-- Set comment to column: "scope" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."scope" IS 'who receives a message-kind row at publish: every roster member, or live membership holders';
-- Set comment to column: "title" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."title" IS 'the headline both surfaces render';
-- Set comment to column: "content" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."content" IS 'the body; editable while draft, frozen in every delivered copy at publish';
-- Set comment to column: "status" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."status" IS 'the one-way lifecycle: draft→published→offline, no revival';
-- Set comment to column: "published_at" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."published_at" IS 'absent until published — the delivery moment';
-- Set comment to column: "offline_at" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."offline_at" IS 'absent until offlined';
-- Set comment to column: "published_by" on table: "announcement"
COMMENT ON COLUMN "public"."announcement"."published_by" IS 'the operator who published; drafts carry none';
-- Create "feedback_message" table
CREATE TABLE "public"."feedback_message" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "feedback_id" character varying NOT NULL,
  "sender" character varying NOT NULL,
  "content" character varying NOT NULL,
  "admin_id" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "feedbackmessage_created_at" to table: "feedback_message"
CREATE INDEX "feedbackmessage_created_at" ON "public"."feedback_message" ("created_at");
-- Create index "feedbackmessage_feedback_id_created_at" to table: "feedback_message"
CREATE INDEX "feedbackmessage_feedback_id_created_at" ON "public"."feedback_message" ("feedback_id", "created_at");
-- Create index "feedbackmessage_updated_at" to table: "feedback_message"
CREATE INDEX "feedbackmessage_updated_at" ON "public"."feedback_message" ("updated_at");
-- Set comment to column: "feedback_id" on table: "feedback_message"
COMMENT ON COLUMN "public"."feedback_message"."feedback_id" IS 'the conversation root this turn belongs to';
-- Set comment to column: "sender" on table: "feedback_message"
COMMENT ON COLUMN "public"."feedback_message"."sender" IS 'who spoke this turn';
-- Set comment to column: "content" on table: "feedback_message"
COMMENT ON COLUMN "public"."feedback_message"."content" IS 'the turn''s text';
-- Set comment to column: "admin_id" on table: "feedback_message"
COMMENT ON COLUMN "public"."feedback_message"."admin_id" IS 'the operator behind an admin turn; absent on member turns';
-- Create "user_feedback" table
CREATE TABLE "public"."user_feedback" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" character varying NOT NULL,
  "category" character varying NOT NULL,
  "content" character varying NOT NULL,
  "contact" character varying NULL,
  "status" character varying NOT NULL DEFAULT 'pending',
  PRIMARY KEY ("id")
);
-- Create index "userfeedback_created_at" to table: "user_feedback"
CREATE INDEX "userfeedback_created_at" ON "public"."user_feedback" ("created_at");
-- Create index "userfeedback_status_created_at" to table: "user_feedback"
CREATE INDEX "userfeedback_status_created_at" ON "public"."user_feedback" ("status", "created_at");
-- Create index "userfeedback_updated_at" to table: "user_feedback"
CREATE INDEX "userfeedback_updated_at" ON "public"."user_feedback" ("updated_at");
-- Create index "userfeedback_user_id_updated_at" to table: "user_feedback"
CREATE INDEX "userfeedback_user_id_updated_at" ON "public"."user_feedback" ("user_id", "updated_at");
-- Set comment to column: "user_id" on table: "user_feedback"
COMMENT ON COLUMN "public"."user_feedback"."user_id" IS 'the asking member identity';
-- Set comment to column: "category" on table: "user_feedback"
COMMENT ON COLUMN "public"."user_feedback"."category" IS 'the submit form''s six problem types';
-- Set comment to column: "content" on table: "user_feedback"
COMMENT ON COLUMN "public"."user_feedback"."content" IS 'the first message of the conversation, denormalized for the list rows';
-- Set comment to column: "contact" on table: "user_feedback"
COMMENT ON COLUMN "public"."user_feedback"."contact" IS 'where the operator may reach the member; absent = in-app contact only';
-- Set comment to column: "status" on table: "user_feedback"
COMMENT ON COLUMN "public"."user_feedback"."status" IS 'the conversation''s state: follow-up walks replied back to pending; closed is terminal';
-- Create "user_message" table
CREATE TABLE "public"."user_message" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "title" character varying NOT NULL,
  "content" character varying NOT NULL,
  "ref_id" character varying NULL,
  "is_read" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "usermessage_created_at" to table: "user_message"
CREATE INDEX "usermessage_created_at" ON "public"."user_message" ("created_at");
-- Create index "usermessage_updated_at" to table: "user_message"
CREATE INDEX "usermessage_updated_at" ON "public"."user_message" ("updated_at");
-- Create index "usermessage_user_id_created_at" to table: "user_message"
CREATE INDEX "usermessage_user_id_created_at" ON "public"."user_message" ("user_id", "created_at");
-- Create index "usermessage_user_id_is_read" to table: "user_message"
CREATE INDEX "usermessage_user_id_is_read" ON "public"."user_message" ("user_id", "is_read");
-- Set comment to column: "user_id" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."user_id" IS 'the receiving member identity';
-- Set comment to column: "type" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."type" IS 'what kind of letter this is; decides the ref''s domain (report→test record, feedback_reply→feedback)';
-- Set comment to column: "title" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."title" IS 'the headline the sender composed; every writer sets one, no fallback shape exists';
-- Set comment to column: "content" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."content" IS 'the letter body, denormalized at delivery so later source changes never rewrite a sent letter';
-- Set comment to column: "ref_id" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."ref_id" IS 'deep-link target whose domain the type names; absent when the letter carries no link';
-- Set comment to column: "is_read" on table: "user_message"
COMMENT ON COLUMN "public"."user_message"."is_read" IS 'the row''s own unread state — the only mutable column, flipped by mark-read';
