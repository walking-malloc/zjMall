package service

import (
	"context"
	"fmt"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/user-service/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type RBACService struct {
	rbacRepo repository.RBACRepository
}

func NewRBACService(rbacRepo repository.RBACRepository) *RBACService {
	return &RBACService{
		rbacRepo: rbacRepo,
	}
}

// AssignRole 为用户分配角色
func (s *RBACService) AssignRole(ctx context.Context, userID string, roleCode string) error {
	// 1. 根据角色代码获取角色
	role, err := s.rbacRepo.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return fmt.Errorf("获取角色失败: %v", err)
	}
	if role == nil {
		return fmt.Errorf("角色不存在: %s", roleCode)
	}

	// 2. 检查用户是否已有该角色
	userRoles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取用户角色失败: %v", err)
	}
	for _, ur := range userRoles {
		if ur.ID == role.ID {
			return fmt.Errorf("用户已拥有该角色")
		}
	}

	// 3. 分配角色
	return s.rbacRepo.AssignRoleToUser(ctx, userID, role.ID)
}

// RemoveRole 移除用户角色
func (s *RBACService) RemoveRole(ctx context.Context, userID string, roleCode string) error {
	// 1. 根据角色代码获取角色
	role, err := s.rbacRepo.GetRoleByCode(ctx, roleCode)
	if err != nil {
		return fmt.Errorf("获取角色失败: %v", err)
	}
	if role == nil {
		return fmt.Errorf("角色不存在: %s", roleCode)
	}

	// 2. 移除角色
	return s.rbacRepo.RemoveRoleFromUser(ctx, userID, role.ID)
}

// GetUserRoles 获取用户角色列表
func (s *RBACService) GetUserRoles(ctx context.Context, userID string) ([]*userv1.RoleInfo, error) {
	roles, err := s.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %v", err)
	}

	result := make([]*userv1.RoleInfo, 0, len(roles))
	for _, role := range roles {
		result = append(result, &userv1.RoleInfo{
			Id:          role.ID,
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
			Status:      int32(role.Status),
			CreatedAt:   timestamppb.New(role.CreatedAt),
			UpdatedAt:   timestamppb.New(role.UpdatedAt),
		})
	}
	return result, nil
}

// ListRoles 查询所有角色列表
func (s *RBACService) ListRoles(ctx context.Context, status *int32) ([]*userv1.RoleInfo, error) {
	var statusPtr *int8
	if status != nil {
		s := int8(*status)
		statusPtr = &s
	}

	roles, err := s.rbacRepo.ListRoles(ctx, statusPtr)
	if err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %v", err)
	}

	result := make([]*userv1.RoleInfo, 0, len(roles))
	for _, role := range roles {
		result = append(result, &userv1.RoleInfo{
			Id:          role.ID,
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
			Status:      int32(role.Status),
			CreatedAt:   timestamppb.New(role.CreatedAt),
			UpdatedAt:   timestamppb.New(role.UpdatedAt),
		})
	}
	return result, nil
}

// GetUserRoleCodes 获取用户角色代码列表（用于JWT）
func (s *RBACService) GetUserRoleCodes(ctx context.Context, userID string) ([]string, error) {
	return s.rbacRepo.GetUserRoleCodes(ctx, userID)
}

// 之前这里有基于数据库的权限查询逻辑（Permission、RolePermission），
// 现在改用 Casbin 配置文件做权限控制，这些方法已不再需要。
