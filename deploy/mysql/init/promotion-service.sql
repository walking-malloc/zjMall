CREATE DATABASE IF NOT EXISTS promotion_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE promotion_db;

-- ============================================
-- 促销服务数据库表结构
-- ============================================

-- ============================================
-- 1. 促销活动表
-- ============================================
CREATE TABLE IF NOT EXISTS promotions (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '促销名称',
    type TINYINT NOT NULL COMMENT '促销类型：1-满减，2-满折，3-直降，4-限时折扣',
    description TEXT COMMENT '促销描述',
    product_ids TEXT COMMENT '适用商品ID列表（JSON数组，空表示全平台）',
    category_ids TEXT COMMENT '适用类目ID列表（JSON数组）',
    condition_value VARCHAR(50) COMMENT '条件值（如：满200）',
    discount_value VARCHAR(50) COMMENT '优惠值（如：减30 或 打8折）',
    start_time TIMESTAMP NOT NULL COMMENT '开始时间',
    end_time TIMESTAMP NOT NULL COMMENT '结束时间',
    max_use_times INT DEFAULT 0 COMMENT '每人限用次数（0表示不限制）',
    total_quota INT DEFAULT 0 COMMENT '总配额（0表示不限制）',
    used_quota INT DEFAULT 0 COMMENT '已使用配额',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    status TINYINT DEFAULT 1 COMMENT '状态：1-草稿，2-进行中，3-已暂停，4-已结束，5-已删除',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    INDEX idx_type_status (type, status, deleted_at),
    INDEX idx_time_range (start_time, end_time, status, deleted_at),
    INDEX idx_status_sort (status, sort_order, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='促销活动表';

-- ============================================
-- 2. 促销使用记录表（用于限购判断）
-- ============================================
CREATE TABLE IF NOT EXISTS promotion_usage_logs (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    promotion_id VARCHAR(26) NOT NULL COMMENT '促销活动ID',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
    order_id VARCHAR(26) COMMENT '订单ID',
    discount_amount DECIMAL(10, 2) COMMENT '优惠金额',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_promotion_user (promotion_id, user_id),
    INDEX idx_user_created (user_id, created_at),
    INDEX idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='促销使用记录表';

-- ============================================
-- 3. 优惠券模板表
-- ============================================
CREATE TABLE IF NOT EXISTS coupon_templates (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '优惠券名称',
    type TINYINT NOT NULL COMMENT '优惠券类型：1-固定金额，2-折扣，3-免运费',
    description TEXT COMMENT '优惠券描述',
    discount_value VARCHAR(50) NOT NULL COMMENT '优惠值（固定金额或折扣）',
    condition_value VARCHAR(50) COMMENT '使用条件（如：满100可用）',
    total_count INT DEFAULT 0 COMMENT '发放总数（0表示不限制）',
    claimed_count INT DEFAULT 0 COMMENT '已领取数量',
    per_user_limit INT DEFAULT 1 COMMENT '每人限领数量',
    valid_start_time TIMESTAMP NOT NULL COMMENT '有效期开始时间',
    valid_end_time TIMESTAMP NOT NULL COMMENT '有效期结束时间',
    valid_days INT DEFAULT 0 COMMENT '领取后有效天数（0表示使用模板有效期）',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    INDEX idx_status_time (status, valid_start_time, valid_end_time, deleted_at),
    INDEX idx_type_status (type, status, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';

-- ============================================
-- 4. 优惠券实例表（用户领取的优惠券）
-- ============================================
CREATE TABLE IF NOT EXISTS coupons (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    template_id VARCHAR(26) NOT NULL COMMENT '优惠券模板ID',
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
    name VARCHAR(100) NOT NULL COMMENT '优惠券名称',
    type TINYINT NOT NULL COMMENT '优惠券类型',
    description TEXT COMMENT '优惠券描述',
    discount_value VARCHAR(50) NOT NULL COMMENT '优惠值',
    condition_value VARCHAR(50) COMMENT '使用条件',
    status TINYINT DEFAULT 1 COMMENT '状态：1-未使用，2-已使用，3-已过期',
    valid_start_time TIMESTAMP NOT NULL COMMENT '有效期开始时间',
    valid_end_time TIMESTAMP NOT NULL COMMENT '有效期结束时间',
    used_at TIMESTAMP NULL COMMENT '使用时间',
    order_id VARCHAR(26) COMMENT '使用的订单ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status),
    INDEX idx_template_id (template_id),
    INDEX idx_user_valid_time (user_id, valid_start_time, valid_end_time),
    INDEX idx_status_valid_time (status, valid_end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券实例表';