# 前端请求流程详解

本文档详细说明一个请求从前端到后端的完整流程，以及每个文件的作用。

---

## 📋 完整流程图

```
用户操作（点击/输入）
    ↓
页面组件（views/Products.vue）
    ↓
调用 API 函数（api/product.js）
    ↓
HTTP 请求封装（api/request.js）
    ↓
请求拦截器（自动添加 token）
    ↓
Vite 代理（vite.config.js）
    ↓
后端服务（localhost:8082）
    ↓
后端响应
    ↓
响应拦截器（统一错误处理）
    ↓
返回数据到组件
    ↓
更新响应式数据
    ↓
Vue 自动更新视图
```

---

## 🔍 详细步骤解析

### 示例：用户访问商品列表页并加载商品

**场景**：用户在浏览器输入 `http://localhost:3000/product/products`

---

## 第一步：应用启动（main.js）

**文件**：`frontend/src/main.js`

```javascript
import { createApp } from 'vue'
import router from './router'      // 导入路由
import App from './App.vue'        // 导入根组件

const app = createApp(App)
app.use(router)                    // 注册路由
app.mount('#app')                  // 挂载到 #app
```

**作用**：
- ✅ 创建 Vue 应用实例
- ✅ 注册路由系统
- ✅ 挂载应用到 DOM

**匹配时机**：应用启动时执行一次

---

## 第二步：根组件（App.vue）

**文件**：`frontend/src/App.vue`

```vue
<template>
  <router-view />  <!-- 路由出口：显示匹配的页面组件 -->
</template>
```

**作用**：
- ✅ 提供路由出口（`<router-view />`）
- ✅ 根据当前路由显示对应的页面组件

**匹配时机**：应用启动后，根据 URL 路径匹配

---

## 第三步：路由匹配（router/index.js）

**文件**：`frontend/src/router/index.js`

**URL 路径**：`/product/products`

**匹配过程**：

```javascript
// 1. 路由表匹配
const routes = [
  {
    path: '/',
    component: Layout,              // 匹配到 Layout 组件
    children: [
      {
        path: 'product/products',   // ✅ 匹配成功
        name: 'Products',
        component: Products         // 加载 Products.vue 组件
      }
    ]
  }
]

// 2. 路由守卫检查
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next({ name: 'Login' })  // 需要登录但未登录，跳转登录页
  } else {
    next()  // ✅ 允许访问，继续
  }
})
```

**匹配结果**：
- ✅ 匹配到 `Layout` 组件（父组件）
- ✅ 匹配到 `Products.vue` 组件（子组件）
- ✅ 路由守卫通过（`requiresAuth: false`）

**匹配时机**：URL 变化时执行

---

## 第四步：布局组件（components/Layout.vue）

**文件**：`frontend/src/components/Layout.vue`

```vue
<template>
  <el-container>
    <Header />                    <!-- 头部导航 -->
    <el-main>
      <router-view />            <!-- 显示 Products.vue -->
    </el-main>
    <Footer />                   <!-- 页脚 -->
  </el-container>
</template>
```

**作用**：
- ✅ 提供统一的页面布局（Header + 内容 + Footer）
- ✅ 通过 `<router-view />` 显示匹配的页面组件

**匹配时机**：路由匹配到 Layout 时渲染

---

## 第五步：页面组件（views/Products.vue）

**文件**：`frontend/src/views/Products.vue`

**组件挂载过程**：

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { getProductList } from '@/api/product'  // 导入 API 函数

const products = ref([])  // 响应式数据

// 组件挂载后执行
onMounted(async () => {
  // ✅ 调用 API 函数加载商品
  const res = await getProductList({ page: 1, page_size: 12 })
  if (res.data.code === 0) {
    products.value = res.data.data  // 更新响应式数据
  }
})
</script>
```

**执行顺序**：
1. ✅ 组件创建（`<script setup>` 执行）
2. ✅ 定义响应式数据（`products = ref([])`）
3. ✅ 组件挂载到 DOM（`onMounted` 执行）
4. ✅ 调用 API 函数（`getProductList()`）

**匹配时机**：路由匹配到 Products 组件时执行

---

## 第六步：API 函数（api/product.js）

**文件**：`frontend/src/api/product.js`

```javascript
import request from './request'  // 导入 HTTP 请求封装

