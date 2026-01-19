-- Active: 1768481430314@@127.0.0.1@3306@cart_db
CREATE DATABASE IF NOT EXISTS cart_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cart_db;

-- 购物车表
CREATE TABLE IF NOT EXISTS cart_items (
    id VARCHAR(26) NOT NULL PRIMARY KEY COMMENT '购物车项ID',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
    product_id VARCHAR(26) NOT NULL COMMENT '商品ID（SPU ID）',
    sku_id VARCHAR(26) NOT NULL COMMENT 'SKU ID',
    product_title VARCHAR(200) NOT NULL COMMENT '商品标题',
    product_image VARCHAR(255) COMMENT '商品主图',
    sku_name VARCHAR(100) COMMENT 'SKU 名称（规格描述）',
    price DECIMAL(10, 2) NOT NULL COMMENT '单价（加购时的价格快照）',
    current_price DECIMAL(10, 2) NOT NULL COMMENT '当前价格（实时查询）',
    quantity INT NOT NULL DEFAULT 1 COMMENT '数量',
    stock INT NOT NULL DEFAULT 0 COMMENT '当前库存',
    is_valid TINYINT(1) DEFAULT 1 COMMENT '是否有效：0-无效，1-有效',
    invalid_reason VARCHAR(100) COMMENT '失效原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_id (user_id),
    INDEX idx_product_sku (product_id, sku_id),
    INDEX idx_user_product_sku (user_id, product_id, sku_id),
    UNIQUE KEY uk_user_sku (user_id, sku_id) COMMENT '同一用户同一SKU只能有一条记录'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车表';

