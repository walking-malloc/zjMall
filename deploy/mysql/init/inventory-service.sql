-- Active: 1769663211548@@127.0.0.1@3307@inventory_db

-- ============================================
-- 库存服务数据库：inventory_db
-- ============================================
CREATE DATABASE IF NOT EXISTS inventory_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE inventory_db;

-- 1. 库存主表（按 SKU 维度）
-- 对应 Go 模型：internal/inventory-service/model/stock.go
-- ============================================
CREATE TABLE IF NOT EXISTS inventory_stocks (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '主键ID',
    sku_id VARCHAR(26) NOT NULL COMMENT 'SKU ID（与商品服务中的 SKUID 对应）',
    available_stock INT NOT NULL DEFAULT 0 COMMENT '可用库存数量',
    version BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号（预留，当前未使用）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_sku_id_stock (sku_id, available_stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存主表（SKU 维度）';


-- ============================================
-- 2. 库存变动明细表（日志）
-- 用于记录每一次库存变更，便于审计与排查问题
-- ============================================
CREATE TABLE IF NOT EXISTS inventory_logs (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '日志ID',
    sku_id VARCHAR(26) NOT NULL COMMENT 'SKU ID',
    change_amount INT NOT NULL COMMENT '库存变动数量：正数增加，负数减少',
    reason VARCHAR(50) NOT NULL COMMENT '变动原因：order_created, order_canceled, manual_adjust 等',
    ref_id VARCHAR(64) DEFAULT NULL COMMENT '关联单号（订单号/操作单号等）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_sku_time (sku_id, created_at),
    INDEX idx_ref_id (ref_id),
    INDEX idx_sku_ref (sku_id, ref_id) COMMENT '用于幂等性检查：查询某个订单是否已扣减过某个SKU的库存'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存变动明细表';


