-- 监控工具管理系统数据库初始化脚本
-- 创建时间：2026-03-08
-- 数据库类型：MySQL

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS monitor_tool_commercial DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE monitor_tool_commercial;

-- ============================================
-- 1. 用户表 (user)
-- ============================================
DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名',
  `phone` VARCHAR(20) NOT NULL COMMENT '手机号',
  `password` VARCHAR(100) NOT NULL COMMENT '密码',
  `avatar` VARCHAR(255) DEFAULT NULL COMMENT '头像地址',
  `member_level` BIGINT DEFAULT 1 COMMENT '会员等级',
  `member_end_at` DATETIME DEFAULT NULL COMMENT '会员结束时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 2. 监控项表 (monitor)
-- ============================================
DROP TABLE IF EXISTS `monitor`;

CREATE TABLE `monitor` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '监控项ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `name` VARCHAR(100) NOT NULL COMMENT '监控项名称',
  `url` VARCHAR(255) NOT NULL COMMENT '监控URL',
  `monitor_type` INT DEFAULT 1 COMMENT '监控类型（1-HTTP/HTTPS）',
  `frequency` INT DEFAULT 60 COMMENT '监控频率（秒）',
  `status` INT DEFAULT 1 COMMENT '状态（0-初始化，1-正常，2-宕机，3-暂停）',
  `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
  `create_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `error_msg` VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
  `last_status` INT DEFAULT 0 COMMENT '上一次状态',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控项表';

-- ============================================
-- 3. 监控历史表 (monitor_history)
-- ============================================
DROP TABLE IF EXISTS `monitor_history`;

CREATE TABLE `monitor_history` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
  `monitor_id` BIGINT NOT NULL COMMENT '监控项ID',
  `status` INT NOT NULL COMMENT '状态',
  `response_time` INT NOT NULL COMMENT '响应时间（毫秒）',
  `error_msg` VARCHAR(500) DEFAULT NULL COMMENT '错误信息',
  `monitor_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '监控时间',
  PRIMARY KEY (`id`),
  KEY `idx_monitor_id` (`monitor_id`),
  KEY `idx_monitor_time` (`monitor_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控历史表';

-- ============================================
-- 4. 告警配置表 (alert_config)
-- ============================================
DROP TABLE IF EXISTS `alert_config`;

CREATE TABLE `alert_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `email` VARCHAR(100) NOT NULL COMMENT '告警邮箱',
  `alert_type` INT DEFAULT 1 COMMENT '默认告警方式（1-邮箱，2-钉钉）',
  `is_enabled` INT DEFAULT 1 COMMENT '告警开关（0-关闭，1-开启）',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警配置表';

-- ============================================
-- 5. 告警记录表 (alert)
-- ============================================
DROP TABLE IF EXISTS `alert`;

CREATE TABLE `alert` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '告警ID',
  `monitor_id` BIGINT NOT NULL COMMENT '监控项ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `alert_type` INT DEFAULT 1 COMMENT '告警类型（1-邮箱，2-钉钉）',
  `alert_sub_type` INT DEFAULT 1 COMMENT '告警子类型（1-宕机告警，2-恢复通知）',
  `status` INT DEFAULT 0 COMMENT '告警状态（0-未发送，1-已发送，2-发送失败）',
  `content` VARCHAR(500) NOT NULL COMMENT '告警内容',
  `send_time` DATETIME DEFAULT NULL COMMENT '发送时间',
  `create_time` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_monitor_id` (`monitor_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警记录表';

-- ============================================
-- 初始化测试数据
-- ============================================

-- 插入测试用户（密码：12345678）
INSERT INTO `user` (`username`, `phone`, `password`, `member_level`) VALUES
('testuser', '13800138000', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 1);

-- ============================================
-- 添加外键约束（在数据插入完成后添加）
-- ============================================

ALTER TABLE `monitor` ADD CONSTRAINT `fk_monitor_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE;

ALTER TABLE `monitor_history` ADD CONSTRAINT `fk_monitor_history_monitor_id` FOREIGN KEY (`monitor_id`) REFERENCES `monitor`(`id`) ON DELETE CASCADE;

ALTER TABLE `alert` ADD CONSTRAINT `fk_alert_monitor_id` FOREIGN KEY (`monitor_id`) REFERENCES `monitor`(`id`) ON DELETE CASCADE;

ALTER TABLE `alert` ADD CONSTRAINT `fk_alert_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE;

ALTER TABLE `alert_config` ADD CONSTRAINT `fk_alert_config_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE CASCADE;

-- ============================================
-- 创建索引优化查询性能
-- ============================================

-- 监控项表索引
CREATE INDEX `idx_monitor_user_status` ON `monitor`(`user_id`, `status`);

-- 告警记录表索引
CREATE INDEX `idx_alert_user_status_time` ON `alert`(`user_id`, `status`, `create_time`);

-- 监控历史表索引
CREATE INDEX `idx_monitor_history_monitor_time` ON `monitor_history`(`monitor_id`, `monitor_time`);
