import request from './request'

// 创建支付单
export function createPayment(data) {
  return request.post('/payments', {
    order_no: data.orderNo,
    pay_channel: data.payChannel,
    return_url: data.returnUrl || ''
  })
}

// 查询支付状态
export function queryPaymentStatus(paymentNo) {
  return request.get(`/payments/${paymentNo}/status`)
}

