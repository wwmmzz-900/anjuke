-- 积分模块数据库表结构

-- 用户积分表
CREATE TABLE IF NOT EXISTS `user_points` (
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `total_points` bigint NOT NULL DEFAULT '0' COMMENT '总积分',
  `available_points` bigint NOT NULL DEFAULT '0' COMMENT '可用积分',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`user_id`),
  KEY `idx_available_points` (`available_points`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户积分表';

-- 积分记录表
CREATE TABLE IF NOT EXISTS `points_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `type` varchar(20) NOT NULL COMMENT '类型:checkin,consume,use',
  `points` bigint NOT NULL COMMENT '积分变动数量',
  `description` varchar(200) DEFAULT NULL COMMENT '描述',
  `order_id` varchar(50) DEFAULT NULL COMMENT '关联订单ID',
  `amount` bigint DEFAULT '0' COMMENT '关联金额(分)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_type` (`type`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分记录表';

-- 签到记录表
CREATE TABLE IF NOT EXISTS `checkin_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `check_date` date NOT NULL COMMENT '签到日期',
  `points` bigint NOT NULL COMMENT '获得积分',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_date` (`user_id`, `check_date`),
  KEY `idx_check_date` (`check_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='签到记录表';

-- 插入一些测试数据
INSERT IGNORE INTO `user_points` (`user_id`, `total_points`, `available_points`) VALUES
(1, 100, 80),
(2, 50, 50),
(3, 200, 150);

INSERT IGNORE INTO `points_records` (`user_id`, `type`, `points`, `description`, `order_id`, `amount`) VALUES
(1, 'checkin', 10, '签到获得积分（连续1天）', NULL, 0),
(1, 'consume', 20, '消费获得积分（订单金额：20.00元）', 'ORDER001', 2000),
(1, 'use', -30, '积分抵扣（抵扣金额：3.00元）', 'ORDER002', 300),
(2, 'checkin', 10, '签到获得积分（连续1天）', NULL, 0),
(2, 'consume', 40, '消费获得积分（订单金额：40.00元）', 'ORDER003', 4000),
(3, 'checkin', 50, '签到获得积分（连续7天）', NULL, 0),
(3, 'consume', 150, '消费获得积分（订单金额：150.00元）', 'ORDER004', 15000);

INSERT IGNORE INTO `checkin_records` (`user_id`, `check_date`, `points`) VALUES
(1, CURDATE(), 10),
(2, CURDATE(), 10),
(3, CURDATE(), 15);