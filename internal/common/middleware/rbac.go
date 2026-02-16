package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"zjMall/internal/common/authz"
	"zjMall/pkg"

	"google.golang.org/grpc/metadata"
)

// ContextKey 用于从 context 中获取角色和权限
const RolesKey ContextKey = "roles"
const PermissionsKey ContextKey = "permissions"

// GetRolesFromContext 从 context 中获取用户角色列表
func GetRolesFromContext(ctx context.Context) []string {
	// 1. 优先从 HTTP context 中获取
	if roles, ok := ctx.Value(RolesKey).([]string); ok && len(roles) > 0 {
		return roles
	}

	// 2. 从 gRPC metadata 中获取
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		roles := md.Get(string(RolesKey))
		if len(roles) > 0 {
			return roles
		}
	}

	return nil
}

// RequireRole 要求用户具有指定角色之一的中间件
// 用法: RequireRole("merchant", "admin")
func RequireRole(requiredRoles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从 context 中获取用户角色
			roles := GetRolesFromContext(r.Context())
			if len(roles) == 0 {
				http.Error(w, `{"code": 403, "message": "权限不足：需要角色权限"}`, http.StatusForbidden)
				return
			}

			// 检查用户是否有任一所需角色
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
				http.Error(w, `{"code": 403, "message": "权限不足：需要角色 "`+strings.Join(requiredRoles, ", ")+`"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission 要求用户具有指定权限之一的中间件
// 注意：这个中间件需要从数据库或缓存中查询用户权限
// 用法: RequirePermission("product:create", "product:update")
func RequirePermission(permissionChecker func(userID string) ([]string, error), requiredPermissions ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从 context 中获取用户ID
			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, `{"code": 401, "message": "未登录"}`, http.StatusUnauthorized)
				return
			}

			// 从数据库或缓存中获取用户权限
			permissions, err := permissionChecker(userID)
			if err != nil {
				log.Printf("获取用户权限失败: %v", err)
				http.Error(w, `{"code": 500, "message": "获取权限失败"}`, http.StatusInternalServerError)
				return
			}

			// 检查用户是否有任一所需权限
			hasPermission := false
			for _, userPerm := range permissions {
				for _, requiredPerm := range requiredPermissions {
					if userPerm == requiredPerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				http.Error(w, `{"code": 403, "message": "权限不足：需要权限 "`+strings.Join(requiredPermissions, ", ")+`"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ExtractRolesFromToken 从 JWT Token 中提取角色并放入 context
// 这个函数应该在认证中间件中调用
func ExtractRolesFromToken(tokenString string) ([]string, error) {
	claims, err := pkg.VerifyJWTWithClaims(tokenString)
	if err != nil {
		return nil, err
	}
	return claims.Roles, nil
}

// CheckRole 检查用户是否具有指定角色之一
// ctx: 上下文（包含用户角色信息）
// requiredRoles: 需要的角色列表（用户只需具有其中一个即可）
// 返回: true 表示有权限，false 表示无权限
func CheckRole(ctx context.Context, requiredRoles ...string) bool {
	roles := GetRolesFromContext(ctx)
	if len(roles) == 0 || len(requiredRoles) == 0 {
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
// ctx: 上下文（包含用户ID）
// permissionChecker: 权限检查函数，用于获取用户权限列表
// requiredPermissions: 需要的权限列表（用户只需具有其中一个即可）
// 返回: (是否有权限, 错误)
func CheckPermission(ctx context.Context, permissionChecker func(userID string) ([]string, error), requiredPermissions ...string) (bool, error) {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		return false, nil
	}

	if permissionChecker == nil {
		return false, nil
	}

	permissions, err := permissionChecker(userID)
	if err != nil {
		return false, err
	}

	if len(permissions) == 0 || len(requiredPermissions) == 0 {
		return false, nil
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

// CasbinRBAC 使用 Casbin 做基于角色的接口权限控制
// sub = role, obj = URL Path, act = HTTP Method
func CasbinRBAC() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 公共路径直接放行（和 Auth 中间件保持一致）
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 获取用户角色（由 Auth 中间件从 JWT 中注入）
			roles := GetRolesFromContext(r.Context())
			log.Printf("[CasbinRBAC] r.Context(): %v", r.Context())
			log.Printf("[CasbinRBAC] roles: %v", roles)
			if len(roles) == 0 {
				http.Error(w, `{"code": 403, "message": "无访问权限：未绑定角色"}`, http.StatusForbidden)
				return
			}

			obj := r.URL.Path
			act := r.Method
			log.Printf("obj:%v act:%v", obj, act)
			// 只要有一个角色通过，就允许访问
			for _, role := range roles {
				ok, err := authz.Enforcer.Enforce(role, obj, act)
				if err != nil {
					log.Printf("Casbin 鉴权失败: role=%s, obj=%s, act=%s, err=%v", role, obj, act, err)
					http.Error(w, `{"code": 500, "message": "权限校验失败"}`, http.StatusInternalServerError)
					return
				}
				if ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"code": 403, "message": "无访问权限"}`, http.StatusForbidden)
		})
	}
}
