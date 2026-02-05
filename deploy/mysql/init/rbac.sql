-- Active: 1769663211548@@127.0.0.1@3307@user_db
-- RBAC权限管理相关表
USE user_db;

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL COMMENT '角色代码（如：customer, merchant, admin）',
    name VARCHAR(100) NOT NULL COMMENT '角色名称',
    description VARCHAR(255) COMMENT '角色描述',
    status TINYINT(1) DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL COMMENT '权限代码（如：product:create, product:update）',
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    resource VARCHAR(50) NOT NULL COMMENT '资源类型（如：product, category, order）',
    action VARCHAR(50) NOT NULL COMMENT '操作类型（如：create, update, delete, view）',
    description VARCHAR(255) COMMENT '权限描述',
    status TINYINT(1) DEFAULT 1 COMMENT '状态：1-启用，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_resource_action (resource, action),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    role_id VARCHAR(26) NOT NULL COMMENT '角色ID',
    permission_id VARCHAR(26) NOT NULL COMMENT '权限ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(26) NOT NULL PRIMARY KEY,
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
    role_id VARCHAR(26) NOT NULL COMMENT '角色ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

-- 初始化角色数据
INSERT INTO roles (id, code, name, description, status) VALUES
('01ARZ3NDEKTSV4RRFFQ69G5FAV', 'customer', '普通用户', '普通用户，只能购买和查看商品', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FBV', 'merchant', '商家运营', '商家运营，可以创建、修改、上下架商品', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FCV', 'admin', '管理员', '系统管理员，拥有所有权限', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description);

-- 初始化权限数据
INSERT INTO permissions (id, code, name, resource, action, description, status) VALUES
-- 商品相关权限
('01ARZ3NDEKTSV4RRFFQ69G5FDV', 'product:create', '创建商品', 'product', 'create', '创建商品权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FEV', 'product:update', '更新商品', 'product', 'update', '更新商品权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FFV', 'product:delete', '删除商品', 'product', 'delete', '删除商品权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FGV', 'product:on_shelf', '上架商品', 'product', 'on_shelf', '上架商品权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FHV', 'product:off_shelf', '下架商品', 'product', 'off_shelf', '下架商品权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FIV', 'product:view', '查看商品', 'product', 'view', '查看商品权限', 1),
-- 类目相关权限
('01ARZ3NDEKTSV4RRFFQ69G5FJV', 'category:manage', '管理类目', 'category', 'manage', '管理类目权限（创建、更新、删除）', 1),
-- 品牌相关权限
('01ARZ3NDEKTSV4RRFFQ69G5FKV', 'brand:manage', '管理品牌', 'brand', 'manage', '管理品牌权限（创建、更新、删除）', 1),
-- 订单相关权限
('01ARZ3NDEKTSV4RRFFQ69G5FLV', 'order:purchase', '购买下单', 'order', 'purchase', '购买下单权限', 1),
('01ARZ3NDEKTSV4RRFFQ69G5FMV', 'order:manage', '管理订单', 'order', 'manage', '管理订单权限', 1),
-- SKU相关权限
('01ARZ3NDEKTSV4RRFFQ69G5FNV', 'sku:manage', '管理SKU', 'sku', 'manage', '管理SKU权限', 1)
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description);

-- 为普通用户角色分配权限
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01ARZ3NDEKTSV4RRFFQ69G5F01', (SELECT id FROM roles WHERE code = 'customer'), (SELECT id FROM permissions WHERE code = 'product:view')),
('01ARZ3NDEKTSV4RRFFQ69G5F02', (SELECT id FROM roles WHERE code = 'customer'), (SELECT id FROM permissions WHERE code = 'order:purchase'))
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为商家运营角色分配权限
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01ARZ3NDEKTSV4RRFFQ69G5F10', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:create')),
('01ARZ3NDEKTSV4RRFFQ69G5F11', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:update')),
('01ARZ3NDEKTSV4RRFFQ69G5F12', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:delete')),
('01ARZ3NDEKTSV4RRFFQ69G5F13', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:on_shelf')),
('01ARZ3NDEKTSV4RRFFQ69G5F14', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:off_shelf')),
('01ARZ3NDEKTSV4RRFFQ69G5F15', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'product:view')),
('01ARZ3NDEKTSV4RRFFQ69G5F16', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'sku:manage')),
('01ARZ3NDEKTSV4RRFFQ69G5F17', (SELECT id FROM roles WHERE code = 'merchant'), (SELECT id FROM permissions WHERE code = 'order:purchase'))
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);

-- 为管理员角色分配所有权限
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01ARZ3NDEKTSV4RRFFQ69G5F20', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:create')),
('01ARZ3NDEKTSV4RRFFQ69G5F21', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:update')),
('01ARZ3NDEKTSV4RRFFQ69G5F22', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:delete')),
('01ARZ3NDEKTSV4RRFFQ69G5F23', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:on_shelf')),
('01ARZ3NDEKTSV4RRFFQ69G5F24', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:off_shelf')),
('01ARZ3NDEKTSV4RRFFQ69G5F25', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'product:view')),
('01ARZ3NDEKTSV4RRFFQ69G5F26', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'category:manage')),
('01ARZ3NDEKTSV4RRFFQ69G5F27', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'brand:manage')),
('01ARZ3NDEKTSV4RRFFQ69G5F28', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'order:purchase')),
('01ARZ3NDEKTSV4RRFFQ69G5F29', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'order:manage')),
('01ARZ3NDEKTSV4RRFFQ69G5F30', (SELECT id FROM roles WHERE code = 'admin'), (SELECT id FROM permissions WHERE code = 'sku:manage'))
ON DUPLICATE KEY UPDATE role_id=VALUES(role_id);
