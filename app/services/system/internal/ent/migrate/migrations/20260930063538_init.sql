-- Create "audit_log" table
CREATE TABLE "public"."audit_log" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant" character varying NOT NULL DEFAULT '',
  "actor" character varying NOT NULL DEFAULT '',
  "principal_type" character varying NOT NULL DEFAULT '',
  "operation" character varying NOT NULL,
  "http_method" character varying NOT NULL DEFAULT '',
  "http_path" character varying NOT NULL DEFAULT '',
  "status" bigint NOT NULL DEFAULT 0,
  "latency_ms" bigint NOT NULL DEFAULT 0,
  "ip" character varying NOT NULL DEFAULT '',
  "user_agent" character varying NOT NULL DEFAULT '',
  "deny_reason" character varying NOT NULL DEFAULT '',
  "event_id" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "auditlog_created_at" to table: "audit_log"
CREATE INDEX "auditlog_created_at" ON "public"."audit_log" ("created_at");
-- Create index "auditlog_event_id" to table: "audit_log"
CREATE UNIQUE INDEX "auditlog_event_id" ON "public"."audit_log" ("event_id");
-- Create index "auditlog_tenant_created_at" to table: "audit_log"
CREATE INDEX "auditlog_tenant_created_at" ON "public"."audit_log" ("tenant", "created_at");
-- Create index "auditlog_updated_at" to table: "audit_log"
CREATE INDEX "auditlog_updated_at" ON "public"."audit_log" ("updated_at");
-- Set comment to column: "tenant" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."tenant" IS 'subject tenant; empty on denied or subject-less events';
-- Set comment to column: "actor" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."actor" IS 'subject user id; empty on subject-less events';
-- Set comment to column: "principal_type" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."principal_type" IS 'user; empty together with actor';
-- Set comment to column: "operation" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."operation" IS 'proto operation path, e.g. /cyber.system.v1.UserService/CreateUser';
-- Set comment to column: "status" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."status" IS 'HTTP status of the outcome';
-- Set comment to column: "deny_reason" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."deny_reason" IS 'NO_GRANT | ABAC_CONSTRAINT; empty when allowed';
-- Set comment to column: "event_id" on table: "audit_log"
COMMENT ON COLUMN "public"."audit_log"."event_id" IS 'publisher-generated dedup key; unique index absorbs redelivery';
-- Create "dept" table
CREATE TABLE "public"."dept" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying NOT NULL,
  "parent_id" character varying NULL,
  "remark" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "dept_created_at" to table: "dept"
CREATE INDEX "dept_created_at" ON "public"."dept" ("created_at");
-- Create index "dept_tenant_id_parent_id_name" to table: "dept"
CREATE UNIQUE INDEX "dept_tenant_id_parent_id_name" ON "public"."dept" ("tenant_id", "parent_id", "name") WHERE (deleted_at IS NULL);
-- Create index "dept_updated_at" to table: "dept"
CREATE INDEX "dept_updated_at" ON "public"."dept" ("updated_at");
-- Set comment to column: "tenant_id" on table: "dept"
COMMENT ON COLUMN "public"."dept"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "dept"
COMMENT ON COLUMN "public"."dept"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "parent_id" on table: "dept"
COMMENT ON COLUMN "public"."dept"."parent_id" IS 'null = tree root';
-- Create "file" table
CREATE TABLE "public"."file" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "key" character varying NOT NULL,
  "name" character varying NOT NULL,
  "content_type" character varying NOT NULL DEFAULT '',
  "size" bigint NOT NULL DEFAULT 0,
  "source" character varying NOT NULL DEFAULT 'client_upload',
  "status" character varying NOT NULL DEFAULT 'uploading',
  "upload_id" character varying NOT NULL DEFAULT '',
  "etag" character varying NOT NULL DEFAULT '',
  "owner_id" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "file_created_at" to table: "file"
