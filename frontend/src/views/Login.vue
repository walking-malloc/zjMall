<template>
  <div class="login-page">
    <!-- 头部 Logo -->
    <div class="login-header">
      <div class="logo" @click="$router.push('/')">
        <h1>zjMall</h1>
      </div>
    </div>

    <!-- 登录主体 -->
    <div class="login-main">
      <div class="login-box">
        <!-- 登录方式标签页 -->
        <div class="login-tabs">
          <div class="tab-item" :class="{ active: activeTab === 'password' }" @click="activeTab = 'password'">
            密码登录
          </div>
          <div class="tab-item" :class="{ active: activeTab === 'sms' }" @click="activeTab = 'sms'">
            短信登录
          </div>
        </div>

        <!-- 密码登录表单 -->
        <div v-show="activeTab === 'password'" class="login-form">
          <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="0">
            <el-form-item prop="phone">
              <el-input v-model="passwordForm.phone" placeholder="账号名/手机号/邮箱" size="large" clearable />
            </el-form-item>

            <el-form-item prop="password">
              <el-input v-model="passwordForm.password" type="password" placeholder="密码" size="large" show-password
                @keyup.enter="handlePasswordLogin" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handlePasswordLogin" size="large" class="login-btn">
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <!-- 短信登录表单 -->
        <div v-show="activeTab === 'sms'" class="login-form">
          <el-form ref="smsFormRef" :model="smsForm" :rules="smsRules" label-width="0">
            <el-form-item prop="phone">
              <el-input v-model="smsForm.phone" placeholder="手机号" size="large" clearable maxlength="11" />
            </el-form-item>

            <el-form-item prop="smsCode">
              <div class="sms-input-group">
                <el-input v-model="smsForm.smsCode" placeholder="验证码" size="large" maxlength="6"
                  @keyup.enter="handleSMSLogin" />
                <el-button :disabled="codeCountdown > 0" @click="handleGetCode" class="code-btn">
                  {{ codeCountdown > 0 ? `${codeCountdown}秒` : '获取验证码' }}
                </el-button>
              </div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSMSLogin" size="large" class="login-btn">
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </div>

        <!-- 底部链接 -->
        <div class="login-footer">
          <div class="footer-links">
            <span class="link-item" @click="$router.push('/register')">立即注册</span>
            <span class="divider">|</span>
            <span class="link-item" @click="handleForgotPassword">忘记密码</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getSMSCode } from '@/api/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('password')
const loading = ref(false)
const codeCountdown = ref(0)

const passwordFormRef = ref(null)
const smsFormRef = ref(null)

const passwordForm = reactive({
  phone: '',
  password: ''
})

const smsForm = reactive({
  phone: '',
  smsCode: ''
})

const passwordRules = {
  phone: [
    { required: true, message: '请输入账号名/手机号/邮箱', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ]
}

const smsRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  smsCode: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码为6位数字', trigger: 'blur' }
  ]
}

// 密码登录
const handlePasswordLogin = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const result = await userStore.loginUser(passwordForm.phone, passwordForm.password)
        if (result.success) {
          ElMessage.success('登录成功')
          const redirect = route.query.redirect || '/'
          router.push(redirect)
        } else {
          ElMessage.error(result.message || '登录失败')
        }
      } catch (error) {
        ElMessage.error('登录失败，请稍后重试')
      } finally {
        loading.value = false
      }
    }
  })
}

// 短信登录
const handleSMSLogin = async () => {
  if (!smsFormRef.value) return

  await smsFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const result = await userStore.loginBySMS(smsForm.phone, smsForm.smsCode)
        if (result.success) {
          ElMessage.success('登录成功')
          const redirect = route.query.redirect || '/'
          router.push(redirect)
        } else {
          ElMessage.error(result.message || '登录失败')
        }
      } catch (error) {
        ElMessage.error('登录失败，请稍后重试')
      } finally {
        loading.value = false
      }
    }
  })
}

// 获取验证码
const handleGetCode = async () => {
  if (!smsForm.phone) {
    ElMessage.warning('请先输入手机号')
    return
  }

  if (!/^1[3-9]\d{9}$/.test(smsForm.phone)) {
    ElMessage.warning('请输入正确的手机号')
    return
  }

  try {
    await getSMSCode(smsForm.phone)
    ElMessage.success('验证码已发送')
    codeCountdown.value = 60
    const timer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (error) {
    ElMessage.error('获取验证码失败，请稍后重试')
  }
}

// 忘记密码
const handleForgotPassword = () => {
  ElMessage.info('忘记密码功能开发中...')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

/* 头部 */
.login-header {
  background: #fff;
  padding: 20px 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.login-header .logo {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  cursor: pointer;
}

.login-header .logo h1 {
  margin: 0;
  color: #3978e4;
  font-size: 32px;
  font-weight: bold;
}

/* 主体 */
.login-main {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
}

.login-box {
  width: 400px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 40px;
}

/* 标签页 */
.login-tabs {
  display: flex;
  border-bottom: 2px solid #e5e5e5;
  margin-bottom: 30px;
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 15px 0;
  font-size: 18px;
  color: #666;
  cursor: pointer;
  transition: all 0.3s;
  position: relative;
}

.tab-item:hover {
  color: #3978e4;
}

.tab-item.active {
  color: #3978e4;
  font-weight: bold;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: #3978e4;
}

/* 表单 */
.login-form {
  margin-bottom: 20px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 4px;
  box-shadow: 0 0 0 1px #dcdfe6 inset;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #c0c4cc inset;
}

.login-form :deep(.el-input.is-focus .el-input__wrapper) {
  box-shadow: 0 0 0 1px #3978e4 inset;
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  background: #3978e4;
  border: none;
  border-radius: 4px;
  transition: all 0.3s;
}

.login-btn:hover {
  background: #3978e4;
}

.login-btn:active {
  background: #3978e4;
}

/* 短信验证码输入组 */
.sms-input-group {
  display: flex;
  gap: 10px;
}

.sms-input-group .el-input {
  flex: 1;
}

.code-btn {
  white-space: nowrap;
  min-width: 100px;
}

/* 底部链接 */
.login-footer {
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e5e5e5;
}

.footer-links {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.link-item {
  color: #999;
  cursor: pointer;
  transition: color 0.3s;
}

.link-item:hover {
  color: #e4393c;
}

.divider {
  color: #e5e5e5;
}

/* 响应式 */
@media (max-width: 768px) {
  .login-box {
    width: 100%;
    max-width: 400px;
    padding: 30px 20px;
  }
}
</style>
