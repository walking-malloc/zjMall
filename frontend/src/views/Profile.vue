<template>
  <div class="profile-page">
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="logo" @click="$router.push('/')">
            <h1>zjMall</h1>
          </div>
          <div class="nav-menu">
            <el-button type="text" @click="$router.push('/')">首页</el-button>
            <el-button type="text" @click="$router.push('product/products')">商品列表</el-button>
          </div>
        </div>
      </el-header>

      <el-main>
        <div class="profile-container">
          <el-card>
            <template #header>
              <h2>个人中心</h2>
            </template>
            
            <div v-if="userStore.userInfo" class="user-info">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="用户ID">
                  {{ userStore.userInfo.id }}
                </el-descriptions-item>
                <el-descriptions-item label="手机号">
                  {{ userStore.userInfo.phone }}
                </el-descriptions-item>
                <el-descriptions-item label="昵称">
                  {{ userStore.userInfo.nickname || '未设置' }}
                </el-descriptions-item>
                <el-descriptions-item label="邮箱">
                  {{ userStore.userInfo.email || '未设置' }}
                </el-descriptions-item>
              </el-descriptions>
              
              <div class="actions" style="margin-top: 20px;">
                <el-button @click="handleLogout">退出登录</el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(async () => {
  // 检查是否已登录（检查 localStorage 中的 token）
  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  // 如果用户信息不存在，尝试从 localStorage 恢复
  if (!userStore.userInfo) {
    const savedUserInfo = localStorage.getItem('userInfo')
    if (savedUserInfo) {
      try {
        const userInfo = JSON.parse(savedUserInfo)
        userStore.setUserInfo(userInfo)
      } catch (error) {
        console.error('解析用户信息失败:', error)
        ElMessage.warning('用户信息已过期，请重新登录')
      }
    } else {
      // 如果没有保存的用户信息，提示用户重新登录
      ElMessage.warning('用户信息已过期，请重新登录')
    }
  }
})
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.logo h1 {
  margin: 0;
  color: #409eff;
  cursor: pointer;
}

.profile-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.user-info {
  padding: 20px 0;
}
</style>

