# OSS 头像上传问题总结

## 📋 问题概览

在实现 OSS 头像上传功能过程中，遇到了以下几个主要问题：

---

## 🔴 问题 1：CORS 预检请求被拦截

### 问题描述
- **错误信息**：`Failed to load resource: net::ERR_FAILED`
- **CORS 错误**：`Response to preflight request doesn't pass access control check: No 'Access-Control-Allow-Origin' header is present on the requested resource`
- **现象**：浏览器发送 OPTIONS 预检请求时被认证中间件拦截，导致 CORS 失败

### 根本原因
1. 浏览器发送 `multipart/form-data` 请求时，会先发送 OPTIONS 预检请求
2. 认证中间件检查 Token 时，拦截了 OPTIONS 请求
3. CORS 中间件虽然会处理 OPTIONS，但如果请求被更内层的中间件拦截，CORS 响应头就无法返回

### 解决方案
在认证中间件中添加 OPTIONS 请求的放行逻辑：

```go
// internal/common/middleware/auth.go
func Auth() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // OPTIONS 预检请求直接放行（由 CORS 中间件处理）
            if r.Method == http.MethodOptions {
                next.ServeHTTP(w, r)
                return
            }
            // ... 其他逻辑
        })
    }
}
```

### 改进 CORS 处理
同时改进了 CORS 中间件对 `null` origin（file:// 协议）的处理：

```go
// internal/common/middleware/cors.go
if isOriginAllowed(origin, config.AllowedOrigins) || origin == "" || origin == "null" {
    if origin == "" || origin == "null" {
        w.Header().Set("Access-Control-Allow-Origin", "*")
    }
    // ...
}
```

---

## 🔴 问题 2：OSS 文件被下载而不是预览

### 问题描述
- **现象**：上传到 OSS 的图片在浏览器中打开时，会触发下载而不是直接显示
- **用户体验**：无法直接在浏览器中查看头像

### 根本原因
OSS 上传时没有设置正确的 HTTP 响应头：
- 缺少 `Content-Type`：浏览器不知道文件类型
- 缺少 `Content-Disposition: inline`：浏览器默认下载

### 解决方案
在上传时设置正确的 HTTP 响应头：

```go
// internal/common/oss/oss.go
func (o *OSSClient) UploadAvatar(userID string, file io.Reader, filename string) (string, error) {
    // ...
    
    // 根据文件扩展名设置 Content-Type
    contentType := getContentType(ext)
    options := []oss.Option{
        oss.ContentType(contentType),     // 设置 Content-Type，让浏览器直接显示图片
        oss.ContentDisposition("inline"), // 设置为 inline，让浏览器预览而不是下载
    }
    err := o.bucket.PutObject(objectName, file, options...)
    
    // ...
}

// 根据文件扩展名返回对应的 Content-Type
func getContentType(ext string) string {
    switch ext {
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".png":
        return "image/png"
    case ".gif":
        return "image/gif"
    case ".webp":
        return "image/webp"
    case ".bmp":
        return "image/bmp"
    case ".svg":
        return "image/svg+xml"
    default:
        return "image/jpeg"
    }
}
```

---

## 🔴 问题 3：OSS ACL 权限设置失败

### 问题描述
- **错误信息**：`oss: service returned error: StatusCode=403, ErrorCode=AccessDenied, ErrorMessage="Put public object acl is not allowed"`
- **现象**：文件上传成功，但设置 ACL 为公共读时失败

### 根本原因
阿里云 OSS Bucket 开启了**"阻止公共访问"**功能，该功能会阻止设置对象 ACL 为公共读。

### 解决方案

#### 方案 1：关闭"阻止公共访问"功能（推荐）
1. 登录阿里云 OSS 控制台
2. 进入 Bucket 管理 → 权限管理 → 阻止公共访问
3. 关闭该功能（或只关闭"阻止公共读"选项）

#### 方案 2：分离上传和 ACL 设置
将 ACL 设置分离，提供更清晰的错误信息：