// API 函数定义
export function getProductList(params) {
  // ✅ 调用 request.get，路径是相对路径
  return request.get('/product/products', { params })
}
```

**作用**：
- ✅ 封装具体的 API 调用
- ✅ 定义请求路径和参数
- ✅ 返回 Promise 对象

**匹配时机**：组件调用 `getProductList()` 时执行

---

## 第七步：HTTP 请求封装（api/request.js）

**文件**：`frontend/src/api/request.js`

**执行流程**：

### 7.1 创建请求配置

```javascript
import { API_BASE_URL } from './config'  // '/api/v1'

const request = axios.create({
  baseURL: API_BASE_URL,  // '/api/v1'
  timeout: 10000
})

// 调用 request.get('/product/products', { params })
// 实际 URL = baseURL + url = '/api/v1' + '/product/products'
// = '/api/v1/product/products'
```

### 7.2 请求拦截器（发送请求前）

```javascript
request.interceptors.request.use(config => {
  // ✅ 1. 从 localStorage 获取 token
  const token = localStorage.getItem('token')
  
  // ✅ 2. 如果有 token，添加到请求头
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  
  // ✅ 3. 打印请求日志（调试用）
  console.log('API 请求:', {
    url: config.url,              // '/product/products'
    baseURL: config.baseURL,      // '/api/v1'
    fullURL: config.baseURL + config.url  // '/api/v1/product/products'
  })
  
  // ✅ 4. 返回配置，继续发送请求
  return config
})
```

**最终请求配置**：
```javascript
{
  method: 'GET',
  url: '/product/products',
  baseURL: '/api/v1',
  fullURL: '/api/v1/product/products',
  params: { page: 1, page_size: 12 },
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer xxx'  // 如果有 token
  }
}
```

**匹配时机**：每次发送 HTTP 请求前执行

---

## 第八步：Vite 代理（vite.config.js）

**文件**：`frontend/vite.config.js`

**代理配置**：

```javascript
server: {
  proxy: {
    '/api/v1/product': {              // ✅ 匹配到商品服务
      target: 'http://localhost:8082', // 转发到后端服务
      changeOrigin: true
    }
  }
}
```

**代理过程**：

```
前端请求: GET http://localhost:3000/api/v1/product/products?page=1&page_size=12
    ↓
Vite 开发服务器接收
    ↓
匹配代理规则: /api/v1/product
    ↓
转发请求: GET http://localhost:8082/api/v1/product/products?page=1&page_size=12
    ↓
后端服务处理
```

**匹配时机**：开发环境下，所有 `/api/v1/product` 开头的请求都会被代理

---

## 第九步：后端服务处理

**后端服务**：`http://localhost:8082`

**处理流程**：
1. 接收 HTTP 请求
2. 解析请求参数（`page=1`, `page_size=12`）
3. 验证 token（从 `Authorization` header）
4. 调用业务逻辑
5. 返回 JSON 响应

**响应示例**：
```json
{
  "code": 0,
  "message": "查询成功",
  "total": 100,
  "data": [
    {
      "id": "123",
      "title": "商品名称",
      "price": "99.00"
    }
  ]
}
```

---

## 第十步：响应拦截器（收到响应后）

**文件**：`frontend/src/api/request.js`

**响应拦截器**：

```javascript
request.interceptors.response.use(
  // ✅ 成功响应
  response => {
    console.log('API 响应:', {
      url: response.config.url,
      status: response.status,
      data: response.data
    })
    return response  // 直接返回响应
  },
  
  // ✅ 错误响应
  error => {
    if (error.response) {
      const { status, data } = error.response
      
      // 401: token 过期
      if (status === 401) {
        localStorage.removeItem('token')
        router.push('/login')  // 跳转登录页
        ElMessage.error('登录已过期，请重新登录')
      }
      // 403: 无权限
      else if (status === 403) {
        ElMessage.error('没有权限访问')
      }
      // 500: 服务器错误
      else if (status >= 500) {
        ElMessage.error('服务器错误，请稍后重试')
      }
      // 其他错误
      else {
        ElMessage.error(data.message || '请求失败')
      }
    } else {
      // 网络错误
      ElMessage.error('网络错误，请检查网络连接')
    }
    
    return Promise.reject(error)  // 抛出错误
  }
)
```

**匹配时机**：收到 HTTP 响应后执行

---

## 第十一步：返回数据到组件

**文件**：`frontend/src/views/Products.vue`

**数据返回过程**：