CREATE INDEX "file_created_at" ON "public"."file" ("created_at");
-- Create index "file_key" to table: "file"
CREATE UNIQUE INDEX "file_key" ON "public"."file" ("key") WHERE (deleted_at IS NULL);
-- Create index "file_tenant_id_created_at" to table: "file"
CREATE INDEX "file_tenant_id_created_at" ON "public"."file" ("tenant_id", "created_at");
-- Create index "file_updated_at" to table: "file"
CREATE INDEX "file_updated_at" ON "public"."file" ("updated_at");
-- Set comment to column: "tenant_id" on table: "file"
COMMENT ON COLUMN "public"."file"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "file"
COMMENT ON COLUMN "public"."file"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "key" on table: "file"
COMMENT ON COLUMN "public"."file"."key" IS 'storage object key; server-generated prefix+xid, never client-chosen';
-- Set comment to column: "name" on table: "file"
COMMENT ON COLUMN "public"."file"."name" IS 'display name; decoupled from the stored key';
-- Set comment to column: "content_type" on table: "file"
COMMENT ON COLUMN "public"."file"."content_type" IS 'declared at upload, backfilled from HEAD at confirm; storage wins';
-- Set comment to column: "size" on table: "file"
COMMENT ON COLUMN "public"."file"."size" IS 'declared at create, replaced by the measured size at confirm';
-- Set comment to column: "source" on table: "file"
COMMENT ON COLUMN "public"."file"."source" IS 'client_upload | server_generated';
-- Set comment to column: "status" on table: "file"
COMMENT ON COLUMN "public"."file"."status" IS 'uploading | confirmed; aborted uploads hard-delete the row';
-- Set comment to column: "upload_id" on table: "file"
COMMENT ON COLUMN "public"."file"."upload_id" IS 'S3 multipart session id; empty on the single-PUT path';
-- Set comment to column: "etag" on table: "file"
COMMENT ON COLUMN "public"."file"."etag" IS 'object ETag captured at confirm; integrity/dedup seam';
-- Set comment to column: "owner_id" on table: "file"
COMMENT ON COLUMN "public"."file"."owner_id" IS 'creating user; datascope self dimension';
-- Create "permission" table
CREATE TABLE "public"."permission" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "role_id" character varying NOT NULL,
  "operation" character varying NOT NULL,
  "effect" character varying NOT NULL DEFAULT 'allow',
  "scope_kind" character varying NOT NULL DEFAULT '',
  "scope_params" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create index "permission_created_at" to table: "permission"
