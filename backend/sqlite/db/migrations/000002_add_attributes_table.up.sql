CREATE TABLE IF NOT EXISTS {{ .Attributes }} (
  `id` TEXT NOT NULL,
  `instance_id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `data` BLOB NOT NULL,
  PRIMARY KEY(`id`, `instance_id`, `execution_id`)
);

-- Move activity attributes to attributes table
INSERT OR IGNORE INTO {{ .Attributes }} (`id`, `instance_id`, `execution_id`, `data`) SELECT `id`, `instance_id`, `execution_id`, `attributes` FROM {{ .Activities }};
ALTER TABLE {{ .Activities }} DROP COLUMN `attributes`;

-- Move history attributes to attributes table
INSERT OR IGNORE INTO {{ .Attributes }} (`id`, `instance_id`, `execution_id`, `data`) SELECT `id`, `instance_id`, `execution_id`, `attributes` FROM {{ .History }};
ALTER TABLE {{ .History }} DROP COLUMN `attributes`;

-- Move pending_events attributes to attributes table
INSERT OR IGNORE INTO {{ .Attributes }} (`id`, `instance_id`, `execution_id`, `data`) SELECT `id`, `instance_id`, `execution_id`, `attributes` FROM {{ .PendingEvents }};
ALTER TABLE {{ .PendingEvents }} DROP COLUMN `attributes`;
