# 前端架构设计指南

本文档详细说明前端项目的设计思路、路由与后端接口的对应关系、参数传递方式，以及每一步的设计考虑。

---

## 🎯 第一步：理解后端接口定义

### 1.1 查看后端 Proto 文件

**目标**：了解后端提供了哪些接口，接口的路径、参数、返回值是什么。

**查看文件**：
- `api/proto/product/product.proto` - 商品服务接口
- `api/proto/user/user.proto` - 用户服务接口

**关键信息提取**：
```protobuf
// 示例：商品列表接口
rpc ListProducts(ListProductsRequest) returns (ListProductsResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/products"  // HTTP 路径
  };
}

message ListProductsRequest {
  int32 page = 1;           // 请求参数
  int32 page_size = 2;
  string category_id = 3;
}

message ListProductsResponse {
  int32 code = 1;
  string message = 2;
  int64 total = 3;
  repeated ProductInfo data = 4;  // 返回字段名是 data
}
```

**设计考虑**：
- ✅ 记录接口路径：`/api/v1/product/products`
- ✅ 记录请求参数：`page`, `page_size`, `category_id`
- ✅ 记录响应字段：`code`, `message`, `total`, `data`
- ✅ 注意：响应字段名是 `data`，不是 `products`

---

## 🔧 第二步：设计 API 层（api/ 目录）

### 2.1 创建 API 配置文件

**文件**：`frontend/src/api/config.js`

```javascript
export const API_BASE_URL = '/api/v1'  // 所有接口的前缀
export const API_TIMEOUT = 10000       // 请求超时时间
```

**设计考虑**：
- ✅ 统一管理 API 基础路径
- ✅ 方便切换开发/生产环境
- ✅ 统一超时配置

---

### 2.2 创建 HTTP 请求封装

**文件**：`frontend/src/api/request.js`

**核心功能**：
1. **创建 Axios 实例**：
   ```javascript
   const request = axios.create({
     baseURL: API_BASE_URL,  // 自动添加 /api/v1 前缀
     timeout: API_TIMEOUT
   })
   ```

2. **请求拦截器**（发送请求前）：
   ```javascript
   request.interceptors.request.use(config => {
     // 自动添加 token
     const token = localStorage.getItem('token')
     if (token) {
       config.headers.Authorization = `Bearer ${token}`
     }
     return config
   })
   ```

3. **响应拦截器**（收到响应后）：
   ```javascript
   request.interceptors.response.use(
     response => response,  // 成功：直接返回
     error => {
       // 失败：统一错误处理
       if (error.response?.status === 401) {
         // token 过期，跳转登录
         router.push('/login')
       }
       return Promise.reject(error)
     }
   )
   ```

**设计考虑**：
- ✅ **统一处理**：所有请求自动添加 token
- ✅ **统一错误处理**：401 自动跳转登录，500 显示错误提示
- ✅ **减少重复代码**：每个 API 调用不需要单独处理 token 和错误

---

### 2.3 创建具体的 API 函数

**文件**：`frontend/src/api/product.js`

**设计思路**：一个后端接口对应一个前端函数

```javascript
import request from './request'

// 函数名：语义化，一看就知道做什么
// 参数：对应后端 proto 的请求参数
// 返回值：Promise，包含后端响应
export function getProductList(params) {
  // 路径：去掉 /api/v1 前缀（request.js 会自动添加）
  // params：GET 请求的查询参数
  return request.get('/product/products', { params })
}

export function getProductDetail(id) {
  // 路径参数：使用模板字符串
  return request.get(`/product/products/${id}`)
}

export function searchProducts(keyword, params) {
  // 多个参数：合并到 params 对象
  return request.get('/product/search', {
    params: {
      keyword,
      ...params  // 展开其他参数（page, page_size 等）
    }
  })
}
```

