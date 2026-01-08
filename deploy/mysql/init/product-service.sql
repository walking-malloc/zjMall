CREATE DATABASE IF NOT EXISTS product_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE product_db;
-- ============================================
-- 商品服务数据库表结构
-- ============================================

-- ============================================
-- 1. 类目表（支持多级类目）
-- ============================================
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    parent_id VARCHAR(26) COMMENT '父类目ID，顶级类目为NULL',
    name VARCHAR(100) NOT NULL COMMENT '类目名称',
    level TINYINT NOT NULL DEFAULT 1 COMMENT '类目层级：1-一级，2-二级，3-三级',
    is_leaf TINYINT DEFAULT 0 COMMENT '是否为叶子节点：0-否，1-是',
    is_visible TINYINT DEFAULT 1 COMMENT '是否在前台展示：0-否，1-是',
    sort_order INT DEFAULT 0 COMMENT '排序权重，数字越大越靠前',
    icon VARCHAR(255) COMMENT '类目图标URL',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    version INT DEFAULT 0 COMMENT '版本号',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    INDEX idx_parent_visible_status (parent_id, is_visible, status),
    UNIQUE KEY uk_parent_name (parent_id, name, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品类目表';

-- ============================================
-- 2. 品牌表
-- ============================================
CREATE TABLE IF NOT EXISTS brands (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '品牌名称',
    logo_url VARCHAR(255) COMMENT '品牌LOGO地址',
    country VARCHAR(50) COMMENT '所属国家/地区',
    description TEXT COMMENT '品牌描述',
    first_letter VARCHAR(1) COMMENT '品牌首字母（用于排序）',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    version INT DEFAULT 0 COMMENT '版本号',
    INDEX idx_first_letter (first_letter),
    INDEX idx_status_sort (status, sort_order),
    UNIQUE KEY uk_name_deleted (name, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌表';

-- ============================================
-- 3. 品牌类目关联表（多对多）
-- ============================================
CREATE TABLE IF NOT EXISTS brand_categories (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    brand_id VARCHAR(26) NOT NULL COMMENT '品牌ID',
    category_id VARCHAR(26) NOT NULL COMMENT '类目ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_brand_category (brand_id, category_id),
    INDEX idx_brand_id (brand_id),
    INDEX idx_category_id (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='品牌类目关联表';

-- ============================================
-- 4. 属性表（属性模板）
-- ============================================
CREATE TABLE IF NOT EXISTS attributes (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    category_id VARCHAR(26) NOT NULL COMMENT '所属类目ID',
    name VARCHAR(100) NOT NULL COMMENT '属性名称（如：颜色、尺寸、存储容量）',
    type TINYINT NOT NULL DEFAULT 1 COMMENT '属性类型：1-销售属性（用于生成SKU），2-非销售属性（仅展示）',
    input_type TINYINT NOT NULL DEFAULT 1 COMMENT '录入方式：1-单选，2-多选，3-文本，4-数值',
    is_required TINYINT DEFAULT 0 COMMENT '是否必填：0-否，1-是',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category_type_required (category_id, type, is_required)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='属性表（属性模板）';

-- ============================================
-- 5. 属性值表
-- ============================================
CREATE TABLE IF NOT EXISTS attribute_values (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    attribute_id VARCHAR(26) NOT NULL COMMENT '所属属性ID',
    value VARCHAR(100) NOT NULL COMMENT '属性值名称（如：红色、XL、128G）',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_attribute_sort (attribute_id, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='属性值表';

-- ============================================
-- 6. 商品表（SPU）
-- ============================================
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    category_id VARCHAR(26) NOT NULL COMMENT '所属类目ID',
    brand_id VARCHAR(26) COMMENT '品牌ID',
    title VARCHAR(200) NOT NULL COMMENT '商品标题',
    subtitle VARCHAR(200) COMMENT '商品副标题/卖点',
    main_image VARCHAR(255) NOT NULL COMMENT '主图URL',
    images TEXT COMMENT '轮播图URL列表（JSON数组）',
    description TEXT COMMENT '商品详情（富文本）',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-草稿，2-待审核，3-已上架，4-已下架，5-已删除',
    on_shelf_time TIMESTAMP NULL COMMENT '上架时间',
    off_shelf_time TIMESTAMP NULL COMMENT '下架时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    -- 前台按类目查询已上架商品（最常用，category_id选择性高放前面）
    INDEX idx_category_status_shelf (category_id, status, on_shelf_time),
    -- 按品牌查询已上架商品
    INDEX idx_brand_status_shelf (brand_id, status, on_shelf_time),
    -- 定时上架任务：查询待上架且到时间的商品
    INDEX idx_status_shelf_time (status, on_shelf_time),
    -- 后台管理：按状态和时间排序
    INDEX idx_status_created (status, created_at),
    -- 全文搜索索引
    FULLTEXT KEY ft_title (title, subtitle) COMMENT '全文索引，用于搜索'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表（SPU）';

-- ============================================
-- 7. SKU表（库存单元）
-- ============================================
CREATE TABLE IF NOT EXISTS skus (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    product_id VARCHAR(26) NOT NULL COMMENT '所属商品ID（SPU）',
    sku_code VARCHAR(50) UNIQUE COMMENT 'SKU编码（内部编码）',
    barcode VARCHAR(50) COMMENT '条形码',
    name VARCHAR(200) COMMENT 'SKU名称（如：黑色 128G）',
    price DECIMAL(10, 2) NOT NULL COMMENT '销售价格',
    original_price DECIMAL(10, 2) COMMENT '划线价/原价',
    cost_price DECIMAL(10, 2) COMMENT '成本价',
    weight DECIMAL(10, 2) COMMENT '重量（单位：kg）',
    volume DECIMAL(10, 2) COMMENT '体积（单位：m³）',
    image VARCHAR(255) COMMENT 'SKU图片（如不同颜色对应不同图片）',
    status TINYINT DEFAULT 1 COMMENT '状态：1-上架，2-下架，3-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '软删除时间',
    INDEX idx_product_status (product_id, status),
    INDEX idx_product_created  (product_id, created_at),
    INDEX idx_price_status (price, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU表（库存单元）';

-- ============================================
-- 8. SKU属性关联表（多对多）
-- ============================================
CREATE TABLE IF NOT EXISTS sku_attributes (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    sku_id VARCHAR(26) NOT NULL COMMENT 'SKU ID',
    attribute_value_id VARCHAR(26) NOT NULL COMMENT '属性值ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_sku_attribute_value (sku_id, attribute_value_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_attribute_value_id (attribute_value_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU属性关联表';

-- ============================================
-- 9. 标签表
-- ============================================
CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '标签名称（如：新品、爆款、清仓）',
    type TINYINT DEFAULT 1 COMMENT '标签类型：1-系统标签，2-运营标签',
    color VARCHAR(20) COMMENT '标签颜色（用于前端展示）',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status_sort(status,sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签表';

-- ============================================
-- 10. 商品标签关联表（多对多）
-- ============================================
CREATE TABLE IF NOT EXISTS product_tags (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    product_id VARCHAR(26) NOT NULL COMMENT '商品ID',
    tag_id VARCHAR(26) NOT NULL COMMENT '标签ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_product_tag (product_id, tag_id),
    INDEX idx_product_id (product_id),
    INDEX idx_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品标签关联表';

-- ============================================
-- 11. 审核日志表
-- ============================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    target_id VARCHAR(26) NOT NULL COMMENT '审核对象ID（商品ID）',
    target_type VARCHAR(50) NOT NULL DEFAULT 'product' COMMENT '审核对象类型：product-商品',
    action TINYINT NOT NULL COMMENT '操作：1-提交审核，2-审核通过，3-审核驳回',
    result TINYINT COMMENT '结果：1-通过，2-驳回',
    reason TEXT COMMENT '审核原因/驳回原因',
    operator_id VARCHAR(26) COMMENT '操作人ID',
    operator_name VARCHAR(50) COMMENT '操作人姓名',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_target (target_id, target_type),
    INDEX idx_operator_id (operator_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审核日志表';

-- ============================================
-- 12. 操作日志表
-- ============================================
CREATE TABLE IF NOT EXISTS operation_logs (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    target_id VARCHAR(26) NOT NULL COMMENT '操作对象ID',
    target_type VARCHAR(50) NOT NULL COMMENT '操作对象类型：product-商品，category-类目，brand-品牌',
    action VARCHAR(50) NOT NULL COMMENT '操作类型：create-创建，update-更新，delete-删除，on_shelf-上架，off_shelf-下架',
    old_value TEXT COMMENT '变更前数据（JSON格式）',
    new_value TEXT COMMENT '变更后数据（JSON格式）',
    operator_id VARCHAR(26) COMMENT '操作人ID',
    operator_name VARCHAR(50) COMMENT '操作人姓名',
    ip_address VARCHAR(50) COMMENT '操作IP地址',
    user_agent VARCHAR(255) COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_target (target_id, target_type),
    INDEX idx_operator_action (operator_id, action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';

