import request from './request'

// 用户注册
export function register(phone, password, smsCode) {
  return request.post('/users/register', {
    phone,
    password,
    confirm_password: password,
    sms_code: smsCode
  })
}

// 用户登录
export function login(phone, password) {
  return request.post('/users/login', {
    phone,
    password
  })
}

// 验证码登录
export function loginBySMS(phone, smsCode) {
  return request.post('/users/login-by-sms', {
    phone,
    sms_code: smsCode
  })
}

// 获取短信验证码
export function getSMSCode(phone) {
  return request.post('/users/sms-code', {
    phone
  })
}

// 获取用户信息（需要从token中获取user_id，这里先简化处理）
export function getUserInfo(userId) {
  return request.get(`/users/${userId}`)
}

// 更新用户信息
// 注意：需要从 token 或 userStore 中获取 user_id
export function updateUserInfo(userId, data) {
  return request.put(`/users/${userId}`, data)
}

// 登出
export function logout() {
  return request.post('/users/logout')
}

