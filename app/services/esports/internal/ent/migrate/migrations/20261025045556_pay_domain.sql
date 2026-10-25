-- Create "entitlement" table
CREATE TABLE "public"."entitlement" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "scope_kind" character varying NOT NULL,
  "scope_ref" character varying NULL,
  "expires_at" timestamptz NULL,
  "source" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "entitlement_created_at" to table: "entitlement"
CREATE INDEX "entitlement_created_at" ON "public"."entitlement" ("created_at");
-- Create index "entitlement_tenant_id_user_id_scope_kind_scope_ref" to table: "entitlement"
CREATE UNIQUE INDEX "entitlement_tenant_id_user_id_scope_kind_scope_ref" ON "public"."entitlement" ("tenant_id", "user_id", "scope_kind", "scope_ref");
-- Create index "entitlement_updated_at" to table: "entitlement"
CREATE INDEX "entitlement_updated_at" ON "public"."entitlement" ("updated_at");
-- Set comment to column: "tenant_id" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "user_id" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."user_id" IS 'holding member identity (system user id), logical reference — same key law as profile';
-- Set comment to column: "scope_kind" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_kind" IS 'grant shape: trial points at one record permanently, membership opens a window';
-- Set comment to column: "scope_ref" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."scope_ref" IS 'trial: the unlocked test record id; membership: null — renewal rows stack instead of dedup';
-- Set comment to column: "expires_at" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."expires_at" IS 'membership window end (paid_at + period days); null on trial rows — permanence is the absence of a deadline';
-- Set comment to column: "source" on table: "entitlement"
COMMENT ON COLUMN "public"."entitlement"."source" IS 'granting artifact id — the order number today; activation codes join this column when they land';
-- Create "pay_order" table
CREATE TABLE "public"."pay_order" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "order_no" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "sku_code" character varying NOT NULL,
  "scope_kind" character varying NOT NULL,
  "scope_ref" character varying NULL,
  "provider" character varying NOT NULL,
  "channel" character varying NOT NULL,
  "amount_fen" bigint NOT NULL,
  "state" character varying NOT NULL DEFAULT 'pending',
  "credential_kind" character varying NOT NULL,
  "credential_value" character varying NOT NULL,
  "transaction_id" character varying NULL,
  "expires_at" timestamptz NOT NULL,
  "paid_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "payorder_created_at" to table: "pay_order"
