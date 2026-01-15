<template>
  <div class="home">
    <el-container>
      <!-- 头部导航 -->
      <el-header>
        <div class="header-content">
          <div class="logo" @click="$router.push('/')">
            <h1>zjMall</h1>
          </div>
          <div class="nav-menu">
            <el-menu
              mode="horizontal"
              :default-active="activeMenu"
              @select="handleMenuSelect"
            >
              <el-menu-item index="home">首页</el-menu-item>
              <el-menu-item index="products">商品</el-menu-item>
            </el-menu>
            <div class="user-actions">
              <template v-if="userStore.isLoggedIn">
                <el-dropdown @command="handleUserCommand">
                  <span class="user-name">
                    {{ userStore.userInfo?.nickname || userStore.userInfo?.phone }}
                    <el-icon><arrow-down /></el-icon>
                  </span>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                      <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
              <template v-else>
                <el-button type="text" @click="$router.push('/login')">登录</el-button>
                <el-button type="primary" @click="$router.push('/register')">注册</el-button>
              </template>
            </div>
          </div>
        </div>
      </el-header>

      <!-- 主要内容 -->
      <el-main>
        <div class="banner">
          <h2>欢迎来到 zjMall</h2>
          <p>发现更多优质商品</p>
          <el-button type="primary" size="large" @click="$router.push('/product/products')">
            开始购物
          </el-button>
        </div>

        <!-- 热门商品 -->
        <div class="hot-products">
          <h3>热门商品</h3>
          <el-row :gutter="20" v-loading="loading">
            <el-col :span="6" v-for="product in products" :key="product.id">
              <el-card class="product-card" @click="goToDetail(product.id)">
                <img :src="product.main_image || '/placeholder.png'" class="product-image" />
                <div class="product-info">
                  <h4>{{ product.title }}</h4>
                  <p class="price">¥{{ product.price || '0.00' }}</p>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getProductList } from '@/api/product'
import { ArrowDown } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

const activeMenu = ref('home')
const products = ref([])
const loading = ref(false)

const handleMenuSelect = (index) => {
  if (index === 'products') {
    router.push('/product/products')
  }
}

const handleUserCommand = (command) => {
  if (command === 'profile') {
    router.push('/profile')
  } else if (command === 'logout') {
    userStore.logout()
    router.push('/')
  }
}

const goToDetail = (id) => {
  router.push(`/product/${id}`)
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await getProductList({ page: 1, page_size: 8 })
    if (res.data.code === 0) {
      products.value = res.data.data || []
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.home {
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

.nav-menu {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-name {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.banner {
  text-align: center;
  padding: 60px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  margin-bottom: 40px;
}

.banner h2 {
  font-size: 48px;
  margin-bottom: 20px;
}

.banner p {
  font-size: 20px;
  margin-bottom: 30px;
}

.hot-products {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.hot-products h3 {
  font-size: 24px;
  margin-bottom: 20px;
}

.product-card {
  cursor: pointer;
  transition: transform 0.3s;
}

.product-card:hover {
  transform: translateY(-5px);
}

.product-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
}

.product-info {
  padding: 10px 0;
}

.product-info h4 {
  font-size: 16px;
  margin-bottom: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price {
  color: #f56c6c;
  font-size: 20px;
  font-weight: bold;
}
</style>

