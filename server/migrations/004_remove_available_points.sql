-- 移除可用积分字段，简化积分系统
-- 只保留总积分，消费积分直接从总积分中扣除

-- 删除可用积分字段
ALTER TABLE `user_points` DROP COLUMN `available_points`;

-- 删除可用积分相关的索引（如果存在）
-- 注意：MySQL 5.7及以下版本不支持 IF EXISTS，如果索引不存在会报错，但不影响功能
-- DROP INDEX `idx_available_points` ON `user_points`;

-- 更新表注释
ALTER TABLE `user_points` COMMENT='用户积分表（简化版，只保留总积分）';