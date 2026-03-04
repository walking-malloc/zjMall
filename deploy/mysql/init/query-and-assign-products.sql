-- 查询现有商品并分配 SKU、品牌、类目
USE product_db;

-- ============================================
-- 1. 查询现有商品
-- ============================================
SELECT 
    id,
    title,
    category_id,
    brand_id,
    status,
    created_at
FROM products
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- ============================================
-- 2. 如果没有类目，先创建一些测试类目
-- ============================================
-- 一级类目：电子产品
INSERT INTO categories (id, parent_id, name, level, is_leaf, is_visible, sort_order, status)
VALUES 
('01CAT000000000000000000001', NULL, '电子产品', 1, 0, 1, 100, 1),
('01CAT000000000000000000002', NULL, '服装鞋帽', 1, 0, 1, 99, 1),
('01CAT000000000000000000003', NULL, '美妆护肤', 1, 0, 1, 98, 1),
('01CAT000000000000000000004', NULL, '家用电器', 1, 0, 1, 97, 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 二级类目：手机
INSERT INTO categories (id, parent_id, name, level, is_leaf, is_visible, sort_order, status)
VALUES 
('01CAT000000000000000000101', '01CAT000000000000000000001', '手机', 2, 0, 1, 100, 1),
('01CAT000000000000000000102', '01CAT000000000000000000001', '电脑', 2, 0, 1, 99, 1),
('01CAT000000000000000000103', '01CAT000000000000000000001', '平板', 2, 1, 1, 98, 1),
('01CAT000000000000000000104', '01CAT000000000000000000001', '耳机', 2, 1, 1, 97, 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 三级类目：智能手机
INSERT INTO categories (id, parent_id, name, level, is_leaf, is_visible, sort_order, status)
VALUES 
('01CAT000000000000000000201', '01CAT000000000000000000101', '智能手机', 3, 1, 1, 100, 1),
('01CAT000000000000000000202', '01CAT000000000000000000101', '功能手机', 3, 1, 1, 99, 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- ============================================
-- 3. 为没有品牌和类目的商品智能分配
-- ============================================
-- 根据商品标题关键词匹配品牌
UPDATE products p
LEFT JOIN brands b ON (
    (p.title LIKE CONCAT('%', b.name, '%') OR p.title LIKE CONCAT('%', LOWER(b.name), '%'))
    AND b.deleted_at IS NULL
    AND b.status = 1
)
SET p.brand_id = b.id
WHERE 
    p.deleted_at IS NULL
    AND (p.brand_id IS NULL OR p.brand_id = '')
    AND b.id IS NOT NULL;

-- 为仍然没有品牌的商品分配默认品牌（Apple 或第一个品牌）
SET @default_brand_id = COALESCE(
    (SELECT id FROM brands WHERE name = 'Apple' AND deleted_at IS NULL LIMIT 1),
    (SELECT id FROM brands WHERE deleted_at IS NULL LIMIT 1)
);

UPDATE products 
SET brand_id = @default_brand_id
WHERE 
    deleted_at IS NULL
    AND (brand_id IS NULL OR brand_id = '')
    AND @default_brand_id IS NOT NULL;

-- 根据商品标题关键词匹配类目
UPDATE products p
LEFT JOIN categories c ON (
    (
        (p.title LIKE '%手机%' OR p.title LIKE '%iPhone%' OR p.title LIKE '%华为%' OR p.title LIKE '%小米%' OR p.title LIKE '%OPPO%' OR p.title LIKE '%vivo%' OR p.title LIKE '%三星%')
        AND c.name = '智能手机'
    )
    OR (
        (p.title LIKE '%电脑%' OR p.title LIKE '%笔记本%' OR p.title LIKE '%MacBook%')
        AND c.name LIKE '%电脑%'
    )
    OR (
        (p.title LIKE '%平板%' OR p.title LIKE '%iPad%')
        AND c.name = '平板'
    )
    OR (
        (p.title LIKE '%耳机%' OR p.title LIKE '%AirPods%')
        AND c.name = '耳机'
    )
    AND c.deleted_at IS NULL
    AND c.status = 1
    AND c.level = 3
)
SET p.category_id = c.id
WHERE 
    p.deleted_at IS NULL
    AND (p.category_id IS NULL OR p.category_id = '')
    AND c.id IS NOT NULL;

-- 为仍然没有类目的商品分配默认类目（智能手机）
SET @default_category_id = COALESCE(
    (SELECT id FROM categories WHERE name = '智能手机' AND level = 3 AND deleted_at IS NULL LIMIT 1),
    (SELECT id FROM categories WHERE level = 3 AND deleted_at IS NULL LIMIT 1)
);

UPDATE products 
SET category_id = @default_category_id
WHERE 
    deleted_at IS NULL
    AND (category_id IS NULL OR category_id = '')
    AND @default_category_id IS NOT NULL;

-- ============================================
-- 4. 为没有 SKU 的商品创建默认 SKU
-- ============================================
-- 为每个没有 SKU 的商品创建一个默认 SKU
-- 根据商品标题推断价格范围（示例逻辑）
INSERT INTO skus (
    id,
    product_id,
    sku_code,
    name,
    price,
    original_price,
    cost_price,
    status,
    created_at
)
SELECT 
    CONCAT('01SKU', UNIX_TIMESTAMP(NOW()), RIGHT(p.id, 10), LPAD(ROW_NUMBER() OVER (ORDER BY p.id), 3, '0')) as id,
    p.id as product_id,
    CONCAT('SKU-', p.id, '-001') as sku_code,
    CONCAT(p.title, ' 默认规格') as name,
    CASE 
        WHEN p.title LIKE '%Pro%' OR p.title LIKE '%Max%' THEN 8999.00
        WHEN p.title LIKE '%Plus%' THEN 6999.00
        WHEN p.title LIKE '%256%' OR p.title LIKE '%512%' THEN 5999.00
        WHEN p.title LIKE '%128%' THEN 4999.00
        ELSE 3999.00
    END as price,
    CASE 
        WHEN p.title LIKE '%Pro%' OR p.title LIKE '%Max%' THEN 9999.00
        WHEN p.title LIKE '%Plus%' THEN 7999.00
        WHEN p.title LIKE '%256%' OR p.title LIKE '%512%' THEN 6999.00
        WHEN p.title LIKE '%128%' THEN 5999.00
        ELSE 4999.00
    END as original_price,
    CASE 
        WHEN p.title LIKE '%Pro%' OR p.title LIKE '%Max%' THEN 7000.00
        WHEN p.title LIKE '%Plus%' THEN 5500.00
        WHEN p.title LIKE '%256%' OR p.title LIKE '%512%' THEN 4500.00
        WHEN p.title LIKE '%128%' THEN 3500.00
        ELSE 2500.00
    END as cost_price,
    1 as status,
    NOW() as created_at
FROM products p
WHERE 
    p.deleted_at IS NULL
    AND p.brand_id IS NOT NULL
    AND p.category_id IS NOT NULL
    AND NOT EXISTS (
        SELECT 1 FROM skus s 
        WHERE s.product_id = p.id 
        AND s.deleted_at IS NULL
    )
LIMIT 100;  -- 限制一次最多创建 100 个 SKU

-- ============================================
-- 5. 查询更新后的商品信息
-- ============================================
SELECT 
    p.id,
    p.title,
    p.category_id,
    c.name as category_name,
    p.brand_id,
    b.name as brand_name,
    COUNT(s.id) as sku_count,
    MIN(s.price) as min_price,
    MAX(s.price) as max_price,
    p.status
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN skus s ON s.product_id = p.id AND s.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.title, p.category_id, c.name, p.brand_id, b.name, p.status
ORDER BY p.created_at DESC;
