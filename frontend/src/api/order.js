import request from './request'

// 生成订单幂等性Token
export function generateOrderToken() {
  return request.get('/orders/token')
}

// 创建订单
export function createOrder(data) {
  return request.post('/orders', {
    items: data.items, // [{ product_id, sku_id, quantity }]
    address_id: data.addressId,
    coupon_id: data.couponId || '',
    buyer_remark: data.buyerRemark || '',
    token: data.token // 幂等性Token
  })
}

// 获取订单详情
export function getOrderDetail(orderNo) {
  return request.get(`/orders/${orderNo}`)
}

// 获取订单列表
export function getOrderList(params = {}) {
  return request.get('/orders', {
    params: {
      status: params.status || 0, // 0 表示全部
      page: params.page || 1,
      page_size: params.pageSize || 20
    }
  })
}

// 取消订单
export function cancelOrder(orderNo, reason = '') {
  return request.post(`/orders/${orderNo}/cancel`, {
    reason
  })
}

