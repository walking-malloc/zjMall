package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	cartv1 "zjMall/gen/go/api/proto/cart"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/internal/cart-service/model"
	"zjMall/internal/cart-service/repository"
	"zjMall/internal/common/client"
	productRepository "zjMall/internal/product-service/repository"
	"zjMall/pkg"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// CartService 购物车服务（业务逻辑层）
type CartService struct {
	cartRepo        repository.CartRepository
	productClient   client.ProductClient   // 商品服务客户端（用于查询商品信息）
	inventoryClient client.InventoryClient // 库存服务客户端（用于库存校验）
	// TODO: 添加促销服务客户端（用于计算优惠）
	// promotionClient promotionv1.PromotionServiceClient
}

// NewCartService 创建购物车服务实例
func NewCartService(cartRepo repository.CartRepository, productClient client.ProductClient, inventoryClient client.InventoryClient) *CartService {
	return &CartService{
		cartRepo:        cartRepo,
		productClient:   productClient,
		inventoryClient: inventoryClient,
	}
}

// AddItem 添加商品到购物车
func (s *CartService) AddItem(ctx context.Context, req *cartv1.AddItemRequest, userID string) (*cartv1.AddItemResponse, error) {
	// 从商品服务获取商品信息（库存由独立库存服务管理，这里不做扣减）
	var productTitle, productImage, skuName string
	var price float64
	var stock int64

	if s.productClient != nil {
		product, skus, err := s.productClient.GetProduct(ctx, req.ProductId)
		if err != nil {
			log.Printf("❌ [Service] AddItem: 获取商品信息失败 - product_id=%s, error=%v", req.ProductId, err)
			return &cartv1.AddItemResponse{
				Code:    1,
				Message: fmt.Sprintf("获取商品信息失败: %v", err),
			}, nil
		}

		// 找到匹配的sku
		var sku *productv1.SkuInfo
		for _, item := range skus {
			if item.Id == req.SkuId {
				sku = item
				break
			}
		}

		if sku == nil {
			log.Printf("⚠️ [Service] AddItem: SKU不存在 - product_id=%s, sku_id=%s", req.ProductId, req.SkuId)
			return &cartv1.AddItemResponse{
				Code:    1,
				Message: "SKU不存在",
			}, nil
		}

		// 检查库存
		if s.inventoryClient != nil {
			stock, err = s.inventoryClient.GetStock(ctx, req.SkuId)
			if err != nil {
				log.Printf("❌ [Service] AddItem: 获取库存失败 - sku_id=%s, error=%v", req.SkuId, err)
				return &cartv1.AddItemResponse{
					Code:    1,
					Message: fmt.Sprintf("获取库存失败: %v", err),
				}, nil
			}
			if stock <= 0 {
				log.Printf("⚠️ [Service] AddItem: 库存不足 - sku_id=%s, stock=%d", req.SkuId, stock)
				return &cartv1.AddItemResponse{
					Code:    1,
					Message: fmt.Sprintf("库存不足: %s", sku.Name),
				}, nil
			}
		}

		// 提取商品信息
		productTitle = product.Title
		productImage = product.MainImage
		if len(product.Images) > 0 {
			productImage = product.Images[0]
		}
		skuName = sku.Name
		price = sku.Price // sku.Price 已经是 float64 类型
	} else {
		// 降级：使用模拟数据（商品服务未配置或不可用）
		productTitle = "商品名称"
		productImage = ""
		skuName = "默认规格"
		price = 99999.00
	}

	// 检查购物车中是否已存在相同 SKU（使用数据库唯一索引优化查询）
	existingItem, err := s.cartRepo.GetCartItemByUserAndSKU(ctx, userID, req.SkuId)
	if err != nil {
		log.Printf("❌ [Service] AddItem: 查询购物车失败 - user_id=%s, sku_id=%s, error=%v", userID, req.SkuId, err)
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: fmt.Sprintf("查询购物车失败: %v", err),
		}, nil
	}

	if existingItem != nil {
		// 已存在，累加数量（库存校验放在结算/下单阶段由库存服务负责）
		newQuantity := existingItem.Quantity + req.Quantity

		if err := s.cartRepo.UpdateItemQuantity(ctx, userID, existingItem.ID, newQuantity); err != nil {
			log.Printf("❌ [Service] AddItem: 更新购物车失败 - user_id=%s, item_id=%s, error=%v", userID, existingItem.ID, err)
			return &cartv1.AddItemResponse{
				Code:    1,
				Message: fmt.Sprintf("更新购物车失败: %v", err),
			}, nil
		}

		return &cartv1.AddItemResponse{
			Code:    0,
			Message: "添加成功",
			Data:    convertCartItemToProto(existingItem),
		}, nil
	}

	// 不存在，创建新项
	item := &model.CartItem{
		BaseModel: pkg.BaseModel{
			ID: pkg.GenerateULID(),
		},
		UserID:       userID,
		ProductID:    req.ProductId,
		SKUID:        req.SkuId,
		ProductTitle: productTitle,
		ProductImage: productImage,
		SKUName:      skuName,
		Price:        price, // float64
		CurrentPrice: price, // float64
		Quantity:     req.Quantity,
		// Stock 字段仅用于展示，可在结算或下单前通过库存服务刷新
		Stock:   int32(stock),
		IsValid: true,
	}

	if err := s.cartRepo.AddItem(ctx, userID, item); err != nil {
		log.Printf("❌ [Service] AddItem: 添加购物车失败 - user_id=%s, product_id=%s, sku_id=%s, error=%v", userID, req.ProductId, req.SkuId, err)
		return &cartv1.AddItemResponse{
			Code:    1,
			Message: fmt.Sprintf("添加购物车失败: %v", err),
		}, nil
	}

	return &cartv1.AddItemResponse{
		Code:    0,
		Message: "添加成功",
		Data:    convertCartItemToProto(item),
	}, nil
}

