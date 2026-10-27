-- Create "agent_config" table
CREATE TABLE "public"."agent_config" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "user_id" character varying NOT NULL,
  "base_url" character varying NOT NULL,
  "api_key" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "agentconfig_created_at" to table: "agent_config"
CREATE INDEX "agentconfig_created_at" ON "public"."agent_config" ("created_at");
-- Create index "agentconfig_tenant_id_created_at" to table: "agent_config"
CREATE INDEX "agentconfig_tenant_id_created_at" ON "public"."agent_config" ("tenant_id", "created_at");
-- Create index "agentconfig_updated_at" to table: "agent_config"
CREATE INDEX "agentconfig_updated_at" ON "public"."agent_config" ("updated_at");
-- Create index "agentconfig_user_id" to table: "agent_config"
CREATE UNIQUE INDEX "agentconfig_user_id" ON "public"."agent_config" ("user_id") WHERE (deleted_at IS NULL);
-- Set comment to column: "tenant_id" on table: "agent_config"
COMMENT ON COLUMN "public"."agent_config"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "agent_config"
COMMENT ON COLUMN "public"."agent_config"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "user_id" on table: "agent_config"
COMMENT ON COLUMN "public"."agent_config"."user_id" IS 'owning user; one live row per user, enforced by the partial unique index';
-- Set comment to column: "base_url" on table: "agent_config"
COMMENT ON COLUMN "public"."agent_config"."base_url" IS 'OpenAI-compatible provider base URL';
-- Set comment to column: "api_key" on table: "agent_config"
COMMENT ON COLUMN "public"."agent_config"."api_key" IS 'AES-GCM ciphertext, base64(nonce||ct); empty = no key set; never returned by any RPC';
-- Create "chat_message" table
CREATE TABLE "public"."chat_message" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "session_id" character varying NOT NULL,
  "role" character varying NOT NULL DEFAULT '',
  "content" text NOT NULL DEFAULT '',
  "reasoning" text NOT NULL DEFAULT '',
  "model" character varying NOT NULL DEFAULT '',
  "finish" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "chatmessage_created_at" to table: "chat_message"
CREATE INDEX "chatmessage_created_at" ON "public"."chat_message" ("created_at");
-- Create index "chatmessage_session_id_created_at" to table: "chat_message"
CREATE INDEX "chatmessage_session_id_created_at" ON "public"."chat_message" ("session_id", "created_at");
-- Create index "chatmessage_updated_at" to table: "chat_message"
CREATE INDEX "chatmessage_updated_at" ON "public"."chat_message" ("updated_at");
-- Set comment to column: "tenant_id" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "session_id" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."session_id" IS 'owning chat_session id; no FK — ownership is enforced by querying through the owner-filtered session';
-- Set comment to column: "role" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."role" IS 'OpenAI-compatible role vocabulary; free-form so S4 tool roles need no migration';
-- Set comment to column: "content" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."content" IS 'turn text; unbounded by design — parts-JSONB is the additive S4+ evolution for multimodal';
-- Set comment to column: "reasoning" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."reasoning" IS 'accumulated model reasoning on the assistant turn; empty on user turns';
-- Set comment to column: "model" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."model" IS 'model id that produced the assistant turn; empty on user turns';
-- Set comment to column: "finish" on table: "chat_message"
COMMENT ON COLUMN "public"."chat_message"."finish" IS 'upstream finish_reason on the assistant turn (stop/length/aborted); empty on user turns';
-- Create "chat_session" table
CREATE TABLE "public"."chat_session" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "owner_id" character varying NOT NULL,
  "title" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "chatsession_created_at" to table: "chat_session"
CREATE INDEX "chatsession_created_at" ON "public"."chat_session" ("created_at");
-- Create index "chatsession_owner_id_updated_at" to table: "chat_session"
CREATE INDEX "chatsession_owner_id_updated_at" ON "public"."chat_session" ("owner_id", "updated_at") WHERE (deleted_at IS NULL);
-- Create index "chatsession_updated_at" to table: "chat_session"
CREATE INDEX "chatsession_updated_at" ON "public"."chat_session" ("updated_at");
-- Set comment to column: "tenant_id" on table: "chat_session"
COMMENT ON COLUMN "public"."chat_session"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "chat_session"
COMMENT ON COLUMN "public"."chat_session"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "owner_id" on table: "chat_session"
COMMENT ON COLUMN "public"."chat_session"."owner_id" IS 'owning user; me-face — every query filters owner_id';
-- Set comment to column: "title" on table: "chat_session"
COMMENT ON COLUMN "public"."chat_session"."title" IS 'first user message clipped to 20 runes + ellipsis; written once at creation';
