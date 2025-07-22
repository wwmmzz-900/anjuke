-- 使用数据库
USE anjuke;

-- 创建聊天会话表
CREATE TABLE IF NOT EXISTS chat_sessions (
    chat_id VARCHAR(64) PRIMARY KEY COMMENT '聊天ID',
    reservation_id BIGINT NOT NULL COMMENT '预约ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    landlord_id BIGINT NOT NULL COMMENT '房东ID',
    house_id BIGINT NOT NULL COMMENT '房源ID',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态：active/closed',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    
    INDEX idx_reservation_id (reservation_id),
    INDEX idx_user_id (user_id),
    INDEX idx_landlord_id (landlord_id),
    INDEX idx_house_id (house_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';

-- 创建聊天消息表
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '消息ID',
    chat_id VARCHAR(64) NOT NULL COMMENT '聊天ID',
    sender_id BIGINT NOT NULL COMMENT '发送者ID',
    sender_name VARCHAR(100) NOT NULL COMMENT '发送者名称',
    receiver_id BIGINT NOT NULL COMMENT '接收者ID',
    receiver_name VARCHAR(100) NOT NULL COMMENT '接收者名称',
    type INT NOT NULL DEFAULT 0 COMMENT '消息类型：0-文本，1-图片，2-语音，3-位置，4-系统消息',
    content TEXT NOT NULL COMMENT '消息内容',
    read BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否已读',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    
    INDEX idx_chat_id (chat_id),
    INDEX idx_sender_id (sender_id),
    INDEX idx_receiver_id (receiver_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';

-- 清空现有测试数据（可选）
DELETE FROM houses WHERE house_id IN (101, 102, 103);
DELETE FROM users WHERE id IN (1001, 1002, 1003, 2001, 2002, 2003);
DELETE FROM user_behavior WHERE user_id IN (1001, 1002, 1003);
DELETE FROM reservations WHERE user_id IN (1001, 1002, 1003);
DELETE FROM chat_sessions;
DELETE FROM chat_messages;

-- 插入测试房源数据
INSERT INTO houses (house_id, landlord_id, title, description, price, area, layout, address, image_url, status, created_at, updated_at) VALUES
(101, 2001, '精装修两室一厅', '地铁口附近，交通便利，精装修', 3500.00, 85.50, '2室1厅1卫', '北京市朝阳区xxx小区', 'https://example.com/house1.jpg', 'available', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(102, 2002, '温馨三室两厅', '小区环境优美，配套设施完善', 4200.00, 120.00, '3室2厅2卫', '北京市海淀区xxx小区', 'https://example.com/house2.jpg', 'available', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(103, 2003, '豪华公寓', '高端小区，装修豪华，设施齐全', 5800.00, 150.00, '3室2厅2卫', '北京市西城区xxx小区', 'https://example.com/house3.jpg', 'available', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入测试用户数据
INSERT INTO users (id, username, nickname, phone, email, user_type, status, created_at, updated_at) VALUES
(1001, 'zhangsan', '张三', '13800138001', 'zhangsan@example.com', 'tenant', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(1002, 'lisi', '李四', '13800138002', 'lisi@example.com', 'tenant', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(1003, 'wangwu', '王五', '13800138003', 'wangwu@example.com', 'tenant', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2001, 'landlord1', '房东一', '13900139001', 'landlord1@example.com', 'landlord', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2002, 'landlord2', '房东二', '13900139002', 'landlord2@example.com', 'landlord', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2003, 'landlord3', '房东三', '13900139003', 'landlord3@example.com', 'landlord', 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入测试用户行为数据
INSERT INTO user_behavior (user_id, house_id, behavior, created_at) VALUES
(1001, 101, 'view', NOW()),
(1001, 102, 'view', NOW()),
(1001, 103, 'like', NOW()),
(1002, 101, 'view', NOW()),
(1002, 103, 'contact', NOW()),
(1003, 102, 'view', NOW());

-- 插入测试预约数据
INSERT INTO reservations (landlord_id, user_id, user_name, house_id, house_title, reserve_time, status, created_at, updated_at) VALUES
(2001, 1001, '张三', 101, '精装修两室一厅', '2025-07-22 14:00:00', 'pending', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2002, 1002, '李四', 102, '温馨三室两厅', '2025-07-22 15:30:00', 'confirmed', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2003, 1003, '王五', 103, '豪华公寓', '2025-07-22 16:00:00', 'cancelled', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());