CREATE INDEX "payorder_created_at" ON "public"."pay_order" ("created_at");
-- Create index "payorder_order_no" to table: "pay_order"
CREATE UNIQUE INDEX "payorder_order_no" ON "public"."pay_order" ("order_no");
-- Create index "payorder_state_expires_at" to table: "pay_order"
CREATE INDEX "payorder_state_expires_at" ON "public"."pay_order" ("state", "expires_at");
-- Create index "payorder_tenant_id_user_id_created_at" to table: "pay_order"
CREATE INDEX "payorder_tenant_id_user_id_created_at" ON "public"."pay_order" ("tenant_id", "user_id", "created_at");
-- Create index "payorder_tenant_id_user_id_scope_kind_scope_ref" to table: "pay_order"
CREATE UNIQUE INDEX "payorder_tenant_id_user_id_scope_kind_scope_ref" ON "public"."pay_order" ("tenant_id", "user_id", "scope_kind", "scope_ref") WHERE ((state)::text = 'pending'::text);
-- Create index "payorder_updated_at" to table: "pay_order"
CREATE INDEX "payorder_updated_at" ON "public"."pay_order" ("updated_at");
-- Set comment to column: "tenant_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "order_no" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."order_no" IS 'business order number (''P''+ULID) — the channel idempotency key (out_trade_no) and the audit join key into bills';
-- Set comment to column: "user_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."user_id" IS 'buying member identity (system user id), logical reference — same key law as profile';
-- Set comment to column: "sku_code" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."sku_code" IS 'sku catalog code at purchase time (snapshot; later sku edits never rewrite history)';
-- Set comment to column: "scope_kind" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."scope_kind" IS 'grant shape: trial unlocks one test record (scope_ref), period opens the membership window';
-- Set comment to column: "scope_ref" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."scope_ref" IS 'trial: the test record id being unlocked; period: null — the window shape comes from the sku';
-- Set comment to column: "provider" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."provider" IS 'payment provider name (wechat, mock) — the key the platform''s container map routes query/close by';
-- Set comment to column: "channel" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."channel" IS 'access-context channel value (provider vocabulary, e.g. wechat_jsapi); fixes how the credential is consumed';
-- Set comment to column: "amount_fen" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."amount_fen" IS 'agreed amount in fen, frozen at creation; settlement verifies the channel-reported amount against it';
-- Set comment to column: "state" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."state" IS 'channel-state projection; paid is terminal — refunds are derived facts, not states';
-- Set comment to column: "credential_kind" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."credential_kind" IS 'how the page consumes the credential (invoke|jump_url|qr_code), snapshotted with the value';
-- Set comment to column: "credential_value" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."credential_value" IS 'channel credential at creation; same-scope reuse replays it with zero channel round trips';
-- Set comment to column: "transaction_id" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."transaction_id" IS 'channel transaction id, backfilled at settlement — the bill-reconciliation join key';
-- Set comment to column: "expires_at" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."expires_at" IS 'payment deadline carried from creation (channel default two hours); expiry closes pending orders';
-- Set comment to column: "paid_at" on table: "pay_order"
COMMENT ON COLUMN "public"."pay_order"."paid_at" IS 'channel-side payment time; the entitlement window starts here for membership';
-- Create "sku" table
CREATE TABLE "public"."sku" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "code" character varying NOT NULL,
  "kind" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "price_fen" bigint NOT NULL,
  "list_price_fen" bigint NOT NULL,
  "period_days" bigint NULL,
  "on_shelf" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "sku_code" to table: "sku"
CREATE UNIQUE INDEX "sku_code" ON "public"."sku" ("code") WHERE (deleted_at IS NULL);
-- Create index "sku_created_at" to table: "sku"
CREATE INDEX "sku_created_at" ON "public"."sku" ("created_at");
-- Create index "sku_updated_at" to table: "sku"
CREATE INDEX "sku_updated_at" ON "public"."sku" ("updated_at");
-- Set comment to column: "tenant_id" on table: "sku"
COMMENT ON COLUMN "public"."sku"."tenant_id" IS 'tenant scope; back-filled on insert and filtered from the request subject';
-- Set comment to column: "deleted_at" on table: "sku"
COMMENT ON COLUMN "public"."sku"."deleted_at" IS 'soft-delete marker; null = live row (delete becomes an update)';
-- Set comment to column: "code" on table: "sku"
COMMENT ON COLUMN "public"."sku"."code" IS 'catalog code join key (e.g. member_month); orders snapshot it, the row ULID never crosses a boundary';
-- Set comment to column: "kind" on table: "sku"
COMMENT ON COLUMN "public"."sku"."kind" IS 'what the sku sells: one trial''s report permanently vs a membership window';
-- Set comment to column: "name" on table: "sku"
COMMENT ON COLUMN "public"."sku"."name" IS 'display title (pay sheet line); also the channel order description source';
-- Set comment to column: "description" on table: "sku"
COMMENT ON COLUMN "public"."sku"."description" IS 'pay-sheet blurb under the title';
-- Set comment to column: "price_fen" on table: "sku"
COMMENT ON COLUMN "public"."sku"."price_fen" IS 'selling price in fen (1/100 yuan) — money stays integer fen end to end';
-- Set comment to column: "list_price_fen" on table: "sku"
COMMENT ON COLUMN "public"."sku"."list_price_fen" IS 'list (pre-discount) price; equals price_fen while no discount is shown, carried so discounting stays a value change';
-- Set comment to column: "period_days" on table: "sku"
COMMENT ON COLUMN "public"."sku"."period_days" IS 'membership window in days (31: the 32nd day expires); null on trial skus';
-- Set comment to column: "on_shelf" on table: "sku"
COMMENT ON COLUMN "public"."sku"."on_shelf" IS 'member-plane visibility switch; off-shelf rows stay for order-history joins';
