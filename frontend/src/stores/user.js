import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, register, getUserInfo } from '@/api/user'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(null)

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

  function logout() {
    setToken('')
    setUserInfo(null)
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    loginUser,
    registerUser,
    fetchUserInfo,
    logout
  }
})

