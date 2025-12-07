-- Migration: Add host_id column to v2_server table
-- Date: 2024-12-07
-- Description: Add host_id field to enable binding servers to hosts for auto-deployment

-- Add host_id column to v2_server table
ALTER TABLE `v2_server` 
ADD COLUMN `host_id` BIGINT NULL DEFAULT NULL COMMENT '绑定的主机ID' AFTER `parent_id`;

-- Add index for host_id
ALTER TABLE `v2_server` 
ADD INDEX `idx_server_host_id` (`host_id`);

-- Add foreign key constraint (optional, uncomment if you want strict referential integrity)
-- ALTER TABLE `v2_server` 
-- ADD CONSTRAINT `fk_server_host_id` 
-- FOREIGN KEY (`host_id`) REFERENCES `v2_host` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;
