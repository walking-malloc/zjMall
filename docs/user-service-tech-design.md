# 用户服务技术方案文档

## 1. 技术栈选型

### 1.1 开发语言与框架
- **语言**：Go 1.21+
- **Web框架**：Gin（HTTP服务）
- **ORM**：GORM（数据库操作）
- **配置管理**：Viper（支持YAML/JSON/环境变量）

### 1.2 数据库
- **主数据库**：MySQL 8.0+
- **缓存**：Redis（用户信息缓存、Token存储）

### 1.3 其他组件
- **日志**：zap（结构化日志）
- **密码加密**：bcrypt
- **Token生成**：JWT（jwt-go）
- **ID生成**：雪花算法（snowflake）或UUID

---

## 2. 架构设计

### 2.1 分层架构
```
Handler层（HTTP接口）
    ↓
Service层（业务逻辑）
    ↓
Repository层（数据访问）
    ↓
Database（MySQL/Redis）
```

### 2.2 目录结构
```
cmd/user-service/
    └── main.go              # 服务入口

internal/user/
    ├── handler/             # HTTP处理器
    ├── service/             # 业务逻辑层
    ├── repository/          # 数据访问层
    └── model/              # 数据模型

pkg/
    ├── logger/             # 日志工具
    ├── jwt/               # JWT工具
    └── errors/            # 错误码定义
```

---

## 3. 数据库设计

### 3.1 用户表（users）
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    phone VARCHAR(11) UNIQUE NOT NULL COMMENT '手机号',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密）',
    nickname VARCHAR(50) COMMENT '昵称',
    avatar VARCHAR(255) COMMENT '头像URL',
    email VARCHAR(100) COMMENT '邮箱',
    gender TINYINT DEFAULT 0 COMMENT '性别：0-未设置，1-男，2-女',
    birthday DATE COMMENT '生日',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常，2-已锁定，3-已注销',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP COMMENT '最后登录时间',
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

### 3.2 收货地址表（addresses）
```sql
CREATE TABLE addresses (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    receiver_name VARCHAR(50) NOT NULL COMMENT '收货人姓名',
    receiver_phone VARCHAR(11) NOT NULL COMMENT '收货人手机号',
    province VARCHAR(50) NOT NULL COMMENT '省份',
    city VARCHAR(50) NOT NULL COMMENT '城市',
    district VARCHAR(50) NOT NULL COMMENT '区县',
    detail VARCHAR(200) NOT NULL COMMENT '详细地址',
    postal_code VARCHAR(6) COMMENT '邮政编码',
    is_default TINYINT DEFAULT 0 COMMENT '是否默认：0-否，1-是',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_default (user_id, is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收货地址表';
```

### 3.3 登录会话表（sessions，可选）
```sql
CREATE TABLE sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    token VARCHAR(255) UNIQUE NOT NULL COMMENT 'Token',
    device_info VARCHAR(100) COMMENT '设备信息',
    ip_address VARCHAR(50) COMMENT 'IP地址',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录会话表';
```

---

## 4. 接口设计

### 4.1 用户注册登录
- `POST /api/v1/users/register` - 注册
- `POST /api/v1/users/login` - 密码登录
- `POST /api/v1/users/login-by-sms` - 验证码登录
- `POST /api/v1/users/logout` - 登出
- `POST /api/v1/users/sms-code` - 获取验证码

### 4.2 用户信息
- `GET /api/v1/users/{userId}` - 查询用户信息
- `PUT /api/v1/users/{userId}` - 更新用户信息
- `GET /api/v1/users/{userId}/status` - 查询用户状态

### 4.3 收货地址
- `POST /api/v1/users/{userId}/addresses` - 添加地址
- `GET /api/v1/users/{userId}/addresses` - 查询地址列表
- `GET /api/v1/users/{userId}/addresses/{addressId}` - 查询单个地址
- `PUT /api/v1/users/{userId}/addresses/{addressId}` - 更新地址
- `DELETE /api/v1/users/{userId}/addresses/{addressId}` - 删除地址
- `PUT /api/v1/users/{userId}/addresses/{addressId}/default` - 设置默认地址