// UpdateItemQuantity 更新购物车商品数量
// 注意：参数校验已在 Handler 层完成，这里只处理业务逻辑
func (s *CartService) UpdateItemQuantity(ctx context.Context, req *cartv1.UpdateItemQuantityRequest, userID string) (*cartv1.UpdateItemQuantityResponse, error) {
	// 获取购物车项
	item, err := s.cartRepo.GetCartItem(ctx, userID, req.ItemId)
	if err != nil {
		log.Printf("❌ [Service] UpdateItemQuantity: 获取购物车项失败 - user_id=%s, item_id=%s, error=%v", userID, req.ItemId, err)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车项失败: %v", err),
		}, nil
	}
	if item == nil {
		log.Printf("⚠️ [Service] UpdateItemQuantity: 购物车项不存在 - user_id=%s, item_id=%s", userID, req.ItemId)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: "购物车项不存在",
		}, nil
	}

	// 注意：这里只检查库存是否充足，不扣减库存
	// 库存扣减应该在订单创建时预占，支付成功后扣减
	if s.inventoryClient != nil {
		stock, err := s.inventoryClient.GetStock(ctx, item.SKUID)
		if err != nil {
			log.Printf("❌ [Service] UpdateItemQuantity: 获取库存失败 - sku_id=%s, error=%v", item.SKUID, err)
			return &cartv1.UpdateItemQuantityResponse{
				Code:    1,
				Message: fmt.Sprintf("获取库存失败: %v", err),
			}, nil
		}
		// 检查新数量是否超过当前可用库存
		if int64(req.Quantity) > stock {
			log.Printf("⚠️ [Service] UpdateItemQuantity: 库存不足 - sku_id=%s, 请求数量=%d, 可用库存=%d", item.SKUID, req.Quantity, stock)
			return &cartv1.UpdateItemQuantityResponse{
				Code:    1,
				Message: fmt.Sprintf("%s库存不足，当前可用库存：%d", item.SKUName, stock),
			}, nil
		}
	}

	// 更新数量
	if err := s.cartRepo.UpdateItemQuantity(ctx, userID, req.ItemId, req.Quantity); err != nil {
		log.Printf("❌ [Service] UpdateItemQuantity: 更新数量失败 - user_id=%s, item_id=%s, quantity=%d, error=%v", userID, req.ItemId, req.Quantity, err)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: fmt.Sprintf("更新数量失败: %v", err),
		}, nil
	}

	// 重新获取更新后的项
	updatedItem, err := s.cartRepo.GetCartItem(ctx, userID, req.ItemId)
	if err != nil {
		log.Printf("❌ [Service] UpdateItemQuantity: 获取更新后的购物车项失败 - user_id=%s, item_id=%s, error=%v", userID, req.ItemId, err)
		return &cartv1.UpdateItemQuantityResponse{
			Code:    1,
			Message: fmt.Sprintf("获取更新后的购物车项失败: %v", err),
		}, nil
	}

	return &cartv1.UpdateItemQuantityResponse{
		Code:    0,
		Message: "更新成功",
		Data:    convertCartItemToProto(updatedItem),
	}, nil
}

