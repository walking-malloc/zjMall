package handler

import (
	"context"
	"fmt"
	"log"
	"strings"
	cartv1 "zjMall/gen/go/api/proto/cart"
	"zjMall/internal/cart-service/service"
	"zjMall/internal/common/middleware"
	"zjMall/pkg"

	"google.golang.org/grpc/metadata"
)

type CartServiceHandler struct {
	cartv1.UnimplementedCartServiceServer
	cartService *service.CartService
}

func NewCartServiceHandler(cartService *service.CartService) *CartServiceHandler {
	return &CartServiceHandler{
		cartService: cartService,
	}
}

// getUserID 从上下文中获取用户 ID
// 优先使用 HTTP 层中间件注入的 user_id；如果没有，则从 gRPC Metadata 的 Authorization 中解析 JWT
func getUserID(ctx context.Context) string {
	// 1. 优先从 HTTP 中间件注入的 Context 中获取
	if userID := middleware.GetUserIDFromContext(ctx); userID != "" {
		return userID
	}

	// 2. 从 gRPC Metadata 中获取 Authorization 头
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	authVals := md.Get("authorization")
	if len(authVals) == 0 {
		return ""
	}

	authHeader := authVals[0]
	if authHeader == "" {
		return ""
	}

	// 去掉前缀 "Bearer "
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return ""
	}

	// 验证 JWT 并解析出 userID
	userID, err := pkg.VerifyJWT(token)
	if err != nil {
		log.Printf("⚠️ [Handler] JWT 验证失败: %v", err)
		return ""
	}
	return userID
}

// AddItem 添加商品到购物车
func (h *CartServiceHandler) AddItem(ctx context.Context, req *cartv1.AddItemRequest) (*cartv1.AddItemResponse, error) {
	// 从 context 中获取用户ID（由认证中间件注入）
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] AddItem: 用户未登录")
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}
	if req.ProductId == "" || req.SkuId == "" {
		log.Printf("⚠️ [Handler] AddItem: 参数校验失败 - product_id=%s, sku_id=%s", req.ProductId, req.SkuId)
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: "商品ID或SKU ID不能为空",
		}, nil
	}
	if req.Quantity <= 0 {
		log.Printf("⚠️ [Handler] AddItem: 数量校验失败 - quantity=%d", req.Quantity)
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: "数量必须大于0",
		}, nil
	}

	resp, err := h.cartService.AddItem(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] AddItem: Service 层返回错误: %v", err)
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: fmt.Sprintf("添加购物车失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] AddItem: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateItemQuantity 更新购物车商品数量
func (h *CartServiceHandler) UpdateItemQuantity(ctx context.Context, req *cartv1.UpdateItemQuantityRequest) (*cartv1.UpdateItemQuantityResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] UpdateItemQuantity: 用户未登录")
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	// 参数校验
	if req.ItemId == "" {
		log.Printf("⚠️ [Handler] UpdateItemQuantity: 参数校验失败 - item_id为空")
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: "购物车项ID不能为空",
		}, nil
	}
	if req.Quantity <= 0 {
		log.Printf("⚠️ [Handler] UpdateItemQuantity: 数量校验失败 - quantity=%d", req.Quantity)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: "数量必须大于0",
		}, nil
	}

	resp, err := h.cartService.UpdateItemQuantity(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] UpdateItemQuantity: Service 层返回错误: %v", err)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: fmt.Sprintf("更新数量失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] UpdateItemQuantity: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// RemoveItem 删除购物车商品
func (h *CartServiceHandler) RemoveItem(ctx context.Context, req *cartv1.RemoveItemRequest) (*cartv1.RemoveItemResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] RemoveItem: 用户未登录")
		return &cartv1.RemoveItemResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	// 参数校验
	if req.ItemId == "" {
		log.Printf("⚠️ [Handler] RemoveItem: 参数校验失败 - item_id为空")
		return &cartv1.RemoveItemResponse{
			Code:    1,
			Message: "购物车项ID不能为空",
		}, nil
	}

	resp, err := h.cartService.RemoveItem(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] RemoveItem: Service 层返回错误: %v", err)
		return &cartv1.RemoveItemResponse{
			Code:    1,
			Message: fmt.Sprintf("删除失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] RemoveItem: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// RemoveItems 批量删除购物车商品
func (h *CartServiceHandler) RemoveItems(ctx context.Context, req *cartv1.RemoveItemsRequest) (*cartv1.RemoveItemsResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] RemoveItems: 用户未登录")
		return &cartv1.RemoveItemsResponse{
			Code:         1,
			Message:      "用户未登录",
			DeletedCount: 0,
		}, nil
	}

	// 参数校验
	if len(req.ItemIds) == 0 {
		log.Printf("⚠️ [Handler] RemoveItems: 参数校验失败 - item_ids为空")
		return &cartv1.RemoveItemsResponse{
			Code:         1,
			Message:      "请选择要删除的商品",
			DeletedCount: 0,
		}, nil
	}

	resp, err := h.cartService.RemoveItems(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] RemoveItems: Service 层返回错误: %v", err)
		return &cartv1.RemoveItemsResponse{
			Code:         1,
			Message:      fmt.Sprintf("批量删除失败: %v", err),
			DeletedCount: 0,
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] RemoveItems: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// ClearCart 清空购物车
func (h *CartServiceHandler) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*cartv1.ClearCartResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] ClearCart: 用户未登录")
		return &cartv1.ClearCartResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	resp, err := h.cartService.ClearCart(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] ClearCart: Service 层返回错误: %v", err)
		return &cartv1.ClearCartResponse{
			Code:    1,
			Message: fmt.Sprintf("清空购物车失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] ClearCart: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// GetCart 获取购物车列表
func (h *CartServiceHandler) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] GetCart: 用户未登录")
		return &cartv1.GetCartResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	resp, err := h.cartService.GetCart(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] GetCart: Service 层返回错误: %v", err)
		return &cartv1.GetCartResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] GetCart: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// GetCartSummary 获取购物车统计信息
func (h *CartServiceHandler) GetCartSummary(ctx context.Context, req *cartv1.GetCartSummaryRequest) (*cartv1.GetCartSummaryResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] GetCartSummary: 用户未登录")
		return &cartv1.GetCartSummaryResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	resp, err := h.cartService.GetCartSummary(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] GetCartSummary: Service 层返回错误: %v", err)
		return &cartv1.GetCartSummaryResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车统计失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] GetCartSummary: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}

// CheckoutPreview 结算预览
func (h *CartServiceHandler) CheckoutPreview(ctx context.Context, req *cartv1.CheckoutPreviewRequest) (*cartv1.CheckoutPreviewResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		log.Printf("⚠️ [Handler] CheckoutPreview: 用户未登录")
		return &cartv1.CheckoutPreviewResponse{
			Code:    1,
			Message: "用户未登录",
		}, nil
	}

	resp, err := h.cartService.CheckoutPreview(ctx, req, userID)
	if err != nil {
		log.Printf("❌ [Handler] CheckoutPreview: Service 层返回错误: %v", err)
		return &cartv1.CheckoutPreviewResponse{
			Code:    1,
			Message: fmt.Sprintf("结算预览失败: %v", err),
		}, nil
	}
	if resp.Code != 0 {
		log.Printf("⚠️ [Handler] CheckoutPreview: Service 层返回业务错误 - code=%d, message=%s", resp.Code, resp.Message)
	}
	return resp, nil
}
