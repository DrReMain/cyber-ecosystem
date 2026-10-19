-- a service database owns its extensions; the initdb scratch already has it, so this line is hand-prepended
CREATE EXTENSION IF NOT EXISTS vector;

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