**设计考虑**：
- ✅ **路径对应**：前端路径 = 后端路径 - `/api/v1`
- ✅ **参数映射**：前端函数参数 → 后端请求参数
- ✅ **函数命名**：清晰、语义化
- ✅ **统一返回**：都返回 Promise，方便使用 async/await

---

## 🗺️ 第三步：设计路由（router/ 目录）

### 3.1 路由与后端接口的关系

**重要理解**：
- **路由 ≠ 后端接口**
- **路由是前端页面路径，后端接口是 API 路径**

**对应关系示例**：

| 前端路由 | 对应页面 | 页面使用的后端接口 |
|---------|---------|------------------|
| `/` | 首页 | `getProductList()` |
| `/product/products` | 商品列表页 | `getProductList()`, `getCategoryList()` |
| `/product/products/:id` | 商品详情页 | `getProductDetail(id)` |
| `/login` | 登录页 | `login()`, `getSMSCode()` |

### 3.2 路由配置

**文件**：`frontend/src/router/index.js`

```javascript
const routes = [
  {
    path: '/',                    // 浏览器地址栏路径
    component: Layout,            // 使用的布局组件
    children: [
      {
        path: '',                 // 子路由路径
        name: 'Home',            // 路由名称（用于编程式导航）
        component: Home,          // 对应的页面组件
        meta: { requiresAuth: false }  // 元信息：是否需要登录
      },
      {
        path: 'product/products/:id',  // 动态路由参数
        name: 'ProductDetail',
        component: ProductDetail,
        meta: { requiresAuth: false }
      }
    ]
  }
]
```

**设计考虑**：
- ✅ **路径设计**：清晰、符合 RESTful 规范
- ✅ **嵌套路由**：使用 Layout 组件统一头部和底部
- ✅ **路由守卫**：根据 `meta.requiresAuth` 判断是否需要登录
- ✅ **动态参数**：使用 `:id` 传递商品 ID

---

## 📦 第四步：设计状态管理（stores/ 目录）

### 4.1 为什么需要状态管理？

**问题场景**：
- 用户登录后，多个页面都需要显示用户信息
- 购物车数据需要在多个页面共享

**解决方案**：使用 Pinia Store 统一管理全局状态

### 4.2 Store 设计

**文件**：`frontend/src/stores/user.js`

```javascript
export const useUserStore = defineStore('user', () => {
  // 1. 定义状态
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(null)

  // 2. 计算属性
  const isLoggedIn = computed(() => !!token.value)

  // 3. 定义方法
  async function loginUser(phone, password) {
    const res = await login(phone, password)  // 调用 API
    if (res.data.code === 0) {
      setToken(res.data.data.token)           // 保存 token
      setUserInfo(res.data.data.user)         // 保存用户信息
    }
  }

  return { token, userInfo, isLoggedIn, loginUser }
})
```

**设计考虑**：
- ✅ **状态持久化**：token 保存到 localStorage
- ✅ **统一管理**：所有用户相关状态都在这里
- ✅ **响应式**：使用 `ref` 和 `computed` 实现响应式

---

## 🎨 第五步：设计页面组件（views/ 目录）

### 5.1 页面组件结构

**文件**：`frontend/src/views/Products.vue`

```vue
<template>
  <!-- 1. HTML 模板：页面结构 -->
  <div class="products-page">
    <el-input v-model="searchKeyword" />  <!-- 双向绑定 -->
    <el-button @click="handleSearch">搜索</el-button>
    <div v-for="product in products" :key="product.id">
      {{ product.title }}
    </div>
  </div>
</template>

<script setup>
// 2. JavaScript 逻辑：数据和方法
import { ref, onMounted } from 'vue'
import { getProductList } from '@/api/product'  // 导入 API 函数

const products = ref([])  // 响应式数据
const searchKeyword = ref('')

// 加载商品列表
const loadProducts = async () => {
  const res = await getProductList({ page: 1, page_size: 12 })
  if (res.data.code === 0) {
    products.value = res.data.data  // 注意：字段名是 data
  }
}

// 页面加载时执行
onMounted(() => {
  loadProducts()
})
</script>

<style scoped>
/* 3. CSS 样式：页面样式 */
.products-page { padding: 20px; }
</style>
```

