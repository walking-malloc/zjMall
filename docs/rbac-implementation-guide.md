# RBAC权限管理系统实现指南

## 概述

本系统实现了基于角色的访问控制（RBAC）权限管理，支持用户角色分配、权限验证等功能。

## 数据库设计

### 表结构

1. **roles** - 角色表
   - `id`: 角色ID（ULID）
   - `code`: 角色代码（customer, merchant, admin）
   - `name`: 角色名称
   - `description`: 角色描述
   - `status`: 状态（1-启用，2-停用）

2. **permissions** - 权限表
   - `id`: 权限ID（ULID）
   - `code`: 权限代码（如：product:create）
   - `name`: 权限名称
   - `resource`: 资源类型（product, category, order等）
   - `action`: 操作类型（create, update, delete, view等）
   - `description`: 权限描述
   - `status`: 状态（1-启用，2-停用）

3. **role_permissions** - 角色权限关联表
   - `role_id`: 角色ID
   - `permission_id`: 权限ID

4. **user_roles** - 用户角色关联表
   - `user_id`: 用户ID
   - `role_id`: 角色ID

### 预置角色和权限

#### 角色
- **customer**（普通用户）：只能购买和查看商品
- **merchant**（商家运营）：可以创建、修改、上下架商品
- **admin**（管理员）：拥有所有权限

#### 权限
- `product:create` - 创建商品
- `product:update` - 更新商品
- `product:delete` - 删除商品
- `product:on_shelf` - 上架商品
- `product:off_shelf` - 下架商品
- `product:view` - 查看商品
- `category:manage` - 管理类目
- `brand:manage` - 管理品牌
- `sku:manage` - 管理SKU
- `order:purchase` - 购买下单
- `order:manage` - 管理订单

## 初始化数据库

执行SQL脚本初始化RBAC相关表：

```bash
mysql -u root -p < deploy/mysql/init/rbac.sql
```

或者手动执行SQL文件中的内容。

## JWT Token中的角色信息

登录成功后，JWT Token中会包含用户的角色信息：

```json
{
  "user_id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "roles": ["customer"],
  "exp": 1234567890,
  "iat": 1234567890
}
```

## 使用权限中间件

### 1. 基于角色的权限验证

在需要特定角色的接口上使用 `RequireRole` 中间件：

```go
// 示例：只有商家运营和管理员可以访问
middleware.RequireRole("merchant", "admin")
```

### 2. 基于权限的权限验证

在需要特定权限的接口上使用 `RequirePermission` 中间件：

```go
// 示例：需要商品创建权限
permissionChecker := func(userID string) ([]string, error) {
    // 从数据库或缓存获取用户权限
    return rbacService.GetUserPermissionCodes(ctx, userID)
}
middleware.RequirePermission(permissionChecker, "product:create")
```

## API接口

### 角色管理

#### 1. 为用户分配角色
```
POST /api/v1/users/{user_id}/roles
Body: {
  "role_code": "merchant"
}
```

#### 2. 移除用户角色
```
DELETE /api/v1/users/{user_id}/roles/{role_code}
```

#### 3. 查询用户角色列表
```
GET /api/v1/users/{user_id}/roles
```

#### 4. 查询用户权限列表
```
GET /api/v1/users/{user_id}/permissions
```

#### 5. 查询所有角色列表
```
GET /api/v1/roles?status=1
```

#### 6. 查询所有权限列表
```
GET /api/v1/permissions?resource=product&status=1
```

## 在商品服务中使用权限控制

### 示例：商品上架接口

在商品服务的handler中，添加权限验证：

```go
// 在product-service的handler中
func (h *ProductServiceHandler) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {
    // 检查用户是否有上架权限
    userID := middleware.GetUserIDFromContext(ctx)
    roles := middleware.GetRolesFromContext(ctx)
    
    // 方式1：基于角色验证
    hasPermission := false
    for _, role := range roles {
        if role == "merchant" || role == "admin" {
            hasPermission = true
            break
        }
    }
    
    if !hasPermission {
        return &productv1.OnShelfProductResponse{
            Code: 403,
            Message: "权限不足：需要商家运营或管理员权限",
        }, nil
    }
    
    // 继续处理业务逻辑...
}
```

### 方式2：使用权限代码验证

```go
// 从用户服务获取权限列表
permissions, err := rbacService.GetUserPermissionCodes(ctx, userID)
hasPermission := false
for _, perm := range permissions {
    if perm == "product:on_shelf" {
        hasPermission = true
        break
    }
}
```

## 使用流程

### 1. 用户注册
- 新用户注册后，默认没有角色（或可以自动分配customer角色）
- 登录时JWT Token中会包含用户的角色信息

### 2. 分配商家角色
```bash
# 管理员为用户分配商家角色
curl -X POST http://localhost:8080/api/v1/users/{user_id}/roles \
  -H "Authorization: Bearer {admin_token}" \
  -H "Content-Type: application/json" \
  -d '{"role_code": "merchant"}'
```

### 3. 商家创建商品
```bash
# 商家创建商品（需要merchant角色）
curl -X POST http://localhost:8080/api/v1/product/products \
  -H "Authorization: Bearer {merchant_token}" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

### 4. 商家上架商品
```bash
# 商家上架商品（需要product:on_shelf权限）
curl -X POST http://localhost:8080/api/v1/product/products/{product_id}/on-shelf \
  -H "Authorization: Bearer {merchant_token}"
```

## 注意事项

1. **JWT Token刷新**：当用户角色发生变化时，需要重新登录以获取新的Token（包含新的角色信息）

2. **权限缓存**：建议将用户权限缓存到Redis，减少数据库查询

3. **权限检查位置**：
   - 网关层：统一权限验证（推荐）
   - 服务层：业务逻辑中的权限验证
   - 数据层：数据访问权限控制

4. **默认角色**：新用户注册时，可以自动分配customer角色

## 扩展建议

1. **权限缓存**：实现权限缓存机制，提高性能
2. **动态权限**：支持动态添加权限，无需修改代码
3. **权限继承**：支持角色权限继承
4. **资源级权限**：支持更细粒度的资源级权限控制（如：只能管理自己的商品）
