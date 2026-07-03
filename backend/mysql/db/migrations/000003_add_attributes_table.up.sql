CREATE TABLE IF NOT EXISTS {{ .Attributes }} (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `event_id` NVARCHAR(128) NOT NULL,
  `instance_id` NVARCHAR(128) NOT NULL,
  `execution_id` NVARCHAR(128) NOT NULL,
  `data` MEDIUMBLOB NOT NULL,

  UNIQUE INDEX {{ .IdxAttributesInstanceIDExecutionIDEventID }} (`instance_id`, `execution_id`, `event_id`),
  INDEX {{ .IdxAttributesEventID }} (`event_id`)
);

-- Move activity attributes to attributes table
INSERT IGNORE INTO {{ .Attributes }} (`event_id`, `instance_id`, `execution_id`, `data`) SELECT `activity_id`, `instance_id`, `execution_id`, `attributes` FROM {{ .Activities }};
ALTER TABLE {{ .Activities }} DROP COLUMN `attributes`;

-- Move history attributes to attributes table
INSERT IGNORE INTO {{ .Attributes }} (`event_id`, `instance_id`, `execution_id`, `data`) SELECT `event_id`, `instance_id`, `execution_id`, `attributes` FROM {{ .History }};
ALTER TABLE {{ .History }} DROP COLUMN `attributes`;

-- Move pending_events attributes to attributes table
INSERT IGNORE INTO {{ .Attributes }} (`event_id`, `instance_id`, `execution_id`, `data`) SELECT `event_id`, `instance_id`, `execution_id`, `attributes` FROM {{ .PendingEvents }};
ALTER TABLE {{ .PendingEvents }} DROP COLUMN `attributes`;
