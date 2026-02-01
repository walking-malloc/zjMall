-- Active: 1769663211548@@127.0.0.1@3307@mysql
-- 支付服务数据库初始化脚本

-- 创建数据库（如需与其他服务隔离，可单独一个库）
CREATE DATABASE IF NOT EXISTS payment_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE payment_db;

-- ============================================
-- 1. 支付单表
-- 对应 Go 模型：internal/payment-service/model/payment.go
-- ============================================
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '主键ID（ULID）',
    payment_no VARCHAR(32) NOT NULL UNIQUE COMMENT '支付单号',
    order_no VARCHAR(32) NOT NULL COMMENT '订单号',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',

    amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '支付金额',

    pay_channel VARCHAR(20) NOT NULL COMMENT '支付渠道：wechat-微信支付，alipay-支付宝，balance-余额支付',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '支付状态：1-待支付，2-支付中，3-支付成功，4-支付失败，5-已关闭，6-已退款',

    trade_no VARCHAR(64) COMMENT '第三方交易号',
    notify_url VARCHAR(255) COMMENT '回调地址',
    return_url VARCHAR(255) COMMENT '返回地址',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    paid_at TIMESTAMP NULL DEFAULT NULL COMMENT '支付时间',
    expired_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
    version INT NOT NULL DEFAULT 0 COMMENT '版本号（乐观锁）',

    INDEX idx_order_no (order_no),
    INDEX idx_user_id (user_id),
    INDEX idx_trade_no (trade_no),
    INDEX idx_status (status),
    INDEX idx_expired_at (expired_at),
    INDEX idx_status_expired (status, expired_at) COMMENT '用于定时任务扫描超时支付单',
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付单表';


-- ============================================
-- 2. 支付日志表
-- 用于记录支付操作的详细日志，便于审计与排查问题
-- 对应场景：创建支付单、支付回调、状态变更等
-- ============================================
CREATE TABLE IF NOT EXISTS payment_logs (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '日志ID（ULID）',
    payment_no VARCHAR(32) NOT NULL COMMENT '支付单号',
    order_no VARCHAR(32) NOT NULL COMMENT '订单号',
    user_id VARCHAR(26) COMMENT '用户ID',
    
    action VARCHAR(50) NOT NULL COMMENT '操作类型：create-创建支付单，callback-支付回调，status_change-状态变更，close-关闭支付单',
    from_status TINYINT COMMENT '变更前状态',
    to_status TINYINT COMMENT '变更后状态',
    
    channel VARCHAR(20) COMMENT '支付渠道',
    amount DECIMAL(10, 2) COMMENT '支付金额',
    trade_no VARCHAR(64) COMMENT '第三方交易号',
    
    request_data TEXT COMMENT '请求数据（JSON格式，用于记录回调参数等）',
    response_data TEXT COMMENT '响应数据（JSON格式）',
    error_message VARCHAR(500) COMMENT '错误信息（如有）',
    
    ip_address VARCHAR(50) COMMENT 'IP地址',
    user_agent VARCHAR(255) COMMENT '用户代理',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_payment_no (payment_no),
    INDEX idx_order_no (order_no),
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at),
    INDEX idx_payment_action (payment_no, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付日志表';


-- ============================================
-- 3. 退款单表
-- 用于记录退款信息，与支付单关联
-- 对应场景：订单退款、部分退款、售后退款等
-- ============================================
CREATE TABLE IF NOT EXISTS refunds (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '退款单ID（ULID）',
    refund_no VARCHAR(32) NOT NULL UNIQUE COMMENT '退款单号',
    payment_no VARCHAR(32) NOT NULL COMMENT '原支付单号',
    order_no VARCHAR(32) NOT NULL COMMENT '订单号',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
    
    refund_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '退款金额',
    refund_reason VARCHAR(255) COMMENT '退款原因',
    refund_type TINYINT NOT NULL DEFAULT 1 COMMENT '退款类型：1-全额退款，2-部分退款',
    
    pay_channel VARCHAR(20) NOT NULL COMMENT '原支付渠道：wechat-微信支付，alipay-支付宝',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '退款状态：1-退款中，2-退款成功，3-退款失败，4-已取消',
    
    trade_no VARCHAR(64) COMMENT '原支付交易号',
    refund_trade_no VARCHAR(64) COMMENT '退款交易号（第三方返回）',
    
    request_data TEXT COMMENT '退款请求数据（JSON格式）',
    response_data TEXT COMMENT '退款响应数据（JSON格式）',
    error_message VARCHAR(500) COMMENT '错误信息（如有）',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    refunded_at TIMESTAMP NULL DEFAULT NULL COMMENT '退款成功时间',
    version INT NOT NULL DEFAULT 0 COMMENT '版本号（乐观锁）',
    
    INDEX idx_payment_no (payment_no),
    INDEX idx_order_no (order_no),
    INDEX idx_user_id (user_id),
    INDEX idx_refund_trade_no (refund_trade_no),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款单表';


-- ============================================
-- 4. 支付渠道配置表
-- 用于管理支付渠道的配置信息（商户号、密钥等）
-- 支持多商户、多应用场景
-- ============================================
CREATE TABLE IF NOT EXISTS payment_channels (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '渠道配置ID（ULID）',
    channel_code VARCHAR(20) NOT NULL COMMENT '渠道代码：wechat-微信支付，alipay-支付宝',
    channel_name VARCHAR(50) NOT NULL COMMENT '渠道名称',
    
    app_id VARCHAR(64) COMMENT '应用ID（微信AppID/支付宝AppID）',
    merchant_id VARCHAR(64) COMMENT '商户号',
    mch_id VARCHAR(64) COMMENT '微信商户号（微信专用）',
    
    api_key VARCHAR(255) COMMENT 'API密钥（加密存储）',
    public_key TEXT COMMENT '公钥（支付宝专用）',
    private_key TEXT COMMENT '私钥（支付宝专用，加密存储）',
    
    notify_url VARCHAR(255) COMMENT '回调地址',
    return_url VARCHAR(255) COMMENT '返回地址',
    
    is_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用：0-禁用，1-启用',
    is_default TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认渠道：0-否，1-是',
    
    environment VARCHAR(20) NOT NULL DEFAULT 'sandbox' COMMENT '环境：sandbox-沙箱，production-生产',
    
    remark VARCHAR(255) COMMENT '备注',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_channel_code (channel_code),
    INDEX idx_is_enabled (is_enabled),
    INDEX idx_environment (environment),
    UNIQUE KEY uk_channel_env (channel_code, environment)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付渠道配置表';

