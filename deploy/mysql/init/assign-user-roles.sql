-- 为所有现有用户分配默认角色
-- 如果用户没有角色，则分配 'user' 角色
USE user_db;

-- 获取 'user' 角色ID（如果不存在则使用 'customer'）
SET @role_id = COALESCE(
    (SELECT id FROM roles WHERE code = 'user' LIMIT 1),
    (SELECT id FROM roles WHERE code = 'customer' LIMIT 1)
);

-- 为所有没有角色的用户分配角色
-- 注意：ID 使用固定前缀 + 时间戳 + 用户ID后10位生成（确保唯一性）
-- 实际生产环境建议使用应用代码生成 ULID
INSERT INTO user_roles (id, user_id, role_id)
SELECT 
    CONCAT('01UR', UNIX_TIMESTAMP(NOW()), RIGHT(u.id, 10)) as id,
    u.id as user_id,
    @role_id as role_id
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id
)
AND @role_id IS NOT NULL;