// RemoveItem 删除购物车商品
// 注意：参数校验已在 Handler 层完成，这里只处理业务逻辑
func (s *CartService) RemoveItem(ctx context.Context, req *cartv1.RemoveItemRequest, userID string) (*cartv1.RemoveItemResponse, error) {
	if err := s.cartRepo.RemoveItem(ctx, userID, req.ItemId); err != nil {
		log.Printf("❌ [Service] RemoveItem: 删除失败 - user_id=%s, item_id=%s, error=%v", userID, req.ItemId, err)
		return &cartv1.RemoveItemResponse{
			Code:    1,
			Message: fmt.Sprintf("删除失败: %v", err),
		}, nil
	}

	return &cartv1.RemoveItemResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// RemoveItems 批量删除购物车商品
// 注意：参数校验已在 Handler 层完成，这里只处理业务逻辑
func (s *CartService) RemoveItems(ctx context.Context, req *cartv1.RemoveItemsRequest, userID string) (*cartv1.RemoveItemsResponse, error) {
	if err := s.cartRepo.RemoveItems(ctx, userID, req.ItemIds); err != nil {
		log.Printf("❌ [Service] RemoveItems: 批量删除失败 - user_id=%s, item_ids=%v, error=%v", userID, req.ItemIds, err)
		return &cartv1.RemoveItemsResponse{
			Code:         1,
			Message:      fmt.Sprintf("批量删除失败: %v", err),
			DeletedCount: 0,
		}, nil
	}

	// 统计实际删除的数量（去重后）
	uniqueCount := make(map[string]bool)
	for _, id := range req.ItemIds {
		uniqueCount[id] = true
	}

	return &cartv1.RemoveItemsResponse{
		Code:         0,
		Message:      "删除成功",
		DeletedCount: int32(len(uniqueCount)), // 返回去重后的数量
	}, nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest, userID string) (*cartv1.ClearCartResponse, error) {
	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		log.Printf("❌ [Service] ClearCart: 清空购物车失败 - user_id=%s, error=%v", userID, err)
		return &cartv1.ClearCartResponse{
			Code:    1,
			Message: fmt.Sprintf("清空购物车失败: %v", err),
		}, nil
	}

	return &cartv1.ClearCartResponse{
		Code:    0,
		Message: "清空成功",
	}, nil
}

// GetCart 获取购物车列表
func (s *CartService) GetCart(ctx context.Context, req *cartv1.GetCartRequest, userID string) (*cartv1.GetCartResponse, error) {
	items, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		log.Printf("❌ [Service] GetCart: 获取购物车失败 - user_id=%s, error=%v", userID, err)
		return &cartv1.GetCartResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车失败: %v", err),
		}, nil
	}

	// 转换为 Proto 格式
	protoItems := make([]*cartv1.CartItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, convertCartItemToProto(item))
	}

	// 计算统计信息
	summary := s.calculateSummary(items)

	return &cartv1.GetCartResponse{
		Code:    0,
		Message: "查询成功",
		Items:   protoItems,
		Summary: summary,
	}, nil
}

