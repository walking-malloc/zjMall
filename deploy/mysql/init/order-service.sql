-- Active: 1769663211548@@127.0.0.1@3307@mysql
-- 订单服务数据库初始化脚本

-- 创建数据库（如需与其他服务隔离，可单独一个库）
CREATE DATABASE IF NOT EXISTS order_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE order_db;

-- ============================================
-- 1. 订单主表
-- 对应 Go 模型：internal/order-service/model/order.go
-- ============================================
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '主键ID（ULID）',
    order_no VARCHAR(32) NOT NULL UNIQUE COMMENT '订单号',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',

    status TINYINT NOT NULL DEFAULT 1 COMMENT '订单状态：1-待支付，2-待发货，3-待收货，4-已完成，5-已关闭',

    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '商品总金额',
    discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '优惠总金额',
    shipping_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '运费金额',
    pay_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '应付金额',

    receiver_name VARCHAR(50) COMMENT '收货人姓名',
    receiver_phone VARCHAR(20) COMMENT '收货人电话',
    receiver_address VARCHAR(255) COMMENT '收货地址（完整地址快照）',

    buyer_remark VARCHAR(255) COMMENT '买家留言',

    pay_channel VARCHAR(20) COMMENT '支付渠道：alipay、wechat 等',
    pay_trade_no VARCHAR(64) COMMENT '支付渠道流水号',

    items_snapshot JSON COMMENT '商品列表精简快照（JSON格式，包含商品基本信息，用于快速查看订单商品）',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    paid_at TIMESTAMP NULL DEFAULT NULL COMMENT '支付时间',
    shipped_at TIMESTAMP NULL DEFAULT NULL COMMENT '发货时间',
    completed_at TIMESTAMP NULL DEFAULT NULL COMMENT '完成时间',
    version INT NOT NULL DEFAULT 0 COMMENT '版本号',

    INDEX idx_user_status (user_id, status),
    INDEX idx_created_at (created_at),
    INDEX idx_order_no_user (order_no, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';


-- ============================================
-- 2. 订单明细表
-- 对应 Go 模型：internal/order-service/model/order_item.go
-- ============================================
CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '明细ID（ULID）',

    order_no VARCHAR(32) NOT NULL COMMENT '订单号',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',

    product_id VARCHAR(26) NOT NULL COMMENT '商品ID（SPU ID 快照）',
    sku_id VARCHAR(26) NOT NULL COMMENT 'SKU ID 快照',
    product_title VARCHAR(200) NOT NULL COMMENT '商品标题快照',
    product_image VARCHAR(255) COMMENT '商品图片快照',
    sku_name VARCHAR(100) COMMENT 'SKU 名称快照',

    price DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '单价快照',
    quantity INT NOT NULL DEFAULT 1 COMMENT '购买数量',
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '小计金额（price * quantity - 分摊优惠）',

    item_snapshot JSON COMMENT '商品详细快照（JSON格式，包含商品完整信息，用于审计和对账）',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_order_no (order_no),
    INDEX idx_user_order (user_id, order_no),
    INDEX idx_product_sku (product_id, sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';


