ALTER TABLE {{ .Activities }} MODIFY COLUMN `attributes` BLOB NOT NULL;
ALTER TABLE {{ .History }} MODIFY COLUMN `attributes` BLOB NOT NULL;
ALTER TABLE {{ .PendingEvents }} MODIFY COLUMN `attributes` BLOB NOT NULL;
