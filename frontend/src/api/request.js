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
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // 添加调试日志
    console.log('API 请求:', {
      url: config.url,
      method: config.method,
      baseURL: config.baseURL,
      fullURL: config.baseURL + config.url,
      params: config.params,
      data: config.data
    })
    return config
  },
  error => {
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

