package handler

import (
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/user-service/service"
)

type UserServiceHandler struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService // 依赖注入：业务逻辑层
}

func NewUserServiceHandler(userService *service.UserService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService, // 初始化 service
	}
}
