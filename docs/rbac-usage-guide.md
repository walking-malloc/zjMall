# RBAC权限控制使用指南

## 概述

本文档介绍如何在各个服务中使用RBAC权限控制系统。

## 权限控制方案

### 方案1：在Handler中直接检查权限（推荐）

这是最简单直接的方式，适合细粒度的权限控制。

#### 步骤1：创建权限检查工具函数

在 `internal/common/middleware/rbac.go` 中添加工具函数：

```go
// CheckRole 检查用户是否具有指定角色之一
func CheckRole(ctx context.Context, requiredRoles ...string) bool {
    roles := GetRolesFromContext(ctx)
    if len(roles) == 0 {
        return false
    }
    
    for _, userRole := range roles {
        for _, requiredRole := range requiredRoles {
            if userRole == requiredRole {
                return true
            }
        }
    }
    return false
}

// CheckPermission 检查用户是否具有指定权限之一
// 需要传入权限检查函数（从用户服务获取权限）
func CheckPermission(ctx context.Context, permissionChecker func(userID string) ([]string, error), requiredPermissions ...string) (bool, error) {
    userID := GetUserIDFromContext(ctx)
    if userID == "" {
        return false, nil
    }
    
    permissions, err := permissionChecker(userID)
    if err != nil {
        return false, err
    }
    
    for _, userPerm := range permissions {
        for _, requiredPerm := range requiredPermissions {
            if userPerm == requiredPerm {
                return true, nil
            }
        }
    }
    return false, nil
}
```

#### 步骤2：在商品服务Handler中使用

**示例：商品上架接口**

```go
// internal/product-service/handler/product-service.go

func (h *ProductServiceHandler) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {
    // 1. 检查用户角色（方式1：基于角色）
    if !middleware.CheckRole(ctx, "merchant", "admin") {
        return &productv1.OnShelfProductResponse{
            Code:    403,
            Message: "权限不足：需要商家运营或管理员权限",
        }, nil
    }
    
    // 或者 2. 检查用户权限（方式2：基于权限）
    // hasPermission, err := middleware.CheckPermission(ctx, permissionChecker, "product:on_shelf")
    // if err != nil || !hasPermission {
    //     return &productv1.OnShelfProductResponse{
    //         Code:    403,
    //         Message: "权限不足：需要 product:on_shelf 权限",
    //     }, nil
    // }
    
    // 继续处理业务逻辑
    return h.productService.OnShelfProduct(ctx, req)
}
```

**示例：商品创建接口**

```go
func (h *ProductServiceHandler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
    // 检查商家或管理员角色
    if !middleware.CheckRole(ctx, "merchant", "admin") {
        return &productv1.CreateProductResponse{
            Code:    403,
            Message: "权限不足：需要商家运营或管理员权限",
        }, nil
    }
    
    return h.productService.CreateProduct(ctx, req)
}
```

**示例：商品查看接口（所有用户都可以）**

```go
func (h *ProductServiceHandler) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
    // 查看商品不需要特殊权限，所有登录用户都可以
    // 如果需要，可以检查 product:view 权限
    return h.productService.GetProduct(ctx, req)
}
```

### 方案2：使用中间件进行路由级权限控制

如果需要在路由级别统一控制权限，可以使用中间件。

#### 步骤1：在商品服务中创建权限检查器

```go
// internal/product-service/handler/product-service.go

// 创建权限检查函数（需要连接用户服务）
func createPermissionChecker() func(userID string) ([]string, error) {
    // 这里需要调用用户服务获取权限
    // 可以通过 gRPC 客户端调用用户服务的 GetUserPermissions
    return func(userID string) ([]string, error) {
        // TODO: 调用用户服务获取权限
        // 或者从缓存中获取
        return []string{}, nil
    }
}
```

#### 步骤2：在main.go中应用中间件

```go
// cmd/product-service/main.go

// 方式1：全局应用（所有接口都需要merchant或admin角色）
srv.UseMiddleware(
    middleware.CORS(middleware.DefaultCORSConfig()),
    middleware.Recovery(),
    middleware.Logging(),
    middleware.TraceID(),
    middleware.Auth(),
    middleware.RequireRole("merchant", "admin"), // 全局权限控制
)

// 方式2：特定路由应用（需要路由支持）
// 注意：当前server实现可能不支持路由级中间件
// 需要在Handler中手动检查
```

### 方案3：在gRPC服务中检查权限

对于gRPC服务，可以在拦截器或Handler中检查权限。

#### 创建gRPC权限拦截器

