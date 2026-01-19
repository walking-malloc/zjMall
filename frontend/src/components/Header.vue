<template>
  <el-header class="app-header">
    <div class="header-content">
      <div class="logo" @click="$router.push('/')">
        <h1>zjMall</h1>
      </div>

      <div class="search-bar">
        <el-input v-model="searchKeyword" placeholder="搜索商品" @keyup.enter="handleSearch" clearable>
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
          <template #append>
            <el-button @click="handleSearch">搜索</el-button>
          </template>
        </el-input>
      </div>

      <div class="nav-menu">
        <el-button type="text" @click="$router.push('/')">首页</el-button>
        <el-button type="text" @click="$router.push('/product/products')">商品</el-button>

        <el-badge :value="cartCount" :hidden="cartCount === 0" class="cart-badge">
          <el-button type="text" @click="$router.push('/cart')">
            <el-icon>
              <ShoppingCart />
            </el-icon>
            购物车
          </el-button>
        </el-badge>

        <div class="user-actions">
          <template v-if="userStore.isLoggedIn">
            <el-dropdown @command="handleUserCommand">
              <span class="user-name">
                <el-avatar :size="30" :src="userStore.userInfo?.avatar">
                  {{ userStore.userInfo?.nickname?.[0] || userStore.userInfo?.phone?.[0] }}
                </el-avatar>
                <span style="margin-left: 8px;">
                  {{ userStore.userInfo?.nickname || userStore.userInfo?.phone }}
                </span>
                <el-icon>
                  <ArrowDown />
                </el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">
                    <el-icon>
                      <User />
                    </el-icon>
                    个人中心
                  </el-dropdown-item>
                  <el-dropdown-item command="orders">
                    <el-icon>
                      <List />
                    </el-icon>
                    我的订单
                  </el-dropdown-item>
                  <el-dropdown-item command="addresses">
                    <el-icon>
                      <Location />
                    </el-icon>
                    收货地址
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon>
                      <SwitchButton />
                    </el-icon>
                    退出登录
                  </el-dropdown-item>
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
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import {
  Search,
  ShoppingCart,
  ArrowDown,
  User,
  List,
  Location,
  SwitchButton
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const searchKeyword = ref('')

const cartCount = computed(() => cartStore.totalCount)

const handleSearch = () => {
  if (searchKeyword.value.trim()) {
    router.push({
      path: '/product/products',
      query: { keyword: searchKeyword.value }
    })
  }
}

const handleUserCommand = (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'orders':
      router.push('/orders')
      break
    case 'addresses':
      router.push('/addresses')
      break
    case 'logout':
      userStore.logout()
      router.push('/')
      break
  }
}
</script>

<style scoped>
.app-header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 0 20px;
  height: 70px !important;
  line-height: 70px;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1400px;
  margin: 0 auto;
  height: 100%;
}

.logo {
  cursor: pointer;
  margin-right: 40px;
}

.logo h1 {
  margin: 0;
  color: #409eff;
  font-size: 24px;
  font-weight: bold;
}

.search-bar {
  flex: 1;
  max-width: 500px;
  margin: 0 40px;
}

.nav-menu {
  display: flex;
  align-items: center;
  gap: 20px;
}

.cart-badge {
  margin-right: 10px;
}

.user-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-name {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-name:hover {
  background-color: #f5f5f5;
}
</style>
