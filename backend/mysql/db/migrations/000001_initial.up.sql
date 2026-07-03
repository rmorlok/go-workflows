CREATE TABLE IF NOT EXISTS {{ .Instances }} (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `instance_id` NVARCHAR(128) NOT NULL,
  `execution_id` NVARCHAR(128) NOT NULL,
  `parent_instance_id` NVARCHAR(128) NULL,
  `parent_execution_id` NVARCHAR(128) NULL,
  `parent_schedule_event_id` BIGINT NULL,
  `metadata` BLOB NULL,
  `state` INT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `sticky_until` DATETIME NULL,
  `worker` NVARCHAR(64) NULL,

  UNIQUE INDEX {{ .IdxInstancesInstanceIDExecutionID }} (`instance_id`, `execution_id`),
  INDEX {{ .IdxInstancesLockedUntilCompletedAt }} (`completed_at`, `locked_until`, `sticky_until`, `worker`),
  INDEX {{ .IdxInstancesParentInstanceIDParentExecutionID }} (`parent_instance_id`, `parent_execution_id`)
);


CREATE TABLE IF NOT EXISTS {{ .PendingEvents }} (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `event_id` NVARCHAR(128) NOT NULL,
  `sequence_id` BIGINT NOT NULL, -- Not used, but keep for now for query compat
  `instance_id` NVARCHAR(128) NOT NULL,
  `execution_id` NVARCHAR(128) NOT NULL,
  `event_type` INT NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` BIGINT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,

  INDEX {{ .IdxPendingEventsInIDExID }} (`instance_id`, `execution_id`),
  INDEX {{ .IdxPendingEventsInIDExIDVisibleAtScheduleEventID }} (`instance_id`, `execution_id`, `visible_at`, `schedule_event_id`)
);


CREATE TABLE IF NOT EXISTS {{ .History }} (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `event_id` NVARCHAR(64) NOT NULL,
  `sequence_id` BIGINT NOT NULL,
  `instance_id` NVARCHAR(128) NOT NULL,
  `execution_id` NVARCHAR(128) NOT NULL,
  `event_type` INT NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` BIGINT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL, -- Is this required?

  INDEX {{ .IdxHistoryInstanceIDExecutionID }} (`instance_id`, `execution_id`),
  INDEX {{ .IdxHistoryInstanceIDExecutionIDSequence }} (`instance_id`, `execution_id`, `sequence_id`)
);


CREATE TABLE IF NOT EXISTS {{ .Activities }} (
  `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `activity_id` NVARCHAR(64) NOT NULL,
  `instance_id` NVARCHAR(128) NOT NULL,
  `execution_id` NVARCHAR(128) NOT NULL,
  `event_type` INT NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` BIGINT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `worker` NVARCHAR(64) NULL,

  UNIQUE INDEX {{ .IdxActivitiesInstanceIDExecutionIDActivityIDWorker }} (`instance_id`, `execution_id`, `activity_id`, `worker`),
  INDEX {{ .IdxActivitiesLockedUntil }} (`locked_until`)
);