```go
// internal/common/middleware/grpc_rbac.go

package middleware

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// UnaryRoleInterceptor gRPC角色权限拦截器
func UnaryRoleInterceptor(requiredRoles ...string) grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // 从context获取角色
        roles := GetRolesFromContext(ctx)
        if len(roles) == 0 {
            return nil, status.Error(codes.PermissionDenied, "权限不足：需要角色权限")
        }
        
        // 检查是否有所需角色
        hasRole := false
        for _, userRole := range roles {
            for _, requiredRole := range requiredRoles {
                if userRole == requiredRole {
                    hasRole = true
                    break
                }
            }
            if hasRole {
                break
            }
        }
        
        if !hasRole {
            return nil, status.Error(codes.PermissionDenied, "权限不足：需要角色 "+strings.Join(requiredRoles, ", "))
        }
        
        return handler(ctx, req)
    }
}
```

#### 在main.go中注册拦截器

```go
// cmd/product-service/main.go

grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(
        middleware.UnaryAuthInterceptor,                    // 认证拦截器
        middleware.UnaryRoleInterceptor("merchant", "admin"), // 权限拦截器
    ),
)
```

## 实际使用示例

### 示例1：商品服务 - 上架商品

```go
// internal/product-service/handler/product-service.go

func (h *ProductServiceHandler) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {
    // 方式1：检查角色（推荐，因为角色在JWT中，无需查数据库）
    if !middleware.CheckRole(ctx, "merchant", "admin") {
        return &productv1.OnShelfProductResponse{
            Code:    403,
            Message: "权限不足：需要商家运营或管理员权限",
        }, nil
    }
    
    return h.productService.OnShelfProduct(ctx, req)
}
```

### 示例2：商品服务 - 创建商品

```go
func (h *ProductServiceHandler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
    // 检查角色
    if !middleware.CheckRole(ctx, "merchant", "admin") {
        return &productv1.CreateProductResponse{
            Code:    403,
            Message: "权限不足：需要商家运营或管理员权限",
        }, nil
    }
    
    return h.productService.CreateProduct(ctx, req)
}
```

### 示例3：订单服务 - 创建订单（所有用户都可以）

```go
// internal/order-service/handler/order-service.go

func (h *OrderServiceHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
    // 创建订单所有登录用户都可以，不需要特殊权限检查
    // 或者检查 order:purchase 权限
    userID := middleware.GetUserIDFromContext(ctx)
    if userID == "" {
        return &orderv1.CreateOrderResponse{
            Code:    401,
            Message: "未登录",
        }, nil
    }
    
    return h.orderService.CreateOrder(ctx, req)
}
```

## 权限检查最佳实践

### 1. 优先使用角色检查（性能更好）

```go
// ✅ 推荐：角色在JWT中，无需查数据库
if !middleware.CheckRole(ctx, "merchant", "admin") {
    return errorResponse
}
```

### 2. 需要细粒度控制时使用权限检查

```go
// 当需要更细粒度的权限控制时
hasPermission, err := middleware.CheckPermission(ctx, permissionChecker, "product:on_shelf")
if err != nil || !hasPermission {
    return errorResponse
}
```

### 3. 缓存权限信息

```go
// 在用户服务中缓存用户权限到Redis
// 减少数据库查询，提高性能
func GetUserPermissionsCached(ctx context.Context, userID string) ([]string, error) {
    // 1. 先查缓存
    // 2. 缓存未命中，查数据库
    // 3. 写入缓存
}
```

## 权限映射表

| 接口 | 所需角色 | 所需权限 | 说明 |
|------|---------|---------|------|
| 创建商品 | merchant, admin | product:create | 商家和管理员可以创建 |
| 更新商品 | merchant, admin | product:update | 商家和管理员可以更新 |
| 删除商品 | merchant, admin | product:delete | 商家和管理员可以删除 |
| 上架商品 | merchant, admin | product:on_shelf | 商家和管理员可以上架 |
| 下架商品 | merchant, admin | product:off_shelf | 商家和管理员可以下架 |
| 查看商品 | customer, merchant, admin | product:view | 所有用户都可以查看 |
| 创建订单 | customer, merchant, admin | order:purchase | 所有用户都可以下单 |

## 注意事项

1. **JWT Token刷新**：当用户角色变化时，需要重新登录获取新Token
2. **权限缓存**：建议将用户权限缓存到Redis，减少数据库查询
3. **性能考虑**：优先使用角色检查（JWT中），权限检查需要查数据库
4. **错误处理**：权限不足时返回统一的错误格式

## 下一步

1. 在 `middleware/rbac.go` 中添加 `CheckRole` 和 `CheckPermission` 工具函数
2. 在各个服务的Handler中添加权限检查
3. 根据业务需求调整权限映射
