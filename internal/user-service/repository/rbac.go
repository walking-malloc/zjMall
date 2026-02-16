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

	// 用户角色相关
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*model.Role, error)
	GetUserRoleCodes(ctx context.Context, userID string) ([]string, error)
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

