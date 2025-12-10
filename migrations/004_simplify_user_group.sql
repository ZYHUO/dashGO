-- 简化用户组表：移除不必要的默认值字段
-- 注意：流量、速度、设备限制应该由套餐（Plan）决定，不应该在用户组中设置

-- 如果你想完全移除这些字段，可以执行以下语句：
-- ALTER TABLE v2_user_group DROP COLUMN IF EXISTS default_transfer_enable;
-- ALTER TABLE v2_user_group DROP COLUMN IF EXISTS default_speed_limit;
-- ALTER TABLE v2_user_group DROP COLUMN IF EXISTS default_device_limit;

-- 但为了向后兼容，我们保留这些字段，只是不再使用它们
-- 如果数据库中已经有这些字段的数据，它们不会影响系统运行

-- 更新默认组的说明
UPDATE v2_user_group SET description = '新注册用户默认组' WHERE id = 1;
UPDATE v2_user_group SET description = '购买基础套餐的用户' WHERE id = 2;
UPDATE v2_user_group SET description = '购买高级套餐的用户' WHERE id = 3;
