package handler

import (
	"context"
	promotionv1 "zjMall/gen/go/api/proto/promotion"
	"zjMall/internal/promotion-service/service"
)

type PromotionServiceHandler struct {
	promotionv1.UnimplementedPromotionServiceServer
	promotionService *service.PromotionService
}

func NewPromotionServiceHandler(promotionService *service.PromotionService) *PromotionServiceHandler {
	return &PromotionServiceHandler{
		promotionService: promotionService,
	}
}

func (h *PromotionServiceHandler) CreatePromotion(ctx context.Context, req *promotionv1.CreatePromotionRequest) (*promotionv1.CreatePromotionResponse, error) {
	return h.promotionService.CreatePromotion(ctx, req)
}

func (h *PromotionServiceHandler) GetPromotion(ctx context.Context, req *promotionv1.GetPromotionRequest) (*promotionv1.GetPromotionResponse, error) {
	return h.promotionService.GetPromotion(ctx, req)
}

func (h *PromotionServiceHandler) ListPromotions(ctx context.Context, req *promotionv1.ListPromotionsRequest) (*promotionv1.ListPromotionsResponse, error) {
	return h.promotionService.ListPromotions(ctx, req)
}

func (h *PromotionServiceHandler) UpdatePromotion(ctx context.Context, req *promotionv1.UpdatePromotionRequest) (*promotionv1.UpdatePromotionResponse, error) {
	// TODO: 实现更新逻辑
	return &promotionv1.UpdatePromotionResponse{
		Code:    0,
		Message: "更新成功",
	}, nil
}

func (h *PromotionServiceHandler) DeletePromotion(ctx context.Context, req *promotionv1.DeletePromotionRequest) (*promotionv1.DeletePromotionResponse, error) {
	// TODO: 实现删除逻辑
	return &promotionv1.DeletePromotionResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

func (h *PromotionServiceHandler) GetAvailablePromotions(ctx context.Context, req *promotionv1.GetAvailablePromotionsRequest) (*promotionv1.GetAvailablePromotionsResponse, error) {
	return h.promotionService.GetAvailablePromotions(ctx, req)
}

func (h *PromotionServiceHandler) CalculateDiscount(ctx context.Context, req *promotionv1.CalculateDiscountRequest) (*promotionv1.CalculateDiscountResponse, error) {
	return h.promotionService.CalculateDiscount(ctx, req)
}

func (h *PromotionServiceHandler) ClaimCoupon(ctx context.Context, req *promotionv1.ClaimCouponRequest) (*promotionv1.ClaimCouponResponse, error) {
	return h.promotionService.ClaimCoupon(ctx, req)
}

func (h *PromotionServiceHandler) ListUserCoupons(ctx context.Context, req *promotionv1.ListUserCouponsRequest) (*promotionv1.ListUserCouponsResponse, error) {
	return h.promotionService.ListUserCoupons(ctx, req)
}

func (h *PromotionServiceHandler) UseCoupon(ctx context.Context, req *promotionv1.UseCouponRequest) (*promotionv1.UseCouponResponse, error) {
	return h.promotionService.UseCoupon(ctx, req)
}