```go
// internal/common/oss/oss.go
// 先上传文件
err := o.bucket.PutObject(objectName, file, options...)
if err != nil {
    return "", fmt.Errorf("上传到 OSS 失败: %w", err)
}

// 然后设置 ACL（如果失败，提供明确提示）
err = o.bucket.SetObjectACL(objectName, oss.ACLPublicRead)
if err != nil {
    return "", fmt.Errorf("设置文件 ACL 为公共读失败: %w (请在 OSS 控制台关闭'阻止公共访问'功能)", err)
}
```

---

## 🔴 问题 4：前端响应判断逻辑错误

### 问题描述
- **现象**：后端返回成功（200 OK），但前端显示"上传失败:上传成功"
- **控制台**：响应数据包含 `message: "上传成功"` 和 `avatar_url`

### 根本原因
1. Protobuf 生成的 JSON tag 使用了 `omitempty`
2. 当 `code: 0`（零值）时，JSON 序列化会**省略**该字段
3. 前端检查 `result.code === 0` 时，因为字段不存在，返回 `undefined`
4. `undefined === 0` 为 `false`，进入失败分支

### 解决方案
改进前端判断逻辑，即使 `code` 字段缺失也能正确识别成功：

```javascript
// test/avatar-upload-test.html
const result = await response.json();

// 判断成功：code === 0 或者有 avatar_url 且 message 包含"成功"
const isSuccess = result.code === 0 || 
                 (result.avatar_url && (result.message && result.message.includes('成功')));

if (isSuccess) {
    // 上传成功
    showResult('✅ 上传成功！', 'success');
    // ...
} else {
    // 上传失败
    showResult(`❌ 上传失败: ${result.message || '未知错误'}`, 'error');
}
```

---

## 🔴 问题 5：路由注册顺序

### 问题描述
- **现象**：自定义 HTTP 路由可能被 gRPC-Gateway 路由覆盖
- **潜在影响**：`/api/v1/users/avatar` 可能无法正确匹配

### 解决方案
将自定义路由注册放在 gRPC-Gateway 之前，确保优先匹配：

```go
// cmd/user-service/main.go
// 注册自定义HTTP路由（头像上传）- 必须在 gRPC-Gateway 之前注册
srv.AddRoute("/api/v1/users/avatar", userServiceHandler.UploadAvatarHTTP)

// 然后注册 HTTP 网关处理器
if err := srv.RegisterHTTPGateway(userv1.RegisterUserServiceHandlerFromEndpoint); err != nil {
    log.Fatalf("failed to register user service gateway: %v", err)
}
```

---

## 📝 经验总结

### 1. CORS 处理
- OPTIONS 预检请求必须在所有认证逻辑之前处理
- 需要同时支持浏览器环境和 `file://` 协议（开发测试场景）

### 2. OSS 配置
- 上传时务必设置正确的 `Content-Type` 和 `Content-Disposition`
- 了解并正确配置 Bucket 的权限策略（阻止公共访问、ACL 等）

### 3. 前端响应处理
- 注意 Protobuf 生成的 JSON 可能省略零值字段
- 使用更健壮的判断逻辑，不仅依赖单一字段

### 4. 路由设计
- 自定义路由应该在自动生成的路由之前注册
- 确保路由优先级正确

---

## ✅ 最终实现方案

采用**后端代理上传**方案：
1. 前端通过 `multipart/form-data` 上传文件到后端
2. 后端接收文件，上传到 OSS
3. 后端设置正确的 HTTP 响应头（Content-Type、Content-Disposition）
4. 后端设置 OSS 对象 ACL 为公共读
5. 后端更新数据库中的头像 URL
6. 返回成功响应给前端

**优势**：
- 简单直接，适合企业环境
- 可以在后端统一处理文件校验、格式转换等
- 不需要前端直接访问 OSS，更安全

---

## 🔗 相关文件

- `internal/common/oss/oss.go` - OSS 上传实现
- `internal/user-service/handler/user-service.go` - HTTP Handler
- `internal/user-service/service/user-service.go` - 业务逻辑
- `internal/common/middleware/auth.go` - 认证中间件（CORS 修复）
- `internal/common/middleware/cors.go` - CORS 中间件
- `test/avatar-upload-test.html` - 前端测试页面

