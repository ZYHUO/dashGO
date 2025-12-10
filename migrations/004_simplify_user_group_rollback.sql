-- 回滚：恢复用户组的默认值字段（如果之前删除了）

-- 如果之前删除了这些字段，可以重新添加：
-- ALTER TABLE v2_user_group ADD COLUMN IF NOT EXISTS default_transfer_enable BIGINT DEFAULT 0 COMMENT '默认流量（字节）';
-- ALTER TABLE v2_user_group ADD COLUMN IF NOT EXISTS default_speed_limit INT COMMENT '默认速度限制（Mbps）';
-- ALTER TABLE v2_user_group ADD COLUMN IF NOT EXISTS default_device_limit INT COMMENT '默认设备数限制';

-- 恢复默认组的原始说明
UPDATE v2_user_group SET description = '新注册用户默认组，流量较少' WHERE id = 1;
UPDATE v2_user_group SET description = '购买基础套餐的用户' WHERE id = 2;
UPDATE v2_user_group SET description = '购买高级套餐的用户' WHERE id = 3;
