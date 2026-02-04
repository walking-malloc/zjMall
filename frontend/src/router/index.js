import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/',
    component: () => import('@/components/Layout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: 'product/products',
        name: 'Products',
        component: () => import('@/views/Products.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: 'product/products/:id',
        name: 'ProductDetail',
        component: () => import('@/views/ProductDetail.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: 'cart',
        name: 'Cart',
        component: () => import('@/views/Cart.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/Orders.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'pay',
        name: 'Pay',
        component: () => import('@/views/Pay.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'addresses',
        name: 'Addresses',
        component: () => import('@/views/Addresses.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'create-order-test',
        name: 'CreateOrderTest',
        component: () => import('@/views/CreateOrderTest.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'checkout',
        name: 'Checkout',
        component: () => import('@/views/Checkout.vue'),
        meta: { requiresAuth: true }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { requiresAuth: false }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    // 同时检查 store 和 localStorage 中的 token（防止 store 未初始化）
    const tokenInStore = userStore.isLoggedIn
    const tokenInStorage = localStorage.getItem('token')
    
    console.log('路由守卫检查:', {
      path: to.path,
      tokenInStore,
      tokenInStorage: !!tokenInStorage,
      tokenValue: tokenInStorage ? tokenInStorage.substring(0, 20) + '...' : 'null'
    })
    
    if (!tokenInStore && !tokenInStorage) {
      // 如果都没有 token，跳转到登录页
      console.log('路由守卫: 未找到 token，跳转到登录页')
      next({ name: 'Login', query: { redirect: to.fullPath } })
    } else {
      // 如果 localStorage 有但 store 没有，同步一下
      if (tokenInStorage && !tokenInStore) {
        console.log('路由守卫: 同步 token 到 store')
        userStore.setToken(tokenInStorage)
      }
      console.log('路由守卫: 允许访问')
      next()
    }
  } else {
    next()
  }
})

export default router