```javascript
// API 函数返回 Promise
const res = await getProductList({ page: 1, page_size: 12 })

// res 的结构：
// {
//   data: {                    // Axios 响应数据
//     code: 0,
//     message: "查询成功",
//     total: 100,
//     data: [商品列表]         // 后端返回的数据字段
//   },
//   status: 200,
//   headers: {...}
// }

// ✅ 检查响应码
if (res.data.code === 0) {
  // ✅ 更新响应式数据
  products.value = res.data.data
  total.value = res.data.total
}
```

**匹配时机**：API 请求完成后执行

---

## 第十二步：Vue 自动更新视图

**Vue 响应式系统**：

```vue
<template>
  <!-- ✅ Vue 自动监听 products 的变化 -->
  <el-col v-for="product in products" :key="product.id">
    {{ product.title }}
  </el-col>
</template>
```

**更新过程**：
1. ✅ `products.value = res.data.data` 更新数据
2. ✅ Vue 检测到数据变化
3. ✅ 自动重新渲染模板
4. ✅ 用户看到商品列表

**匹配时机**：响应式数据变化时自动执行

---

## 📊 完整文件匹配顺序

### 场景：访问商品列表页

| 步骤 | 文件 | 作用 | 执行时机 |
|------|------|------|---------|
| 1 | `main.js` | 应用启动 | 应用启动时 |
| 2 | `App.vue` | 根组件 | 应用启动后 |
| 3 | `router/index.js` | 路由匹配 | URL 变化时 |
| 4 | `components/Layout.vue` | 布局组件 | 路由匹配后 |
| 5 | `views/Products.vue` | 页面组件 | 路由匹配后 |
| 6 | `api/product.js` | API 函数 | 组件调用时 |
| 7 | `api/config.js` | API 配置 | 创建 request 时 |
| 8 | `api/request.js` | HTTP 封装 | 每次请求时 |
| 9 | `vite.config.js` | 代理配置 | 开发环境请求时 |
| 10 | 后端服务 | 处理请求 | 收到请求时 |
| 11 | `api/request.js` | 响应拦截 | 收到响应时 |
| 12 | `views/Products.vue` | 更新视图 | 数据更新时 |

---

## 🔑 关键文件说明

### 1. main.js - 应用入口
- **作用**：创建 Vue 应用，注册插件
- **执行时机**：应用启动时执行一次
- **关键代码**：`app.use(router)`, `app.mount('#app')`

### 2. router/index.js - 路由配置
- **作用**：定义路由规则，路由守卫
- **执行时机**：URL 变化时执行
- **关键代码**：`routes` 数组，`beforeEach` 守卫

### 3. api/request.js - HTTP 封装
- **作用**：统一处理请求和响应
- **执行时机**：每次 API 调用时执行
- **关键代码**：`interceptors.request`, `interceptors.response`

### 4. vite.config.js - 开发配置
- **作用**：代理 API 请求到后端
- **执行时机**：开发环境请求时执行
- **关键代码**：`proxy` 配置

---

## 💡 常见问题

### Q1: 为什么请求路径要去掉 `/api/v1`？
**A**: 因为 `request.js` 中已经配置了 `baseURL: '/api/v1'`，所以 API 函数中只需要写相对路径。

### Q2: 请求拦截器和响应拦截器的执行顺序？
**A**: 
1. 请求拦截器（发送请求前）
2. 发送 HTTP 请求
3. 响应拦截器（收到响应后）

### Q3: 路由守卫在哪里执行？
**A**: 在 `router/index.js` 的 `beforeEach` 中执行，在组件加载前执行。

### Q4: 如何调试请求流程？
**A**: 
- 查看浏览器 Console：有请求和响应的日志
- 查看 Network 标签：查看实际的 HTTP 请求
- 在关键位置添加 `console.log`

---

## ✅ 总结

**核心流程**：
1. **路由匹配** → 确定显示哪个页面组件
2. **组件挂载** → 执行 `onMounted`，调用 API
3. **API 调用** → 经过请求拦截器，发送 HTTP 请求
4. **代理转发** → Vite 代理到后端服务
5. **后端处理** → 返回 JSON 响应
6. **响应处理** → 经过响应拦截器，返回数据
7. **更新视图** → Vue 自动更新页面

**关键理解**：
- 路由 ≠ API 接口（路由是页面路径，API 是数据接口）
- 请求拦截器在发送前执行（添加 token）
- 响应拦截器在收到后执行（处理错误）
- Vite 代理只在开发环境生效

