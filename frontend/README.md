# zjMall 前端项目

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 接口
│   │   ├── config.js     # API 配置
│   │   ├── request.js     # Axios 实例和拦截器
│   │   ├── product.js    # 商品相关 API
│   │   └── user.js       # 用户相关 API
│   ├── components/       # 公共组件
│   │   ├── Header.vue    # 头部导航组件
│   │   ├── Footer.vue    # 页脚组件
│   │   └── Layout.vue     # 布局组件
│   ├── stores/           # Pinia 状态管理
│   │   ├── user.js       # 用户状态
│   │   └── cart.js       # 购物车状态
│   ├── views/            # 页面组件
│   │   ├── Home.vue      # 首页
│   │   ├── Products.vue  # 商品列表页
│   │   ├── ProductDetail.vue  # 商品详情页
│   │   ├── Cart.vue      # 购物车页
│   │   ├── Login.vue     # 登录页
│   │   ├── Register.vue  # 注册页
│   │   ├── Profile.vue   # 个人中心
│   │   ├── Orders.vue    # 订单列表
│   │   ├── Addresses.vue # 收货地址管理
│   │   └── NotFound.vue  # 404 页面
│   ├── router/           # 路由配置
│   │   └── index.js
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── package.json
└── vite.config.js
```

## 功能特性

### ✅ 已实现功能

1. **用户系统**
   - 用户注册/登录
   - 验证码登录
   - 用户信息管理
   - 个人中心

2. **商品系统**
   - 商品列表展示（分页、筛选、排序）
   - 商品详情页（SKU选择、数量选择）
   - 商品搜索
   - 分类导航
   - 品牌筛选

3. **购物车**
   - 添加商品到购物车
   - 购物车商品管理（增删改）
   - 购物车数量统计
   - 本地存储持久化

4. **订单系统**（UI已完成，待后端接口）
   - 订单列表
   - 订单状态筛选
   - 订单详情

5. **地址管理**（UI已完成，待后端接口）
   - 收货地址列表
   - 新增/编辑地址
   - 删除地址
   - 设置默认地址

6. **公共组件**
   - 统一的 Header 导航
   - Footer 页脚
   - Layout 布局组件

### 🚧 待完善功能

1. **订单结算**
   - 结算页面
   - 支付集成

2. **商品详情**
   - SKU 完整实现（需要后端支持）
   - 商品图片轮播
   - 商品评价

3. **用户体验优化**
   - 加载动画
   - 骨架屏
   - 错误边界处理

## 技术栈

- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite
- **UI 组件库**: Element Plus
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **HTTP 客户端**: Axios
- **图标**: @element-plus/icons-vue

## 开发指南

### 安装依赖

```bash
cd frontend
npm install
```

### 启动开发服务器

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## API 配置

前端 API 基础路径配置在 `src/api/config.js` 中：

```javascript
export const API_BASE_URL = '/api/v1'
```

开发环境通过 Vite 代理配置（`vite.config.js`）转发到后端服务。

## 状态管理

### User Store (`stores/user.js`)
- `token`: 用户 token
- `userInfo`: 用户信息
- `isLoggedIn`: 登录状态
- `loginUser()`: 登录
- `registerUser()`: 注册
- `logout()`: 登出

### Cart Store (`stores/cart.js`)
- `items`: 购物车商品列表
- `totalCount`: 总数量
- `totalPrice`: 总金额
- `addItem()`: 添加商品
- `updateQuantity()`: 更新数量
- `removeItem()`: 删除商品
- `clearCart()`: 清空购物车

## 路由说明

- `/` - 首页
- `/product/products` - 商品列表
- `/product/products/:id` - 商品详情
- `/cart` - 购物车
- `/login` - 登录
- `/register` - 注册
- `/profile` - 个人中心
- `/orders` - 订单列表
- `/addresses` - 收货地址
- `/*` - 404 页面

## 注意事项

1. 购物车数据存储在 `localStorage` 中，页面刷新后数据会保留
2. 用户 token 存储在 `localStorage` 中
3. 部分功能（订单、地址）的 API 接口待后端实现
4. 商品 SKU 选择功能需要后端返回完整的 SKU 数据结构

## 待办事项

- [ ] 实现订单结算页面
- [ ] 完善商品 SKU 选择逻辑
- [ ] 添加商品评价功能
- [ ] 优化移动端适配
- [ ] 添加加载动画和骨架屏
- [ ] 完善错误处理机制
