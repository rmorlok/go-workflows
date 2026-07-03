-- Add queue to instances
ALTER TABLE {{ .Instances }} ADD COLUMN `queue` NVARCHAR(128) DEFAULT '';

-- Update index
DROP INDEX IF EXISTS {{ .IdxInstancesLockedUntilCompletedAt }} ;
CREATE INDEX {{ .IdxInstancesLockedUntilCompletedAtQueue }} ON {{ .Instances }} (`completed_at`, `locked_until`, `sticky_until`, `worker`, `queue`);

-- Add queue to activities
ALTER TABLE {{ .Activities }} ADD COLUMN `queue` NVARCHAR(128) DEFAULT '';

-- Update index
DROP INDEX IF EXISTS {{ .IdxActivitiesInstanceIDExecutionIDWorker }};
CREATE INDEX {{ .IdxActivitiesInstanceIDExecutionIDWorkerQueue }} ON {{ .Activities }} (`instance_id`, `execution_id`, `worker`, `queue`);

DROP INDEX IF EXISTS {{ .IdxActivitiesLockedUntil }};

CREATE INDEX {{ .IdxActivitiesLockedUntilQueue }} ON {{ .Activities }} (`locked_until`, `queue`);
