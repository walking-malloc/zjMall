# 前后端接口对照文档

本文档详细列出了前端 API 调用与后端 proto 定义的对应关系。

## 配置说明

- **前端 Base URL**: `/api/v1` (定义在 `frontend/src/api/config.js`)
- **后端 Gateway**: 所有接口通过 `/api/v1` 前缀访问

---

## 商品服务 (Product Service)

### 1. 获取商品列表

**前端调用** (`frontend/src/api/product.js`):
```javascript
getProductList(params)
GET /product/products
```

**后端定义** (`api/proto/product/product.proto`):
```protobuf
rpc ListProducts(ListProductsRequest) returns (ListProductsResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/products"
  };
}
```

**实际请求路径**: `/api/v1/product/products` ✅

**响应字段**:
- Proto 定义: `ListProductsResponse.data` (repeated ProductInfo)
- 前端使用: `res.data.data` ✅

---

### 2. 获取商品详情

**前端调用** (`frontend/src/api/product.js`):
```javascript
getProductDetail(id)
GET /product/products/{id}
```

**后端定义** (`api/proto/product/product.proto`):
```protobuf
rpc GetProduct(GetProductRequest) returns (GetProductResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/products/{product_id}"
  };
}
```

**实际请求路径**: `/api/v1/product/products/{product_id}` ✅

**响应字段**:
- Proto 定义: `GetProductResponse.product` (ProductInfo)
- 前端使用: `res.data.product` (需要修复为 `res.data.product`)

---

### 3. 搜索商品

**前端调用** (`frontend/src/api/product.js`):
```javascript
searchProducts(keyword, params)
GET /product/search?keyword=xxx&page=1&page_size=20
```

**后端定义** (`api/proto/product/product.proto`):
```protobuf
rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/search"
  };
}
```

**实际请求路径**: `/api/v1/product/search` ✅

**响应字段**:
- Proto 定义: `SearchProductsResponse.products` (repeated ProductInfo)
- 前端使用: `res.data.products` (已修复)

---

### 4. 获取类目列表

**前端调用** (`frontend/src/api/product.js`):
```javascript
getCategoryList()
GET /product/categories
```

**后端定义** (`api/proto/product/product.proto`):
```protobuf
rpc ListCategories(ListCategoriesRequest) returns (ListCategoriesResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/categories"
  };
}
```

**实际请求路径**: `/api/v1/product/categories` ✅

**响应字段**:
- Proto 定义: `ListCategoriesResponse.data` (repeated CategoryInfo)
- 前端使用: `res.data.data` ✅

---

### 5. 获取品牌列表

**前端调用** (`frontend/src/api/product.js`):
```javascript
getBrandList(params)
GET /product/brands?page=1&page_size=20
```

**后端定义** (`api/proto/product/product.proto`):
```protobuf
rpc ListBrands(ListBrandsRequest) returns (ListBrandsResponse) {
  option (google.api.http) = {
    get: "/api/v1/product/brands"
  };
}
```

**实际请求路径**: `/api/v1/product/brands` ✅

**响应字段**:
- Proto 定义: `ListBrandsResponse.data` (repeated BrandInfo)
- 前端使用: `res.data.data` ✅

---

## 用户服务 (User Service)

### 1. 用户注册

**前端调用** (`frontend/src/api/user.js`):
```javascript
register(phone, password, smsCode)
POST /users/register
Body: { phone, password, confirm_password, sms_code }
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse) {
  option (google.api.http) = {
    post: "/api/v1/users/register"
    body: "*"
  };
}
```

**实际请求路径**: `/api/v1/users/register` ✅

**响应字段**:
- Proto 定义: `RegisterResponse.data` (RegisterData: { user, token })
- 前端使用: `res.data.data` ✅

---

### 2. 用户登录

**前端调用** (`frontend/src/api/user.js`):
```javascript
login(phone, password)
POST /users/login
Body: { phone, password }
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc Login(LoginRequest) returns (LoginResponse) {
  option (google.api.http) = {
    post: "/api/v1/users/login"
    body: "*"
  };
}
```

**实际请求路径**: `/api/v1/users/login` ✅

