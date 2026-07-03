ALTER TABLE {{ .Activities }} ADD COLUMN `attributes` BLOB NULL;
UPDATE {{ .Activities }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .Activities }}.`id` = {{ .Attributes }}.`id` AND {{ .Activities }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .Activities }}.`execution_id` = {{ .Attributes }}.`execution_id`;

ALTER TABLE {{ .History }} ADD COLUMN `attributes` BLOB NULL;
UPDATE {{ .History }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .History }}.`id` = {{ .Attributes }}.`id` AND {{ .History }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .History }}.`execution_id` = {{ .Attributes }}.`execution_id`;

ALTER TABLE {{ .PendingEvents }} ADD COLUMN `attributes` BLOB NULL;
UPDATE {{ .PendingEvents }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .PendingEvents }}.`id` = {{ .Attributes }}.`id` AND {{ .PendingEvents }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .PendingEvents }}.`execution_id` = {{ .Attributes }}.`execution_id`;

-- Drop attributes table
DROP TABLE {{ .Attributes }};
