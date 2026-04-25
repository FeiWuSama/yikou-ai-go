-- 创建 Nacos 数据库
CREATE DATABASE IF NOT EXISTS nacos DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建应用数据库
CREATE DATABASE IF NOT EXISTS yikou_ai DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用应用数据库
USE yikou_ai;

-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `user_account` varchar(256) NOT NULL COMMENT '用户账号',
    `user_password` varchar(512) NOT NULL COMMENT '用户密码',
    `user_name` varchar(256) DEFAULT NULL COMMENT '用户昵称',
    `user_avatar` varchar(1024) DEFAULT NULL COMMENT '用户头像',
    `user_profile` varchar(512) DEFAULT NULL COMMENT '用户简介',
    `user_role` varchar(256) NOT NULL DEFAULT 'user' COMMENT '用户角色',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_delete` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_account` (`user_account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 应用表
CREATE TABLE IF NOT EXISTS `app` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '应用ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `app_name` varchar(256) NOT NULL COMMENT '应用名称',
    `app_description` varchar(512) DEFAULT NULL COMMENT '应用描述',
    `app_type` varchar(64) NOT NULL COMMENT '应用类型',
    `app_status` int NOT NULL DEFAULT 0 COMMENT '应用状态',
    `app_code` longtext COMMENT '应用代码',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_delete` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用表';

-- 聊天历史表
CREATE TABLE IF NOT EXISTS `chat_history` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '聊天历史ID',
    `app_id` bigint NOT NULL COMMENT '应用ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `message_type` varchar(64) NOT NULL COMMENT '消息类型',
    `message_content` text NOT NULL COMMENT '消息内容',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_delete` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    KEY `idx_app_id` (`app_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天历史表';
