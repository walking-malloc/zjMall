import request from './request'

// 添加商品到购物车
export function addCartItem(data) {
  // data: { product_id, sku_id, quantity }
  return request.post('/cart/items', data)
}

// 更新购物车商品数量
export function updateCartItemQuantity(itemId, quantity) {
  return request.put(`/cart/items/${itemId}/quantity`, {
    item_id: itemId,
    quantity
  })
}

// 删除单个购物车商品
export function removeCartItem(itemId) {
  return request.delete(`/cart/items/${itemId}`)
}

// 批量删除购物车商品
export function removeCartItems(itemIds) {
  return request.post('/cart/items/batch-delete', {
    item_ids: itemIds
  })
}

// 清空购物车
export function clearCart() {
  return request.delete('/cart')
}

// 获取购物车列表
export function getCart() {
  return request.get('/cart')
}

// 获取购物车统计信息
export function getCartSummary() {
  return request.get('/cart/summary')
}

// 结算预览
export function checkoutPreview(payload) {
  // payload: { item_ids?: string[], address_id?: string, coupon_id?: string }
  return request.post('/cart/checkout-preview', payload)
}


