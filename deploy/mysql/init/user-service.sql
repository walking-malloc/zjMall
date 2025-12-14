-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(26) not null PRIMARY KEY,
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
    last_login_at TIMESTAMP COMMENT '最后登录时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 收获地址表
CREATE TABLE IF NOT EXISTS addresses (
    id VARCHAR(26) NOT NULL PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(26) NOT NULL COMMENT '用户ID',
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