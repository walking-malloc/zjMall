package handler

import (
	"context"
	commonv1 "zjMall/gen/go/api/proto/common"
)

type HealthHandler struct {
	commonv1.UnimplementedHealthServiceServer
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(ctx context.Context, req *commonv1.HealthCheckRequest) (*commonv1.HealthCheckResponse, error) {
	return &commonv1.HealthCheckResponse{
		Status:  commonv1.HealthCheckResponse_SERVING,
		Message: "test successful",
	}, nil
}
