ALTER TABLE {{ .Activities }} ADD COLUMN `attributes` MEDIUMBLOB NULL;
UPDATE {{ .Activities }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .Activities }}.`event_id` = {{ .Attributes }}.`event_id` AND {{ .Activities }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .Activities }}.`execution_id` = {{ .Attributes }}.`execution_id`;
ALTER TABLE {{ .Activities }} MODIFY COLUMN `attributes` MEDIUMBLOB NOT NULL;

ALTER TABLE {{ .History }} ADD COLUMN `attributes` MEDIUMBLOB NULL;
UPDATE {{ .History }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .History }}.`event_id` = {{ .Attributes }}.`event_id` AND {{ .History }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .History }}.`execution_id` = {{ .Attributes }}.`execution_id`;
ALTER TABLE {{ .History }} MODIFY COLUMN `attributes` MEDIUMBLOB NOT NULL;

ALTER TABLE {{ .PendingEvents }} ADD COLUMN `attributes` MEDIUMBLOB NULL;
UPDATE {{ .PendingEvents }} SET `attributes` = {{ .Attributes }}.`data` FROM {{ .Attributes }} WHERE {{ .PendingEvents }}.`event_id` = {{ .Attributes }}.`event_id` AND {{ .PendingEvents }}.`instance_id` = {{ .Attributes }}.`instance_id` AND {{ .PendingEvents }}.`execution_id` = {{ .Attributes }}.`execution_id`;
ALTER TABLE {{ .PendingEvents }} MODIFY COLUMN `attributes` MEDIUMBLOB NOT NULL;


-- Drop attributes table
DROP TABLE {{ .Attributes }};