CREATE INDEX "permission_created_at" ON "public"."permission" ("created_at");
-- Create index "permission_role_id_operation" to table: "permission"
CREATE UNIQUE INDEX "permission_role_id_operation" ON "public"."permission" ("role_id", "operation");
-- Create index "permission_updated_at" to table: "permission"
CREATE INDEX "permission_updated_at" ON "public"."permission" ("updated_at");
-- Set comment to column: "tenant_id" on table: "permission"
COMMENT ON COLUMN "public"."permission"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "operation" on table: "permission"
COMMENT ON COLUMN "public"."permission"."operation" IS 'operation pattern: exact match or prefix wildcard ending in /*';
-- Set comment to column: "effect" on table: "permission"
COMMENT ON COLUMN "public"."permission"."effect" IS 'reserved; always ''allow'' in v1 (no deny)';
-- Set comment to column: "scope_kind" on table: "permission"
COMMENT ON COLUMN "public"."permission"."scope_kind" IS 'row-scope narrowing: '''' | all | self | dept_tree; empty = tenant-wide';
-- Set comment to column: "scope_params" on table: "permission"
COMMENT ON COLUMN "public"."permission"."scope_params" IS 'scope payload (e.g. dept_ids); shape per scope_kind';
-- Create "permission_policy" table
CREATE TABLE "public"."permission_policy" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "permission_id" character varying NOT NULL,
  "policy_id" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "permissionpolicy_created_at" to table: "permission_policy"
CREATE INDEX "permissionpolicy_created_at" ON "public"."permission_policy" ("created_at");
-- Create index "permissionpolicy_permission_id_policy_id" to table: "permission_policy"
CREATE UNIQUE INDEX "permissionpolicy_permission_id_policy_id" ON "public"."permission_policy" ("permission_id", "policy_id");
-- Create index "permissionpolicy_updated_at" to table: "permission_policy"
CREATE INDEX "permissionpolicy_updated_at" ON "public"."permission_policy" ("updated_at");
-- Set comment to column: "tenant_id" on table: "permission_policy"
COMMENT ON COLUMN "public"."permission_policy"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Create "policy" table
CREATE TABLE "public"."policy" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "kind" character varying NOT NULL,
  "name" character varying NOT NULL,
  "params" jsonb NOT NULL,
  "enabled" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "authzpolicy_created_at" to table: "policy"
CREATE INDEX "authzpolicy_created_at" ON "public"."policy" ("created_at");
-- Create index "authzpolicy_tenant_id_kind_name" to table: "policy"
CREATE UNIQUE INDEX "authzpolicy_tenant_id_kind_name" ON "public"."policy" ("tenant_id", "kind", "name");
-- Create index "authzpolicy_updated_at" to table: "policy"
CREATE INDEX "authzpolicy_updated_at" ON "public"."policy" ("updated_at");
-- Set comment to column: "tenant_id" on table: "policy"
COMMENT ON COLUMN "public"."policy"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "kind" on table: "policy"
COMMENT ON COLUMN "public"."policy"."kind" IS 'policy type; must match a registered engine plugin';
-- Set comment to column: "params" on table: "policy"
COMMENT ON COLUMN "public"."policy"."params" IS 'typed plugin payload as jsonb; read whole, never queried internally';
-- Set comment to column: "enabled" on table: "policy"
COMMENT ON COLUMN "public"."policy"."enabled" IS 'false unlinks the constraint at compile (relax direction)';
-- Create "principal_role" table
CREATE TABLE "public"."principal_role" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "principal_type" character varying NOT NULL,
  "principal_id" character varying NOT NULL,
  "role_id" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "principalrole_created_at" to table: "principal_role"
CREATE INDEX "principalrole_created_at" ON "public"."principal_role" ("created_at");
-- Create index "principalrole_principal_type_principal_id_role_id" to table: "principal_role"
CREATE UNIQUE INDEX "principalrole_principal_type_principal_id_role_id" ON "public"."principal_role" ("principal_type", "principal_id", "role_id");
-- Create index "principalrole_updated_at" to table: "principal_role"
CREATE INDEX "principalrole_updated_at" ON "public"."principal_role" ("updated_at");
-- Set comment to column: "tenant_id" on table: "principal_role"
COMMENT ON COLUMN "public"."principal_role"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "principal_type" on table: "principal_role"
COMMENT ON COLUMN "public"."principal_role"."principal_type" IS 'user; plain string so new kinds need no migration';
-- Create "role" table
CREATE TABLE "public"."role" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "enabled" boolean NOT NULL DEFAULT true,
  "remark" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "role_created_at" to table: "role"
CREATE INDEX "role_created_at" ON "public"."role" ("created_at");
-- Create index "role_tenant_id_code" to table: "role"
CREATE UNIQUE INDEX "role_tenant_id_code" ON "public"."role" ("tenant_id", "code");
-- Create index "role_updated_at" to table: "role"
CREATE INDEX "role_updated_at" ON "public"."role" ("updated_at");
-- Set comment to column: "tenant_id" on table: "role"
COMMENT ON COLUMN "public"."role"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "code" on table: "role"
COMMENT ON COLUMN "public"."role"."code" IS 'stable identifier; user role bindings reference codes, not ids';
-- Set comment to column: "enabled" on table: "role"
COMMENT ON COLUMN "public"."role"."enabled" IS 'false drops the role''s grants at engine compile (tighten direction)';
-- Create "user" table
CREATE TABLE "public"."user" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "email" character varying NOT NULL,
  "password_hash" character varying NOT NULL,
  "dept_id" character varying NULL,
  "avatar" character varying NULL,
  "enabled" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "user_created_at" to table: "user"
CREATE INDEX "user_created_at" ON "public"."user" ("created_at");
-- Create index "user_tenant_id_email" to table: "user"
CREATE UNIQUE INDEX "user_tenant_id_email" ON "public"."user" ("tenant_id", "email") WHERE (deleted_at IS NULL);
-- Create index "user_updated_at" to table: "user"
CREATE INDEX "user_updated_at" ON "public"."user" ("updated_at");
-- Set comment to column: "tenant_id" on table: "user"
COMMENT ON COLUMN "public"."user"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "user"
COMMENT ON COLUMN "public"."user"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "password_hash" on table: "user"
COMMENT ON COLUMN "public"."user"."password_hash" IS 'argon2 digest; plaintext is never stored';
-- Set comment to column: "avatar" on table: "user"
COMMENT ON COLUMN "public"."user"."avatar" IS 'referenced file id; display URLs are minted at read time';
-- Set comment to column: "enabled" on table: "user"
COMMENT ON COLUMN "public"."user"."enabled" IS 'false rejects login and drops grants at engine compile';