**设计考虑**：
- ✅ **组件化**：每个页面是一个独立组件
- ✅ **数据驱动**：通过 `ref` 管理状态，数据变化自动更新视图
- ✅ **生命周期**：`onMounted` 时加载数据
- ✅ **API 调用**：在组件中调用 API 函数

---

## 🔄 完整的数据流

### 示例：商品列表页加载商品

```
1. 用户访问 /product/products
   ↓
2. 路由匹配到 Products.vue 组件
   ↓
3. 组件挂载，执行 onMounted()
   ↓
4. 调用 loadProducts() 函数
   ↓
5. 调用 getProductList({ page: 1, page_size: 12 })
   ↓
6. request.get('/product/products', { params: {...} })
   ↓
7. 请求拦截器：自动添加 token 到 header
   ↓
8. 实际请求：GET /api/v1/product/products?page=1&page_size=12
   ↓
9. Vite 代理：转发到 http://localhost:8082/api/v1/product/products
   ↓
10. 后端处理：返回 JSON 响应
    {
      "code": 0,
      "message": "查询成功",
      "total": 100,
      "data": [商品列表]
    }
   ↓
11. 响应拦截器：检查 code，处理错误
   ↓
12. 组件接收响应：res.data.data
   ↓
13. 更新响应式数据：products.value = res.data.data
   ↓
14. Vue 自动更新视图：显示商品列表
```

---

## 📋 参数传递详解

### 1. GET 请求参数传递

**场景**：获取商品列表，需要传递分页参数

**前端代码**：
```javascript
// API 函数定义
export function getProductList(params) {
  return request.get('/product/products', { params })
}

// 组件中调用
const params = {
  page: 1,
  page_size: 12,
  category_id: '123'
}
const res = await getProductList(params)
```

**实际请求**：
```
GET /api/v1/product/products?page=1&page_size=12&category_id=123
```

**设计考虑**：
- ✅ 使用 `params` 对象，Axios 自动转换为查询参数
- ✅ 参数名与后端 proto 定义一致

---

### 2. POST 请求参数传递

**场景**：用户登录，需要传递手机号和密码

**前端代码**：
```javascript
// API 函数定义
export function login(phone, password) {
  return request.post('/users/login', {
    phone,
    password
  })
}

// 组件中调用
const res = await login('13800138000', '123456')
```

**实际请求**：
```
POST /api/v1/users/login
Content-Type: application/json

{
  "phone": "13800138000",
  "password": "123456"
}
```

**设计考虑**：
- ✅ POST 请求参数放在请求体中（body）
- ✅ Axios 自动序列化为 JSON
- ✅ 参数名与后端 proto 定义一致

---

### 3. 路径参数传递

**场景**：获取商品详情，商品 ID 在路径中

**前端代码**：
```javascript
// API 函数定义
export function getProductDetail(id) {
  return request.get(`/product/products/${id}`)
}

// 组件中调用
const productId = route.params.id  // 从路由获取
const res = await getProductDetail(productId)
```

**实际请求**：
```
GET /api/v1/product/products/123456
```

**设计考虑**：
- ✅ 使用模板字符串拼接路径参数
- ✅ 参数从路由中获取（`route.params.id`）

---

### 4. 路由参数传递

**场景**：从商品列表页跳转到商品详情页

**前端代码**：
```javascript
// 方式1：使用路由名称
router.push({ name: 'ProductDetail', params: { id: '123' } })

// 方式2：使用路径
router.push('/product/products/123')

// 方式3：使用查询参数
router.push({ path: '/product/products', query: { category_id: '123' } })
```

**设计考虑**：
- ✅ `params`：路径参数，会出现在 URL 路径中
- ✅ `query`：查询参数，会出现在 URL 查询字符串中

---

