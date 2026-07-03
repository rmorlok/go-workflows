ALTER TABLE {{ .Instances }} ADD COLUMN `queue` NVARCHAR(128) DEFAULT '';

-- Update index
DROP INDEX {{ .IdxInstancesLockedUntilCompletedAt }} ON {{ .Instances }};
CREATE INDEX {{ .IdxInstancesLockedUntilCompletedAtQueue }} ON {{ .Instances }} (`completed_at`, `locked_until`, `sticky_until`, `worker`, `queue`);

ALTER TABLE {{ .Activities }} ADD COLUMN `queue` NVARCHAR(128) DEFAULT '';

-- Update index
DROP INDEX {{ .IdxActivitiesLockedUntil }} ON {{ .Activities }};
CREATE INDEX {{ .IdxActivitiesLockedUntilQueue }} ON {{ .Activities }} (`locked_until`, `queue`);
