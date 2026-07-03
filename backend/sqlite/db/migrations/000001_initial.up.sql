CREATE TABLE IF NOT EXISTS {{ .Instances }} (
  `id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `parent_instance_id` TEXT NULL,
  `parent_execution_id` TEXT NULL,
  `parent_schedule_event_id` INTEGER NULL,
  `metadata` TEXT NULL,
  `state` INTEGER NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `sticky_until` DATETIME NULL,
  `worker` TEXT NULL,
  PRIMARY KEY(`id`, `execution_id`)
);

CREATE INDEX IF NOT EXISTS {{ .IdxInstancesIDExecutionID }} ON {{ .Instances }} (`id`, `execution_id`);
CREATE INDEX IF NOT EXISTS {{ .IdxInstancesLockedUntilCompletedAt }} ON {{ .Instances }} (`locked_until`, `sticky_until`, `completed_at`, `worker`);
CREATE INDEX IF NOT EXISTS {{ .IdxInstancesParentInstanceIDParentExecutionID }} ON {{ .Instances }} (`parent_instance_id`, `parent_execution_id`);

CREATE TABLE IF NOT EXISTS {{ .PendingEvents }} (
  `id` TEXT,
  `sequence_id` INTEGER NOT NULL, -- not used but keep for now for query compat
  `instance_id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,
  PRIMARY KEY(`id`, `instance_id`)
);

CREATE INDEX IF NOT EXISTS {{ .IdxPendingEventsInstanceIDExecutionIDVisibleAtScheduleEventID }} ON {{ .PendingEvents }} (`instance_id`, `execution_id`, `visible_at`, `schedule_event_id`);

CREATE TABLE IF NOT EXISTS {{ .History }} (
  `id` TEXT,
  `sequence_id` INTEGER NOT NULL,
  `instance_id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,
  PRIMARY KEY(`id`, `instance_id`)
);

CREATE INDEX IF NOT EXISTS {{ .IdxHistoryInstanceSequence }} ON {{ .History }} (`instance_id`, `execution_id`, `sequence_id`);

CREATE TABLE IF NOT EXISTS {{ .Activities }} (
  `id` TEXT PRIMARY KEY,
  `instance_id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `worker` TEXT NULL
);


CREATE INDEX IF NOT EXISTS {{ .IdxActivitiesIDWorker }} ON {{ .Activities }} (`id`, `worker`);
CREATE INDEX IF NOT EXISTS {{ .IdxActivitiesLockedUntil }} ON {{ .Activities }} (`locked_until`);
CREATE INDEX IF NOT EXISTS {{ .IdxActivitiesInstanceIDExecutionIDWorker }} ON {{ .Activities }} (`instance_id`, `execution_id`, `worker`);