## 🎯 设计步骤总结

### 第一步：理解后端接口
1. 查看 proto 文件，了解接口定义
2. 记录接口路径、参数、返回值
3. 特别注意响应字段名

### 第二步：设计 API 层
1. 创建 `config.js` 配置基础路径
2. 创建 `request.js` 封装 HTTP 请求
3. 创建 `product.js`、`user.js` 等 API 函数文件
4. 每个后端接口对应一个前端函数

### 第三步：设计路由
1. 规划页面路由结构
2. 配置路由表（路径、组件、权限）
3. 设置路由守卫（登录验证）

### 第四步：设计状态管理
1. 识别需要全局共享的状态（用户信息、购物车）
2. 创建对应的 Store
3. 定义状态、计算属性、方法

### 第五步：设计页面组件
1. 创建页面组件文件
2. 在组件中调用 API 函数
3. 处理响应数据，更新视图
4. 处理用户交互（点击、输入等）

---

## 🔍 关键设计原则

### 1. 单一职责
- API 层：只负责 HTTP 请求
- 路由层：只负责页面跳转
- Store 层：只负责状态管理
- 组件层：只负责 UI 展示和用户交互

### 2. 统一规范
- API 路径：统一使用 `/api/v1` 前缀
- 错误处理：统一在响应拦截器中处理
- Token 管理：统一在请求拦截器中添加

### 3. 数据流向清晰
```
用户操作 → 组件方法 → API 函数 → HTTP 请求 → 后端接口
                                                      ↓
视图更新 ← 响应式数据 ← API 响应 ← HTTP 响应 ← 后端响应
```

### 4. 响应式数据
- 使用 `ref()` 定义响应式变量
- 数据变化自动更新视图
- 不需要手动操作 DOM

---

## 📝 常见问题

### Q1: 为什么 API 路径要去掉 `/api/v1`？
**A**: 因为 `request.js` 中已经配置了 `baseURL: '/api/v1'`，所以 API 函数中只需要写相对路径。

### Q2: 如何知道响应数据的字段名？
**A**: 查看后端 proto 文件中的响应 message 定义，字段名必须完全一致。

### Q3: 路由参数和查询参数的区别？
**A**: 
- 路径参数：`/product/products/123` → `route.params.id = '123'`
- 查询参数：`/product/products?category_id=123` → `route.query.category_id = '123'`

### Q4: 什么时候用 Store，什么时候用组件内的 ref？
**A**: 
- Store：需要跨组件共享的数据（用户信息、购物车）
- ref：只在当前组件使用的数据（表单输入、临时状态）

---

## ✅ 检查清单

设计前端功能时，按以下步骤检查：

- [ ] 后端接口是否已定义？（查看 proto 文件）
- [ ] API 函数是否已创建？（api/ 目录）
- [ ] 路由是否已配置？（router/index.js）
- [ ] 页面组件是否已创建？（views/ 目录）
- [ ] 是否需要 Store？（需要跨组件共享的数据）
- [ ] 参数传递是否正确？（路径参数、查询参数、请求体）
- [ ] 响应字段名是否正确？（与 proto 定义一致）
- [ ] 错误处理是否完善？（API 调用失败时的处理）

---

## 🎓 总结

**核心思想**：
1. **前后端分离**：前端路由 ≠ 后端接口
2. **分层设计**：API 层 → 路由层 → Store 层 → 组件层
3. **数据驱动**：使用响应式数据，自动更新视图
4. **统一规范**：路径、参数、错误处理都统一管理

**关键文件**：
- `api/config.js` - API 配置
- `api/request.js` - HTTP 封装
- `api/product.js` - 商品 API
- `router/index.js` - 路由配置
- `stores/user.js` - 用户状态
- `views/Products.vue` - 页面组件

**记住**：前端代码的设计思路是**自顶向下**（从页面需求 → API 调用 → 后端接口），而实现是**自底向上**（从 API 层 → 路由层 → 组件层）。

