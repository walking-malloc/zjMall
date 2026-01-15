package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	promotionv1 "zjMall/gen/go/api/proto/promotion"
	"zjMall/internal/promotion-service/model"
	"zjMall/internal/promotion-service/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ClaimCoupon 领取优惠券
func (s *PromotionService) ClaimCoupon(ctx context.Context, req *promotionv1.ClaimCouponRequest) (*promotionv1.ClaimCouponResponse, error) {
	// 获取优惠券模板
	template, err := s.couponRepo.GetCouponTemplateByID(ctx, req.TemplateId)
	if err != nil || template == nil {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: "优惠券模板不存在",
		}, nil
	}

	// 检查模板状态
	if template.Status != 1 {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: "优惠券模板已停用",
		}, nil
	}

	// 检查是否在有效期内
	now := time.Now()
	if now.Before(template.ValidStartTime) || now.After(template.ValidEndTime) {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: "优惠券不在有效期内",
		}, nil
	}

	// 检查总配额
	if template.TotalCount > 0 && template.ClaimedCount >= template.TotalCount {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: "优惠券已领完",
		}, nil
	}

	// 检查用户限领数量
	userCount, err := s.couponRepo.GetUserCouponCountByTemplate(ctx, req.UserId, req.TemplateId)
	if err != nil {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}
	if userCount >= template.PerUserLimit {
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: "已达到限领数量",
		}, nil
	}

	// 计算优惠券有效期
	validStartTime := template.ValidStartTime
	validEndTime := template.ValidEndTime
	if template.ValidDays > 0 {
		validStartTime = now
		validEndTime = now.Add(time.Duration(template.ValidDays) * 24 * time.Hour)
	}

	// 创建优惠券实例
	coupon := &model.Coupon{
		TemplateID:     template.ID,
		UserID:         req.UserId,
		Name:           template.Name,
		Type:           template.Type,
		Description:    template.Description,
		DiscountValue:  template.DiscountValue,
		ConditionValue: template.ConditionValue,
		Status:         model.CouponStatusUnused,
		ValidStartTime: validStartTime,
		ValidEndTime:   validEndTime,
	}

	if err := s.couponRepo.CreateCoupon(ctx, coupon); err != nil {
		log.Printf("创建优惠券失败: %v", err)
		return &promotionv1.ClaimCouponResponse{
			Code:    1,
			Message: fmt.Sprintf("领取失败: %v", err),
		}, nil
	}

	// 更新模板已领取数量
	if err := s.couponRepo.IncrementTemplateClaimedCount(ctx, template.ID); err != nil {
		log.Printf("更新模板已领取数量失败: %v", err)
	}

	return &promotionv1.ClaimCouponResponse{
		Code:     0,
		Message:  "领取成功",
		CouponId: coupon.ID,
	}, nil
}

// ListUserCoupons 查询用户优惠券列表
func (s *PromotionService) ListUserCoupons(ctx context.Context, req *promotionv1.ListUserCouponsRequest) (*promotionv1.ListUserCouponsResponse, error) {
	filter := &repository.CouponListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		UserID:   req.UserId,
		Status:   int8(req.Status),
		Offset:   int((req.Page - 1) * req.PageSize),
		Limit:    int(req.PageSize),
	}

	coupons, total, err := s.couponRepo.ListUserCoupons(ctx, filter)
	if err != nil {
		return &promotionv1.ListUserCouponsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	// 检查并更新过期优惠券状态
	now := time.Now()
	var couponInfos []*promotionv1.CouponInfo
	for _, coupon := range coupons {
		// 如果优惠券已过期且未使用，更新状态
		if coupon.Status == model.CouponStatusUnused && now.After(coupon.ValidEndTime) {
			coupon.Status = model.CouponStatusExpired
			s.couponRepo.UpdateCoupon(ctx, coupon)
		}
		couponInfos = append(couponInfos, s.toCouponInfo(coupon))
	}

	return &promotionv1.ListUserCouponsResponse{
		Code:    0,
		Message: "查询成功",
		Data:    couponInfos,
		Total:   total,
	}, nil
}

// UseCoupon 核销优惠券
func (s *PromotionService) UseCoupon(ctx context.Context, req *promotionv1.UseCouponRequest) (*promotionv1.UseCouponResponse, error) {
	// TODO: 这里需要你实现核销逻辑：
	// 1. 检查优惠券是否存在
	// 2. 检查优惠券是否属于该用户
	// 3. 检查优惠券状态（未使用）
	// 4. 检查优惠券是否在有效期内
	// 5. 检查订单金额是否满足使用条件
	// 6. 更新优惠券状态为已使用
	// 7. 记录使用信息

	coupon, err := s.couponRepo.GetCouponByID(ctx, req.CouponId)
	if err != nil || coupon == nil {
		return &promotionv1.UseCouponResponse{
			Code:    1,
			Message: "优惠券不存在",
		}, nil
	}

	if coupon.UserID != req.UserId {
		return &promotionv1.UseCouponResponse{
			Code:    1,
			Message: "优惠券不属于该用户",
		}, nil
	}

	if coupon.Status != model.CouponStatusUnused {
		return &promotionv1.UseCouponResponse{
			Code:    1,
			Message: "优惠券已使用或已过期",
		}, nil
	}

	now := time.Now()
	if now.Before(coupon.ValidStartTime) || now.After(coupon.ValidEndTime) {
		return &promotionv1.UseCouponResponse{
			Code:    1,
			Message: "优惠券不在有效期内",
		}, nil
	}

	// 检查使用条件
	if coupon.ConditionValue != "" {
		conditionAmount, err := strconv.ParseFloat(coupon.ConditionValue, 64)
		if err == nil && req.OrderAmount < conditionAmount {
			return &promotionv1.UseCouponResponse{
				Code:    1,
				Message: fmt.Sprintf("订单金额未满足使用条件（需满%.2f元）", conditionAmount),
			}, nil
		}
	}

	// 更新优惠券状态
	nowTime := time.Now()
	coupon.Status = model.CouponStatusUsed
	coupon.UsedAt = &nowTime
	coupon.OrderID = req.OrderId

	if err := s.couponRepo.UpdateCoupon(ctx, coupon); err != nil {
		log.Printf("核销优惠券失败: %v", err)
		return &promotionv1.UseCouponResponse{
			Code:    1,
			Message: fmt.Sprintf("核销失败: %v", err),
		}, nil
	}

	return &promotionv1.UseCouponResponse{
		Code:    0,
		Message: "核销成功",
	}, nil
}

// toCouponInfo 转换为 CouponInfo
func (s *PromotionService) toCouponInfo(c *model.Coupon) *promotionv1.CouponInfo {
	couponInfo := &promotionv1.CouponInfo{
		Id:             c.ID,
		TemplateId:     c.TemplateID,
		Name:           c.Name,
		Type:           promotionv1.CouponType(c.Type),
		Description:    c.Description,
		DiscountValue:  c.DiscountValue,
		ConditionValue: c.ConditionValue,
		UserId:         c.UserID,
		Status:         promotionv1.CouponStatus(c.Status),
		ValidStartTime: timestamppb.New(c.ValidStartTime),
		ValidEndTime:   timestamppb.New(c.ValidEndTime),
		CreatedAt:      timestamppb.New(c.CreatedAt),
	}

	if c.UsedAt != nil {
		couponInfo.UsedAt = timestamppb.New(*c.UsedAt)
	}
	if c.OrderID != "" {
		couponInfo.OrderId = c.OrderID
	}

	return couponInfo
}