// GetCartSummary 获取购物车统计信息
func (s *CartService) GetCartSummary(ctx context.Context, req *cartv1.GetCartSummaryRequest, userID string) (*cartv1.GetCartSummaryResponse, error) {
	items, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		log.Printf("❌ [Service] GetCartSummary: 获取购物车失败 - user_id=%s, error=%v", userID, err)
		return &cartv1.GetCartSummaryResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车失败: %v", err),
		}, nil
	}

	summary := s.calculateSummary(items)

	return &cartv1.GetCartSummaryResponse{
		Code:    0,
		Message: "查询成功",
		Data:    summary,
	}, nil
}

// CheckoutPreview 结算预览（计算价格和优惠）
func (s *CartService) CheckoutPreview(ctx context.Context, req *cartv1.CheckoutPreviewRequest, userID string) (*cartv1.CheckoutPreviewResponse, error) {
	// 获取购物车所有商品
	allItems, err := s.cartRepo.GetCartItems(ctx, userID)
	if err != nil {
		log.Printf("❌ [Service] CheckoutPreview: 获取购物车失败 - user_id=%s, error=%v", userID, err)
		return &cartv1.CheckoutPreviewResponse{
			Code:    1,
			Message: fmt.Sprintf("获取购物车失败: %v", err),
		}, nil
	}

	// 筛选选中的商品
	var selectedItems []*model.CartItem
	if len(req.ItemIds) == 0 {
		// 未指定，选择所有有效商品
		for _, item := range allItems {
			if item.IsValid {
				selectedItems = append(selectedItems, item)
			}
		}
	} else {
		// 选择指定的商品
		itemMap := make(map[string]*model.CartItem)
		for _, item := range allItems {
			itemMap[item.ID] = item
		}
		for _, itemID := range req.ItemIds {
			if item, ok := itemMap[itemID]; ok {
				selectedItems = append(selectedItems, item)
			}
		}
	}

	if len(selectedItems) == 0 {
		log.Printf("⚠️ [Service] CheckoutPreview: 没有选中商品 - user_id=%s, item_ids=%v", userID, req.ItemIds)
		return &cartv1.CheckoutPreviewResponse{
			Code:    1,
			Message: "请选择要结算的商品",
		}, nil
	}

	// 实时更新商品信息（价格、库存、状态）
	// 使用并发调用优化性能，避免串行等待
	s.updateProductInfoForCheckout(ctx, selectedItems)

	// TODO: 调用促销服务，计算促销优惠
	// TODO: 调用用户服务，获取配送地址
	// TODO: 计算运费

	// 计算商品总金额（原价）
	productTotal := s.calculateProductTotal(selectedItems)

	// TODO: 计算促销优惠（调用促销服务）
	// promotionDiscount, promotions, err := s.calculatePromotionDiscount(ctx, selectedItems, productTotal)
	// if err != nil {
	//     log.Printf("计算促销优惠失败: %v", err)
	// }
	promotionDiscount := "0.00"             // 临时值
	promotions := []*cartv1.PromotionInfo{} // 临时值

	// TODO: 计算优惠券优惠（调用促销服务）
	// couponDiscount := "0.00"
	// var coupon *cartv1.CouponInfo
	// if req.CouponId != "" {
	//     couponDiscount, coupon, err = s.calculateCouponDiscount(ctx, userID, req.CouponId, productTotal, promotionDiscount)
	//     if err != nil {
	//         log.Printf("计算优惠券优惠失败: %v", err)
	//     }
	// }
	couponDiscount := "0.00"      // 临时值
	var coupon *cartv1.CouponInfo // 临时值

	// TODO: 计算运费（调用物流服务或根据规则计算）
	shippingFee := "0.00" // 临时值

	// 计算最终实付金额
	finalAmount := s.calculateFinalAmount(productTotal, promotionDiscount, couponDiscount, shippingFee)

	// TODO: 获取配送地址信息
	// address, err := s.getAddressInfo(ctx, userID, req.AddressId)
	// if err != nil {
	//     log.Printf("获取地址信息失败: %v", err)
	// }
	var address *cartv1.AddressInfo // 临时值

	// 转换为 Proto 格式
	protoItems := make([]*cartv1.CartItem, 0, len(selectedItems))
	for _, item := range selectedItems {
		protoItems = append(protoItems, convertCartItemToProto(item))
	}

	return &cartv1.CheckoutPreviewResponse{
		Code:    0,
		Message: "查询成功",
		Data: &cartv1.CheckoutPreviewData{
			Items:             protoItems,
			ProductTotal:      productTotal,
			PromotionDiscount: promotionDiscount,
			CouponDiscount:    couponDiscount,
			ShippingFee:       shippingFee,
			FinalAmount:       finalAmount,
			Promotions:        promotions,
			Coupon:            coupon,
			Address:           address,
		},
	}, nil
}

