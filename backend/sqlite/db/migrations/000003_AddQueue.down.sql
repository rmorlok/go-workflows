-- Drop column from instances and activities
ALTER TABLE {{ .Instances }} DROP COLUMN `queue`;
ALTER TABLE {{ .Activities }} DROP COLUMN `queue`;

-- Update index
DROP INDEX IF EXISTS {{ .IdxInstancesLockedUntilCompletedAtQueue }};
CREATE INDEX {{ .IdxInstancesLockedUntilCompletedAt }} ON {{ .Instances }} (`completed_at`, `locked_until`, `sticky_until`, `worker`);

DROP INDEX IF EXISTS {{ .IdxActivitiesInstanceIDExecutionIDWorkerQueue }};
CREATE INDEX {{ .IdxActivitiesInstanceIDExecutionIDWorker }} ON {{ .Activities }} (`instance_id`, `execution_id`, `id`, `worker`);
