DROP TABLE IF EXISTS {{ .Instances }};

CREATE TABLE {{ .Instances }} (
  id bigserial NOT NULL PRIMARY KEY,
  instance_id varchar(128) NOT NULL,
  execution_id varchar(128) NOT NULL,
  parent_instance_id varchar(128) NULL,
  parent_execution_id varchar(128) NULL,
  parent_schedule_event_id numeric NULL,
  metadata bytea NULL,
  state int NOT NULL,
  queue varchar(128) DEFAULT '' NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at timestamptz NULL,
  locked_until timestamptz NULL,
  sticky_until timestamptz NULL,
  worker varchar(64) NULL
);

CREATE UNIQUE INDEX {{ .IdxInstancesInstanceIDExecutionID }} on {{ .Instances }} (instance_id, execution_id);
CREATE INDEX {{ .IdxInstancesLockedUntilCompletedAtQueue }} ON {{ .Instances }} (completed_at, locked_until, sticky_until, worker, queue);
CREATE INDEX {{ .IdxInstancesParentInstanceIDParentExecutionID }} ON {{ .Instances }} (parent_instance_id, parent_execution_id);

DROP TABLE IF EXISTS {{ .PendingEvents }};
CREATE TABLE {{ .PendingEvents }} (
  id bigserial NOT NULL PRIMARY KEY,
  event_id varchar(128) NOT NULL,
  sequence_id bigserial NOT NULL, -- Not used, but keep for now for query compat
  instance_id varchar(128) NOT NULL,
  execution_id varchar(128) NOT NULL,
  event_type int NOT NULL,
  timestamp timestamptz NOT NULL,
  schedule_event_id bigserial NOT NULL,
  visible_at timestamptz NULL
);

CREATE INDEX {{ .IdxPendingEventsInIDExID }} ON {{ .PendingEvents }} (instance_id, execution_id);
CREATE INDEX {{ .IdxPendingEventsInIDExIDVisibleAtScheduleEventID }} ON {{ .PendingEvents }} (instance_id, execution_id, visible_at, schedule_event_id);

DROP TABLE IF EXISTS {{ .History }};
CREATE TABLE IF NOT EXISTS {{ .History }} (
  id bigserial NOT NULL PRIMARY KEY,
  event_id varchar(128) NOT NULL,
  sequence_id bigserial NOT NULL,
  instance_id varchar(128) NOT NULL,
  execution_id varchar(128) NOT NULL,
  event_type int NOT NULL,
  timestamp timestamptz NOT NULL,
  schedule_event_id bigserial NOT NULL,
  visible_at timestamptz NULL
);

CREATE INDEX {{ .IdxHistoryInstanceIDExecutionID }} ON {{ .History }} (instance_id, execution_id);
CREATE INDEX {{ .IdxHistoryInstanceIDExecutionIDSequence }} ON {{ .History }} (instance_id, execution_id, sequence_id);

DROP TABLE IF EXISTS {{ .Activities }};
CREATE TABLE IF NOT EXISTS {{ .Activities }} (
  id bigserial NOT NULL PRIMARY KEY,
  activity_id varchar(128) NOT NULL,
  instance_id varchar(128) NOT NULL,
  execution_id varchar(128) NOT NULL,
  event_type int NOT NULL,
  queue varchar(128) DEFAULT '' NOT NULL,
  timestamp timestamptz NOT NULL,
  schedule_event_id bigserial NOT NULL,
  visible_at timestamptz NULL,
  locked_until timestamptz NULL,
  worker VARCHAR(64) NULL
);

CREATE UNIQUE INDEX {{ .IdxActivitiesInstanceIDExecutionIDActivityIDWorker }} ON {{ .Activities }} (instance_id, execution_id, activity_id, worker);
CREATE INDEX {{ .IdxActivitiesLockedUntilQueue }} ON {{ .Activities }} (locked_until, queue);

DROP TABLE IF EXISTS {{ .Attributes }};
CREATE TABLE {{ .Attributes }} (
  id BIGSERIAL NOT NULL PRIMARY KEY,
  event_id varchar(128) NOT NULL,
  instance_id varchar(128) NOT NULL,
  execution_id varchar(128) NOT NULL,
  data bytea NOT NULL
);

CREATE UNIQUE INDEX {{ .IdxAttributesInstanceIDExecutionIDEventID }} on {{ .Attributes }} (instance_id, execution_id, event_id);
CREATE INDEX {{ .IdxAttributesEventID }} on {{ .Attributes }} (event_id);
