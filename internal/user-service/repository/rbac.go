package repository

import (
	"context"
	"zjMall/internal/user-service/model"

	"gorm.io/gorm"
)

type RBACRepository interface {
	// 角色相关
	GetRoleByCode(ctx context.Context, code string) (*model.Role, error)
	GetRoleByID(ctx context.Context, id string) (*model.Role, error)
	ListRoles(ctx context.Context, status *int8) ([]*model.Role, error)

	// 权限相关
	GetPermissionByCode(ctx context.Context, code string) (*model.Permission, error)
	ListPermissions(ctx context.Context, resource *string, status *int8) ([]*model.Permission, error)

	// 用户角色相关
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	GetUserRoleCodes(ctx context.Context, userID string) ([]string, error)

	// 角色权限相关
	GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error)
	GetUserPermissionCodes(ctx context.Context, userID string) ([]string, error)
}

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{
		db: db,
	}
}

// GetRoleByCode 根据角色代码获取角色
func (r *rbacRepository) GetRoleByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByID 根据角色ID获取角色
func (r *rbacRepository) GetRoleByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles 查询角色列表
func (r *rbacRepository) ListRoles(ctx context.Context, status *int8) ([]*model.Role, error) {
	var roles []*model.Role
	query := r.db.WithContext(ctx)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetPermissionByCode 根据权限代码获取权限
func (r *rbacRepository) GetPermissionByCode(ctx context.Context, code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// ListPermissions 查询权限列表
func (r *rbacRepository) ListPermissions(ctx context.Context, resource *string, status *int8) ([]*model.Permission, error) {
	var permissions []*model.Permission
	query := r.db.WithContext(ctx)
	if resource != nil {
		query = query.Where("resource = ?", *resource)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// AssignRoleToUser 为用户分配角色
func (r *rbacRepository) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	userRole := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

// RemoveRoleFromUser 移除用户角色
func (r *rbacRepository) RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error
}

// GetUserRoles 获取用户的所有角色
func (r *rbacRepository) GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).
		Model(&model.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, 1).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetUserRoleCodes 获取用户的所有角色代码
func (r *rbacRepository) GetUserRoleCodes(ctx context.Context, userID string) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&model.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, 1).
		Pluck("roles.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// GetRolePermissions 获取角色的所有权限
func (r *rbacRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.WithContext(ctx).
		Model(&model.Permission{}).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ? AND permissions.status = ?", roleID, 1).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetUserPermissions 获取用户的所有权限（通过角色）
func (r *rbacRepository) GetUserPermissions(ctx context.Context, userID string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("DISTINCT permissions.*").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_roles.user_id = ? AND roles.status = ? AND permissions.status = ?", userID, 1, 1).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetUserPermissionCodes 获取用户的所有权限代码
func (r *rbacRepository) GetUserPermissionCodes(ctx context.Context, userID string) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("DISTINCT permissions.code").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_roles.user_id = ? AND roles.status = ? AND permissions.status = ?", userID, 1, 1).
		Pluck("permissions.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}