**响应字段**:
- Proto 定义: `LoginResponse.data` (LoginData: { user, token, expires_at })
- 前端使用: `res.data.data` ✅

---

### 3. 验证码登录

**前端调用** (`frontend/src/api/user.js`):
```javascript
loginBySMS(phone, smsCode)
POST /users/login-by-sms
Body: { phone, sms_code }
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc LoginBySMS(LoginBySMSRequest) returns (LoginResponse) {
  option (google.api.http) = {
    post: "/api/v1/users/login-by-sms"
    body: "*"
  };
}
```

**实际请求路径**: `/api/v1/users/login-by-sms` ✅

**响应字段**:
- Proto 定义: `LoginResponse.data` (LoginData: { user, token, expires_at })
- 前端使用: `res.data.data` ✅

---

### 4. 获取短信验证码

**前端调用** (`frontend/src/api/user.js`):
```javascript
getSMSCode(phone)
POST /users/sms-code
Body: { phone }
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc GetSMSCode(GetSMSCodeRequest) returns (GetSMSCodeResponse) {
  option (google.api.http) = {
    post: "/api/v1/users/sms-code"
    body: "*"
  };
}
```

**实际请求路径**: `/api/v1/users/sms-code` ✅

**响应字段**:
- Proto 定义: `GetSMSCodeResponse` (无数据字段)
- 前端使用: 仅检查 `res.data.code` ✅

---

### 5. 获取用户信息

**前端调用** (`frontend/src/api/user.js`):
```javascript
getUserInfo(userId)
GET /users/{userId}
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc GetUser(GetUserRequest) returns (GetUserResponse) {
  option (google.api.http) = {
    get: "/api/v1/users/{user_id}"
  };
}
```

**实际请求路径**: `/api/v1/users/{user_id}` ✅

**响应字段**:
- Proto 定义: `GetUserResponse.data` (UserInfo)
- 前端使用: `res.data.data` ✅

---

### 6. 更新用户信息

**前端调用** (`frontend/src/api/user.js`):
```javascript
updateUserInfo(data)
PUT /users/me
Body: { nickname, email, gender, birthday }
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
  option (google.api.http) = {
    put: "/api/v1/users/{user_id}"
    body: "*"
  };
}
```

**⚠️ 不匹配**: 
- 前端使用: `/users/me`
- 后端定义: `/api/v1/users/{user_id}`

**需要修复**: 前端应该使用 `/users/{user_id}` 或后端需要添加 `/users/me` 接口

---

### 7. 登出

**前端调用** (`frontend/src/api/user.js`):
```javascript
logout()
POST /users/logout
```

**后端定义** (`api/proto/user/user.proto`):
```protobuf
rpc Logout(LogoutRequest) returns (LogoutResponse) {
  option (google.api.http) = {
    post: "/api/v1/users/logout"
    body: "*"
  };
}
```

**实际请求路径**: `/api/v1/users/logout` ✅

**响应字段**:
- Proto 定义: `LogoutResponse` (无数据字段)
- 前端使用: 仅检查 `res.data.code` ✅

---

## 总结

### ✅ 已匹配的接口
- 商品列表 (ListProducts)
- 商品搜索 (SearchProducts) - 已修复响应字段
- 类目列表 (ListCategories)
- 品牌列表 (ListBrands)
- 用户注册 (Register)
- 用户登录 (Login)
- 验证码登录 (LoginBySMS)
- 获取短信验证码 (GetSMSCode)
- 获取用户信息 (GetUser)
- 登出 (Logout)

### ✅ 已修复的接口
1. **商品详情 (GetProduct)**: 响应字段已修复为 `res.data.product` ✅
2. **商品搜索 (SearchProducts)**: 响应字段已修复为 `res.data.products` ✅
3. **更新用户信息 (UpdateUser)**: API 路径已修复，需要传入 `userId` 参数 ✅

### 📝 注意事项
- 所有前端 API 调用都会自动加上 `/api/v1` 前缀（通过 `baseURL` 配置）
- 响应字段名必须与 proto 定义完全一致
- 路径参数（如 `{product_id}`, `{user_id}`）需要正确传递