// ============================================
// 辅助函数
// ============================================

// convertCartItemToProto 转换购物车项为 Proto 格式
func convertCartItemToProto(item *model.CartItem) *cartv1.CartItem {
	return &cartv1.CartItem{
		Id:            item.ID,
		UserId:        item.UserID,
		ProductId:     item.ProductID,
		SkuId:         item.SKUID,
		ProductTitle:  item.ProductTitle,
		ProductImage:  item.ProductImage,
		SkuName:       item.SKUName,
		Price:         item.PriceString(),        // 转换为字符串
		CurrentPrice:  item.CurrentPriceString(), // 转换为字符串
		Quantity:      item.Quantity,
		Stock:         item.Stock,
		IsValid:       item.IsValid,
		InvalidReason: item.InvalidReason,
		CreatedAt:     timestamppb.New(item.CreatedAt),
		UpdatedAt:     timestamppb.New(item.UpdatedAt),
	}
}

// calculateSummary 计算购物车统计信息
func (s *CartService) calculateSummary(items []*model.CartItem) *cartv1.CartSummary {
	var totalItems int32
	var totalQuantity int32
	var totalPrice float64
	hasInvalidItems := false

	for _, item := range items {
		if item.IsValid {
			totalItems++
			totalQuantity += item.Quantity
			// 使用当前价格计算（已经是 float64）
			totalPrice += item.CurrentPrice * float64(item.Quantity)
		} else {
			hasInvalidItems = true
		}
	}

	return &cartv1.CartSummary{
		TotalItems:      totalItems,
		TotalQuantity:   totalQuantity,
		TotalPrice:      fmt.Sprintf("%.2f", totalPrice),
		HasInvalidItems: hasInvalidItems,
	}
}

// calculateProductTotal 计算商品总金额
func (s *CartService) calculateProductTotal(items []*model.CartItem) string {
	var total float64
	for _, item := range items {
		// 使用当前价格计算（已经是 float64）
		total += item.CurrentPrice * float64(item.Quantity)
	}
	return fmt.Sprintf("%.2f", total)
}

// calculateFinalAmount 计算最终实付金额
// 公式：最终金额 = 商品总价 - 促销优惠 - 优惠券优惠 + 运费
func (s *CartService) calculateFinalAmount(productTotal, promotionDiscount, couponDiscount, shippingFee string) string {
	product, _ := strconv.ParseFloat(productTotal, 64)
	promotion, _ := strconv.ParseFloat(promotionDiscount, 64)
	coupon, _ := strconv.ParseFloat(couponDiscount, 64)
	shipping, _ := strconv.ParseFloat(shippingFee, 64)

	final := product - promotion - coupon + shipping
	if final < 0 {
		final = 0 // 防止负数
	}
	return fmt.Sprintf("%.2f", final)
}

