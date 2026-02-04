-- Active: 1769663211548@@127.0.0.1@3307@payment_db
-- 支付服务 Outbox 表，用于实现 Outbox 模式的可靠事件投递
USE payment_db;

CREATE TABLE IF NOT EXISTS payment_outbox (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    event_type   VARCHAR(64)     NOT NULL COMMENT '事件类型，如 payment.succeeded',
    aggregate_id VARCHAR(64)     NOT NULL COMMENT '聚合ID，如 payment_no 或 order_no',
    payload      JSON            NOT NULL COMMENT '事件载荷，JSON 格式',
    status       TINYINT         NOT NULL DEFAULT 0 COMMENT '状态：0-待发送，1-已发送，2-发送失败',
    retry_count  INT             NOT NULL DEFAULT 0 COMMENT '重试次数',
    error_msg    VARCHAR(500)             DEFAULT NULL COMMENT '最近一次错误信息',
    created_at   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_status_created_at (status, created_at),
    INDEX idx_aggregate_id (aggregate_id),
    INDEX idx_event_type (event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付服务 Outbox 表';

