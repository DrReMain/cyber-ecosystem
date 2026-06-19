package pg

// schemaStmts 幂等建表语句（逐条 Exec；pgx 扩展协议不支持单次多语句）。在独立的 `mq` 库
// 执行，不走 atlas。messages/deliveries/subscribers/dlq 四张表 + 索引。设计见
// docs/pg-mq-design.md §4。
var schemaStmts = []string{
	`CREATE TABLE IF NOT EXISTS messages (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
	`CREATE INDEX IF NOT EXISTS idx_messages_topic_id ON messages (topic, id)`,

	`CREATE TABLE IF NOT EXISTS deliveries (
  id         BIGSERIAL PRIMARY KEY,
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  deliveries INT NOT NULL DEFAULT 0,
  due_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (group_name, message_id)
)`,
	`DROP INDEX IF EXISTS idx_deliveries_poll`,
	`CREATE INDEX IF NOT EXISTS idx_deliveries_poll ON deliveries (group_name, topic, due_at, message_id)`,

	`CREATE TABLE IF NOT EXISTS subscribers (
  group_name TEXT NOT NULL,
  topic      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_name, topic)
)`,

	`CREATE TABLE IF NOT EXISTS dlq (
  id         BIGSERIAL PRIMARY KEY,
  topic      TEXT NOT NULL,
  group_name TEXT NOT NULL,
  payload    BYTEA NOT NULL,
  headers    JSONB NOT NULL DEFAULT '{}',
  deliveries INT NOT NULL,
  error      TEXT NOT NULL,
  dead_at    TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
}