### 4.4 账号安全
- `PUT /api/v1/users/{userId}/password` - 修改密码
- `PUT /api/v1/users/{userId}/phone` - 绑定手机号

---

## 5. 关键技术实现

### 5.1 密码加密
- 使用 `bcrypt` 加密存储
- 加密强度：cost=10

### 5.2 Token生成与校验
- 使用 JWT（JSON Web Token）
- Payload包含：userId、exp（过期时间）、iat（签发时间）
- Token有效期：7天（默认），30天（记住我）
- 存储：Redis（key: `token:{token}`, value: userId, TTL: 过期时间）

### 5.3 验证码管理
- 存储：Redis（key: `sms_code:{phone}`, value: 验证码, TTL: 5分钟）
- 防刷：同一手机号60秒内只能获取一次（Redis key: `sms_limit:{phone}`, TTL: 60秒）

### 5.4 用户信息缓存
- 缓存策略：查询时先查Redis，未命中再查MySQL并写入缓存
- Redis key: `user:{userId}`, TTL: 1小时
- 更新用户信息时，删除缓存

### 5.5 地址列表缓存
- Redis key: `user_addresses:{userId}`, TTL: 30分钟
- 地址增删改时，删除缓存

### 5.6 账号锁定机制
- 密码错误5次后锁定30分钟
- Redis key: `user_lock:{userId}`, value: 锁定时间戳, TTL: 30分钟

---

## 6. 错误处理

### 6.1 统一错误码
```go
const (
    ErrCodeUserNotFound = 10001  // 用户不存在
    ErrCodePasswordWrong = 10002 // 密码错误
    ErrCodePhoneExists = 10003   // 手机号已注册
    ErrCodeSmsCodeWrong = 10004  // 验证码错误
    ErrCodeUserLocked = 10005    // 账号已锁定
    // ...
)
```

### 6.2 统一响应格式
```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

---

## 7. 性能优化

### 7.1 数据库优化
- 用户表phone字段建立唯一索引
- 地址表user_id建立索引
- 查询用户信息时使用主键（id）

### 7.2 缓存策略
- 用户信息缓存（Redis）
- 地址列表缓存（Redis）
- Token存储（Redis）

### 7.3 连接池配置
- MySQL连接池：最大连接数100，空闲连接10
- Redis连接池：最大连接数50

---

## 8. 安全方案

### 8.1 密码安全
- 使用bcrypt加密，不存储明文
- 密码复杂度校验（8-20位，包含字母和数字）

### 8.2 Token安全
- Token存储在Redis，支持主动失效
- 登出时删除Token
- Token过期时间合理设置

### 8.3 接口安全
- 敏感接口需要Token验证（中间件）
- 验证码防刷限制（60秒一次）
- 密码错误锁定机制

### 8.4 数据安全
- 手机号脱敏显示（138****5678）
- 敏感信息不记录日志

---

## 9. 部署方案

### 9.1 服务配置
- 端口：8081（可配置）
- 环境变量：数据库连接、Redis连接、JWT密钥

### 9.2 健康检查
- `GET /healthz` - 健康检查接口
- 检查MySQL和Redis连接状态

### 9.3 日志
- 使用zap结构化日志
- 日志级别：开发环境DEBUG，生产环境INFO
- 输出到文件和控制台

---

## 10. 开发计划

### 10.1 第一阶段（MVP）
1. 数据库表创建
2. 用户注册/登录（密码+验证码）
3. 用户信息查询/更新
4. 收货地址CRUD
5. 基础缓存

### 10.2 第二阶段
1. Token管理优化
2. 账号锁定机制
3. 密码修改
4. 手机号绑定

### 10.3 第三阶段
1. 性能优化
2. 安全加固
3. 监控告警

---

**文档版本：** v1.0  
**最后更新：** 2024-12-10

