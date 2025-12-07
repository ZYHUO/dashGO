-- Rollback Migration: Remove host_id column from v2_server table
-- Date: 2024-12-07

-- Remove foreign key constraint (if it was added)
-- ALTER TABLE `v2_server` DROP FOREIGN KEY `fk_server_host_id`;

-- Remove index
ALTER TABLE `v2_server` DROP INDEX `idx_server_host_id`;

-- Remove host_id column
ALTER TABLE `v2_server` DROP COLUMN `host_id`;
