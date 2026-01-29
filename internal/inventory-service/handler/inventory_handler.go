package handler

import (
	"context"
	"fmt"
	"log"

	inventoryv1 "zjMall/gen/go/api/proto/inventory"
	"zjMall/internal/inventory-service/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// InventoryHandler 是库存服务对外的 gRPC 入口
type InventoryHandler struct {
	inventoryv1.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

// NewInventoryHandler 创建库存 Handler
func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// GetStock 查询单个 SKU 库存
func (h *InventoryHandler) GetStock(ctx context.Context, req *inventoryv1.GetStockRequest) (*inventoryv1.GetStockResponse, error) {
	if req.SkuId == "" {
		return &inventoryv1.GetStockResponse{
			Code:    1,
			Message: "sku_id 不能为空",
		}, nil
	}

	stock, err := h.svc.GetStock(ctx, req.SkuId)
	if err != nil {
		log.Printf("❌ [InventoryHandler] GetStock: 查询失败 sku_id=%s, err=%v", req.SkuId, err)
		return &inventoryv1.GetStockResponse{
			Code:    1,
			Message: fmt.Sprintf("查询库存失败: %v", err),
		}, nil
	}

	if stock == nil {
		// 不存在时返回 code=0, data 为空，方便调用方判断
		return &inventoryv1.GetStockResponse{
			Code:    0,
			Message: "success",
		}, nil
	}

	return &inventoryv1.GetStockResponse{
		Code:    0,
		Message: "success",
		Data: &inventoryv1.Stock{
			Id:             stock.ID,
			SkuId:          stock.SKUID,
			AvailableStock: stock.AvailableStock,
			Version:        stock.Version,
			CreatedAt:      timestamppb.New(stock.CreatedAt),
			UpdatedAt:      timestamppb.New(stock.UpdatedAt),
		},
	}, nil
}

// BatchGetStock 批量查询多个 SKU 库存
func (h *InventoryHandler) BatchGetStock(ctx context.Context, req *inventoryv1.BatchGetStockRequest) (*inventoryv1.BatchGetStockResponse, error) {
	if len(req.SkuIds) == 0 {
		return &inventoryv1.BatchGetStockResponse{
			Code:    0,
			Message: "success",
			Data:    map[string]*inventoryv1.Stock{},
		}, nil
	}

	stocks, err := h.svc.BatchGetStock(ctx, req.SkuIds)
	if err != nil {
		log.Printf("❌ [InventoryHandler] BatchGetStock: 批量查询失败 sku_ids=%v, err=%v", req.SkuIds, err)
		return &inventoryv1.BatchGetStockResponse{
			Code:    1,
			Message: fmt.Sprintf("批量查询库存失败: %v", err),
		}, nil
	}

	result := make(map[string]*inventoryv1.Stock, len(stocks))
	for skuID, s := range stocks {
		if s == nil {
			continue
		}
		result[skuID] = &inventoryv1.Stock{
			Id:             s.ID,
			SkuId:          s.SKUID,
			AvailableStock: s.AvailableStock,
			Version:        s.Version,
			CreatedAt:      timestamppb.New(s.CreatedAt),
			UpdatedAt:      timestamppb.New(s.UpdatedAt),
		}
	}

	return &inventoryv1.BatchGetStockResponse{
		Code:    0,
		Message: "success",
		Data:    result,
	}, nil
}

// DeductStock 扣减库存（通常由订单服务调用）
func (h *InventoryHandler) DeductStock(ctx context.Context, req *inventoryv1.DeductStockRequest) (*inventoryv1.DeductStockResponse, error) {
	if len(req.Items) == 0 {
		return &inventoryv1.DeductStockResponse{
			Code:    1,
			Message: "items 不能为空",
		}, nil
	}

	items := make([]service.ItemQuantity, 0, len(req.Items))
	for _, it := range req.Items {
		if it.SkuId == "" || it.Quantity <= 0 {
			return &inventoryv1.DeductStockResponse{
				Code:    1,
				Message: "sku_id 不能为空且 quantity 必须大于 0",
			}, nil
		}
		items = append(items, service.ItemQuantity{
			SKUID:    it.SkuId,
			Quantity: it.Quantity,
		})
	}

	if err := h.svc.TryDeductStocks(ctx, req.OrderId, items); err != nil {
		log.Printf("❌ [InventoryHandler] DeductStock: 扣减失败 order_id=%s, err=%v", req.OrderId, err)
		return &inventoryv1.DeductStockResponse{
			Code:    1,
			Message: fmt.Sprintf("扣减库存失败: %v", err),
		}, nil
	}

	return &inventoryv1.DeductStockResponse{
		Code:    0,
		Message: "success",
	}, nil
}

// RollbackStock 回滚库存（订单取消/关闭时调用）
func (h *InventoryHandler) RollbackStock(ctx context.Context, req *inventoryv1.RollbackStockRequest) (*inventoryv1.RollbackStockResponse, error) {
	if len(req.Items) == 0 {
		return &inventoryv1.RollbackStockResponse{
			Code:    1,
			Message: "items 不能为空",
		}, nil
	}

	items := make([]service.ItemQuantity, 0, len(req.Items))
	for _, it := range req.Items {
		if it.SkuId == "" || it.Quantity <= 0 {
			continue
		}
		items = append(items, service.ItemQuantity{
			SKUID:    it.SkuId,
			Quantity: it.Quantity,
		})
	}

	if err := h.svc.RollbackStocks(ctx, req.OrderId, items); err != nil {
		log.Printf("❌ [InventoryHandler] RollbackStock: 回滚失败 order_id=%s, err=%v", req.GetOrderId(), err)
		return &inventoryv1.RollbackStockResponse{
			Code:    1,
			Message: fmt.Sprintf("回滚库存失败: %v", err),
		}, nil
	}

	return &inventoryv1.RollbackStockResponse{
		Code:    0,
		Message: "success",
	}, nil
}