// updateProductInfoForCheckout 结算预览时实时更新商品信息（价格、库存、状态）
// 使用并发调用优化性能，避免串行等待
func (s *CartService) updateProductInfoForCheckout(ctx context.Context, items []*model.CartItem) {
	if len(items) == 0 {
		return
	}

	var wg sync.WaitGroup
	mu := sync.Mutex{}

	// 1. 并发获取商品信息（价格、状态）
	if s.productClient != nil {
		wg.Add(len(items))
		for _, item := range items {
			go func(item *model.CartItem) {
				defer wg.Done()
				product, skus, err := s.productClient.GetProduct(ctx, item.ProductID)
				if err != nil {
					log.Printf("⚠️ [Service] CheckoutPreview: 获取商品信息失败 - product_id=%s, error=%v", item.ProductID, err)
					// 降级处理：使用缓存的价格，标记为可能失效
					return
				}

				if product == nil || len(skus) == 0 {
					log.Printf("⚠️ [Service] CheckoutPreview: 商品不存在或没有SKU - product_id=%s", item.ProductID)
					mu.Lock()
					item.IsValid = false
					item.InvalidReason = "商品不存在或已下架"
					mu.Unlock()
					return
				}

				// 查找对应的 SKU
				var targetSKU *productv1.SkuInfo
				for _, sku := range skus {
					if sku.Id == item.SKUID {
						targetSKU = sku
						break
					}
				}

				if targetSKU == nil {
					log.Printf("⚠️ [Service] CheckoutPreview: SKU不存在 - sku_id=%s", item.SKUID)
					mu.Lock()
					item.IsValid = false
					item.InvalidReason = "SKU不存在或已下架"
					mu.Unlock()
					return
				}

				// 更新商品信息
				mu.Lock()
				// 更新当前价格（实时价格）
				if targetSKU.Price > 0 {
					item.CurrentPrice = targetSKU.Price
				}

				// 校验商品状态（商品状态：1-待审核，2-审核失败，3-待上架，4-已上架，5-已下架）
				if product.Status != productRepository.ProductStatusOnShelf {
					log.Printf("商品状态异常: %d", product.Status)
					item.IsValid = false
					item.InvalidReason = "商品状态异常"

				}
				mu.Unlock()
			}(item)
		}
	}

	// 2. 并发获取库存信息
	if s.inventoryClient != nil {
		// 收集所有 skuID（去重）
		skuIDSet := make(map[string]bool)
		skuIDToItems := make(map[string][]*model.CartItem)
		for _, item := range items {
			if !skuIDSet[item.SKUID] {
				skuIDSet[item.SKUID] = true
				skuIDToItems[item.SKUID] = []*model.CartItem{item}
			} else {
				skuIDToItems[item.SKUID] = append(skuIDToItems[item.SKUID], item)
			}
		}

		// 转换为 skuID 列表
		skuIDs := make([]string, 0, len(skuIDSet))
		for skuID := range skuIDSet {
			skuIDs = append(skuIDs, skuID)
		}

		// 批量获取库存
		stocksMap, err := s.inventoryClient.BatchGetStock(ctx, skuIDs)
		if err != nil {
			log.Printf("⚠️ [Service] CheckoutPreview: 批量获取库存失败 - error=%v", err)
			// 降级处理：库存服务调用失败时，标记所有商品为需要重新校验
			// 注意：不修改 item.Stock，因为旧值可能不准确
			mu.Lock()
			for _, item := range items {
				// 如果商品状态正常但库存信息获取失败，标记为无效
				if item.IsValid {
					item.IsValid = false
					item.InvalidReason = "库存信息获取失败，请稍后重试"
				}
			}
			mu.Unlock()
		} else {
			// 更新库存信息
			mu.Lock()
			for skuID, stockInfo := range stocksMap {
				if items, ok := skuIDToItems[skuID]; ok {
					for _, item := range items {
						item.Stock = int32(stockInfo.AvailableStock)
						// 如果库存不足，标记为无效
						if item.Stock <= 0 {
							item.Stock = 0
							item.IsValid = false
							item.InvalidReason = "库存不足"
						} else if item.Quantity > item.Stock {
							// 如果购买数量超过库存，标记为无效
							item.IsValid = false
							item.InvalidReason = fmt.Sprintf("库存不足，当前可用库存：%d", item.Stock)
						}
					}
				}
			}
			mu.Unlock()
		}
	}

	// 等待所有商品信息查询完成
	wg.Wait()
}
