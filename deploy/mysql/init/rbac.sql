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

