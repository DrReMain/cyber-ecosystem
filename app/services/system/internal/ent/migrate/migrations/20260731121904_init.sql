-- Create "item" table
CREATE TABLE "public"."item" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "sort" character varying NOT NULL COLLATE "C",
  "deleted_at" timestamptz NULL,
  "name" character varying NOT NULL,
  "description" character varying NOT NULL DEFAULT '',
  "status" character varying NOT NULL DEFAULT 'active',
  PRIMARY KEY ("id")
);
-- Create index "item_created_at" to table: "item"
CREATE INDEX "item_created_at" ON "public"."item" ("created_at");
-- Create index "item_name" to table: "item"
CREATE UNIQUE INDEX "item_name" ON "public"."item" ("name") WHERE (deleted_at IS NULL);
-- Create index "item_sort" to table: "item"
CREATE UNIQUE INDEX "item_sort" ON "public"."item" ("sort") WHERE (deleted_at IS NULL);
-- Create index "item_updated_at" to table: "item"
CREATE INDEX "item_updated_at" ON "public"."item" ("updated_at");
-- Set comment to column: "sort" on table: "item"
COMMENT ON COLUMN "public"."item"."sort" IS 'fractional index for ordering';
-- Set comment to column: "status" on table: "item"
COMMENT ON COLUMN "public"."item"."status" IS 'active/inactive';
