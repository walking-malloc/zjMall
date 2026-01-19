import axios from 'axios'
import { API_BASE_URL, API_TIMEOUT } from './config'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器，每次发送请求时直接执行此函数
request.interceptors.request.use(
  config => {
    // 强制从 localStorage 读取 token（确保获取最新值）
    const token = localStorage.getItem('token')
    
    // 详细日志
    console.log('=== 请求拦截器执行 ===', {
      url: config.url,
      method: config.method,
      tokenExists: !!token,
      tokenValue: token ? token.substring(0, 50) + '...' : 'null',
      tokenLength: token ? token.length : 0,
      allLocalStorageKeys: Object.keys(localStorage)
    })
    
    if (token && token.trim() !== '') {
      // 确保设置 Authorization header
      config.headers.Authorization = `Bearer ${token.trim()}`
      console.log('✅ Token 已设置到请求头:', {
        url: config.url,
        headerValue: config.headers.Authorization.substring(0, 50) + '...'
      })
    } else {
      console.error('❌ Token 不存在或为空:', {
        url: config.url,
        token: token,
        localStorageKeys: Object.keys(localStorage),
        localStorageToken: localStorage.getItem('token')
      })
    }
    
    // 最终确认
    console.log('请求配置:', {
      url: config.url,
      method: config.method,
      baseURL: config.baseURL,
      fullURL: config.baseURL + config.url,
      hasAuthorization: !!config.headers.Authorization,
      authorizationHeader: config.headers.Authorization ? config.headers.Authorization.substring(0, 50) + '...' : '未设置'
    })
    
    return config
  },
  error => {
    console.error('请求拦截器错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    // 添加调试日志
    console.log('API 响应:', {
      url: response.config.url,
      method: response.config.method,
      status: response.status,
      data: response.data
    })
    
    // 检查业务错误码（后端返回 code !== 0 表示业务失败）
    if (response.data && response.data.code !== undefined && response.data.code !== 0) {
      const errorMessage = response.data.message || '操作失败'
      ElMessage.error(errorMessage)
      // 如果是未登录错误，跳转到登录页
      if (response.data.code === 401 || errorMessage.includes('未登录') || errorMessage.includes('登录')) {
        localStorage.removeItem('token')
        router.push('/login')
      }
      // 返回一个 rejected promise，让调用方知道失败了
      return Promise.reject(new Error(errorMessage))
    }
    
    return response
  },
  error => {
    if (error.response) {
      const { status, data } = error.response
      
      if (status === 401) {
        localStorage.removeItem('token')
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
      } else if (status === 403) {
        ElMessage.error('没有权限访问')
      } else if (status >= 500) {
        ElMessage.error('服务器错误，请稍后重试')
      } else {
        ElMessage.error(data.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default request

