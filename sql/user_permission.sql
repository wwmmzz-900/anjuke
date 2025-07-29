-- 用户权限表
CREATE TABLE `user_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permissions` text COMMENT '权限列表(JSON格式)',
  `operator_id` bigint DEFAULT NULL COMMENT '操作者ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户权限表';

-- 插入基础权限数据
INSERT INTO `permission` (`permission_id`, `name`, `description`) VALUES
(1, 'READ', '读取权限'),
(2, 'WRITE', '写入权限'),
(3, 'DELETE', '删除权限'),
(4, 'ADMIN', '管理员权限'),
(5, 'PUBLISH_HOUSE', '发布房源权限'),
(6, 'MANAGE_USER', '用户管理权限'),
(7, 'MANAGE_TRANSACTION', '交易管理权限'),
(8, 'CUSTOMER_SERVICE', '客服权限');

-- 插入基础角色数据
INSERT INTO `role` (`role_id`, `name`, `description`) VALUES
(0, 'ROLE_GUEST', '游客用户，只能浏览基本信息'),
(1, 'ROLE_NORMAL_USER', '普通用户，可以浏览和基本操作'),
(2, 'ROLE_VIP_USER', 'VIP用户，享有更多特权'),
(3, 'ROLE_LANDLORD', '房东用户，可以发布和管理房源'),
(4, 'ROLE_AGENT', '经纪人，可以管理房源和用户'),
(5, 'ROLE_ADMIN', '管理员，拥有大部分管理权限'),
(6, 'ROLE_SUPER_ADMIN', '超级管理员，拥有所有权限');

-- 插入角色权限关联数据
-- 游客权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES (0, 1);

-- 普通用户权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(1, 1), (1, 2);

-- VIP用户权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(2, 1), (2, 2), (2, 5);

-- 房东权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(3, 1), (3, 2), (3, 5);

-- 经纪人权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(4, 1), (4, 2), (4, 5), (4, 6), (4, 8);

-- 管理员权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(5, 1), (5, 2), (5, 3), (5, 5), (5, 6), (5, 7), (5, 8);

-- 超级管理员权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES 
(6, 1), (6, 2), (6, 3), (6, 4), (6, 5), (6, 6), (6, 7), (6, 8);