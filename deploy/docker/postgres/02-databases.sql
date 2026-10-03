-- Database shells only; schemas are owned by their consumers:
--   atlas_dev -> scratch schema for `system:migrate:diff` (Atlas rebuilds it per diff)
--   mq        -> shared-go/capability/mq/pg (created on first connect)
--   system    -> ent/Atlas migrations (system:migrate:apply)
--   agent     -> ent/Atlas migrations (agent:migrate:apply)
-- Names are shared facts with the service configs (db_name / mq dsn).
CREATE DATABASE atlas_dev;
\connect atlas_dev
DO $$ BEGIN EXECUTE 'CREATE EXTENSION IF NOT EXISTS vector'; EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'vector not installed, skipping'; END $$;
\connect postgres

CREATE DATABASE mq;
CREATE DATABASE system;
CREATE DATABASE agent;
