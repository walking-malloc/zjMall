import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, register, getUserInfo, loginBySMS as loginBySMSApi } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  // 初始化时尝试从 localStorage 恢复用户信息
  let initialUserInfo = null
  try {
    const savedUserInfo = localStorage.getItem('userInfo')
    if (savedUserInfo) {
      initialUserInfo = JSON.parse(savedUserInfo)
    }
  } catch (error) {
    console.error('恢复用户信息失败:', error)
  }
  const userInfo = ref(initialUserInfo)

  const isLoggedIn = computed(() => !!token.value)

  function setToken(newToken) {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  }

  function setUserInfo(info) {
    userInfo.value = info
    // 同时保存到 localStorage，方便页面刷新后恢复
    if (info) {
      localStorage.setItem('userInfo', JSON.stringify(info))
    } else {
      localStorage.removeItem('userInfo')
    }
  }

  async function loginUser(phone, password) {
    try {
      const res = await login(phone, password)
      if (res.data.code === 0) {
        setToken(res.data.data.token)
        // 登录响应中包含用户信息，直接设置
        if (res.data.data.user) {
          setUserInfo(res.data.data.user)
        }
        return { success: true }
      } else {
        return { success: false, message: res.data.message }
      }
    } catch (error) {
      return { success: false, message: error.message || '登录失败' }
    }
  }

  async function registerUser(phone, password, smsCode) {
    try {
      const res = await register(phone, password, smsCode)
      if (res.data.code === 0) {
        setToken(res.data.data.token)
        // 注册响应中包含用户信息，直接设置
        if (res.data.data.user) {
          setUserInfo(res.data.data.user)
        }
        return { success: true }
      } else {
        return { success: false, message: res.data.message }
      }
    } catch (error) {
      return { success: false, message: error.message || '注册失败' }
    }
  }

  async function fetchUserInfo() {
    if (!token.value) return
    // 从token中解析用户ID（简化处理，实际应该从token payload中解析）
    // 这里暂时从userInfo中获取，如果没有则跳过
    if (!userInfo.value || !userInfo.value.id) return
    try {
      const res = await getUserInfo(userInfo.value.id)
      if (res.data.code === 0) {
        setUserInfo(res.data.data)
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
    }
  }

  async function loginBySMS(phone, smsCode) {
    try {
      const res = await loginBySMSApi(phone, smsCode)
      if (res.data.code === 0) {
        setToken(res.data.data.token)
        if (res.data.data.user) {
          setUserInfo(res.data.data.user)
        }
        return { success: true }
      } else {
        return { success: false, message: res.data.message }
      }
    } catch (error) {
      return { success: false, message: error.message || '登录失败' }
    }
  }

  function logout() {
    setToken('')
    setUserInfo(null)
    // 清除 localStorage 中的用户信息
    localStorage.removeItem('userInfo')
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    setToken,
    setUserInfo,
    loginUser,
    loginBySMS,
    registerUser,
    fetchUserInfo,
    logout
  }
})

