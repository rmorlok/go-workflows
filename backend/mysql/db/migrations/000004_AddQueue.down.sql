-- Remove queue column from instances
ALTER TABLE {{ .Instances }} DROP COLUMN `queue`;

-- Update index
DROP INDEX {{ .IdxInstancesLockedUntilCompletedAtQueue }} ON {{ .Instances }};
CREATE INDEX {{ .IdxInstancesLockedUntilCompletedAt }} ON {{ .Instances }} (`completed_at`, `locked_until`, `sticky_until`, `worker`);

-- Update index
DROP INDEX {{ .IdxActivitiesLockedUntilQueue }} ON {{ .Activities }};

-- Remove queue column from activities
ALTER TABLE {{ .Activities }} DROP COLUMN `queue`;

CREATE INDEX {{ .IdxActivitiesLockedUntil }} ON {{ .Activities }} (`locked_until`);
