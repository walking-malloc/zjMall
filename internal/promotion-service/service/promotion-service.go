package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	promotionv1 "zjMall/gen/go/api/proto/promotion"
	"zjMall/internal/promotion-service/model"
	"zjMall/internal/promotion-service/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PromotionService struct {
	promotionRepo repository.PromotionRepository
	couponRepo    repository.CouponRepository
}

func NewPromotionService(
	promotionRepo repository.PromotionRepository,
	couponRepo repository.CouponRepository,
) *PromotionService {
	return &PromotionService{
		promotionRepo: promotionRepo,
		couponRepo:    couponRepo,
	}
}

// CreatePromotion 创建促销活动
func (s *PromotionService) CreatePromotion(ctx context.Context, req *promotionv1.CreatePromotionRequest) (*promotionv1.CreatePromotionResponse, error) {
	// 参数校验
	validator := NewCreatePromotionRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &promotionv1.CreatePromotionResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	// 转换商品ID和类目ID为JSON
	productIDsJSON, _ := json.Marshal(req.ProductIds)
	categoryIDsJSON, _ := json.Marshal(req.CategoryIds)

	promotion := &model.Promotion{
		Name:           req.Name,
		Type:           int8(req.Type),
		Description:    req.Description,
		ProductIDs:     string(productIDsJSON),
		CategoryIDs:    string(categoryIDsJSON),
		ConditionValue: req.ConditionValue,
		DiscountValue:  req.DiscountValue,
		StartTime:      req.StartTime.AsTime(),
		EndTime:        req.EndTime.AsTime(),
		MaxUseTimes:    req.MaxUseTimes,
		TotalQuota:     req.TotalQuota,
		SortOrder:      req.SortOrder,
		Status:         model.PromotionStatusDraft,
	}

	if err := s.promotionRepo.CreatePromotion(ctx, promotion); err != nil {
		log.Printf("创建促销活动失败: %v", err)
		return &promotionv1.CreatePromotionResponse{
			Code:    1,
			Message: fmt.Sprintf("创建失败: %v", err),
		}, nil
	}

	return &promotionv1.CreatePromotionResponse{
		Code:        0,
		Message:     "创建成功",
		PromotionId: promotion.ID,
	}, nil
}

// GetPromotion 获取促销活动详情
func (s *PromotionService) GetPromotion(ctx context.Context, req *promotionv1.GetPromotionRequest) (*promotionv1.GetPromotionResponse, error) {
	promotion, err := s.promotionRepo.GetPromotionByID(ctx, req.PromotionId)
	if err != nil {
		return &promotionv1.GetPromotionResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}
	if promotion == nil {
		return &promotionv1.GetPromotionResponse{
			Code:    1,
			Message: "促销活动不存在",
		}, nil
	}

	return &promotionv1.GetPromotionResponse{
		Code:    0,
		Message: "查询成功",
		Data:    s.toPromotionInfo(promotion),
	}, nil
}

// ListPromotions 查询促销活动列表
func (s *PromotionService) ListPromotions(ctx context.Context, req *promotionv1.ListPromotionsRequest) (*promotionv1.ListPromotionsResponse, error) {
	filter := &repository.PromotionListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     int8(req.Type),
		Status:   int8(req.Status),
		Keyword:  req.Keyword,
		Offset:   int((req.Page - 1) * req.PageSize),
		Limit:    int(req.PageSize),
	}

	promotions, total, err := s.promotionRepo.ListPromotions(ctx, filter)
	if err != nil {
		return &promotionv1.ListPromotionsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	var promotionInfos []*promotionv1.PromotionInfo
	for _, p := range promotions {
		promotionInfos = append(promotionInfos, s.toPromotionInfo(p))
	}

	return &promotionv1.ListPromotionsResponse{
		Code:    0,
		Message: "查询成功",
		Data:    promotionInfos,
		Total:   total,
	}, nil
}

// GetAvailablePromotions 查询可用促销活动
func (s *PromotionService) GetAvailablePromotions(ctx context.Context, req *promotionv1.GetAvailablePromotionsRequest) (*promotionv1.GetAvailablePromotionsResponse, error) {
	promotions, err := s.promotionRepo.GetAvailablePromotions(ctx, req.ProductIds, req.CategoryId, req.TotalAmount)
	if err != nil {
		return &promotionv1.GetAvailablePromotionsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	var promotionInfos []*promotionv1.PromotionInfo
	for _, p := range promotions {
		// TODO: 这里需要你实现复杂的匹配逻辑：
		// 1. 检查商品ID是否匹配
		// 2. 检查类目ID是否匹配
		// 3. 检查金额是否满足条件
		// 4. 检查用户限购次数
		// 5. 检查总配额
		promotionInfos = append(promotionInfos, s.toPromotionInfo(p))
	}

	return &promotionv1.GetAvailablePromotionsResponse{
		Code:    0,
		Message: "查询成功",
		Data:    promotionInfos,
	}, nil
}

// CalculateDiscount 计算优惠金额
func (s *PromotionService) CalculateDiscount(ctx context.Context, req *promotionv1.CalculateDiscountRequest) (*promotionv1.CalculateDiscountResponse, error) {
	// TODO: 这里需要你实现复杂的优惠计算逻辑：
	// 1. 获取可用促销活动
	// 2. 计算促销优惠金额（满减、满折等）
	// 3. 如果有优惠券，计算优惠券优惠金额
	// 4. 处理优惠叠加规则
	// 5. 返回最终优惠明细

	totalAmount := req.TotalAmount
	promotionDiscount := 0.0
	couponDiscount := 0.0

	// 简化实现：这里只是框架，实际计算逻辑需要你来实现
	// 例如：满减计算、满折计算、优惠券计算等

	finalAmount := totalAmount - promotionDiscount - couponDiscount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &promotionv1.CalculateDiscountResponse{
		Code:    0,
		Message: "计算成功",
		Data: &promotionv1.DiscountDetail{
			TotalAmount:       totalAmount,
			PromotionDiscount: promotionDiscount,
			CouponDiscount:    couponDiscount,
			TotalDiscount:     promotionDiscount + couponDiscount,
			FinalAmount:       finalAmount,
		},
	}, nil
}

// toPromotionInfo 转换为 PromotionInfo
func (s *PromotionService) toPromotionInfo(p *model.Promotion) *promotionv1.PromotionInfo {
	var productIDs []string
	var categoryIDs []string
	json.Unmarshal([]byte(p.ProductIDs), &productIDs)
	json.Unmarshal([]byte(p.CategoryIDs), &categoryIDs)

	return &promotionv1.PromotionInfo{
		Id:             p.ID,
		Name:           p.Name,
		Type:           promotionv1.PromotionType(p.Type),
		Description:    p.Description,
		ProductIds:     productIDs,
		CategoryIds:    categoryIDs,
		ConditionValue: p.ConditionValue,
		DiscountValue:  p.DiscountValue,
		StartTime:      timestamppb.New(p.StartTime),
		EndTime:        timestamppb.New(p.EndTime),
		MaxUseTimes:    p.MaxUseTimes,
		TotalQuota:     p.TotalQuota,
		UsedQuota:      p.UsedQuota,
		SortOrder:      p.SortOrder,
		Status:         promotionv1.PromotionStatus(p.Status),
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}
}
