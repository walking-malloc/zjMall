import request from './request'

// 创建支付单
export function createPayment(data) {
  return request.post('/payments', {
    order_no: data.orderNo,
    pay_channel: data.payChannel,
    return_url: data.returnUrl || '',
    token: data.token || ''
  })
}

// 查询支付状态
export function queryPaymentStatus(paymentNo) {
  return request.get(`/payments/${paymentNo}/status`)
}

// 生成支付幂等性 Token（需要携带订单号，作为 order_no 查询参数）
export function generatePaymentToken(orderNo) {
  return request.get('/payments/token', {
    params: {
      order_no: orderNo
    }
  })
}
