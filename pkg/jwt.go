package pkg

import (
	"errors"
	"time"
	"zjMall/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// 包级别的配置（启动时初始化）
var (
	jwtSecret           []byte
	defaultExpiresIn    time.Duration
	rememberMeExpiresIn time.Duration
)

// Claims JWT 载荷结构
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// InitJWT 初始化 JWT 配置（在 main.go 启动时调用一次）
func InitJWT(cfg *config.JWTConfig) {
	jwtSecret = []byte(cfg.Secret)
	defaultExpiresIn = cfg.ExpiresIn
	rememberMeExpiresIn = cfg.RememberMeExpiresIn
}

// GenerateJWT 生成 JWT Token
// userID: 用户 ID
// expiresIn: Token 过期时长（例如：7 * 24 * time.Hour 表示 7 天）
// 返回: token 字符串、过期时间戳、错误
func GenerateJWT(userID string, expiresIn time.Duration) (string, int64, error) {
	// 计算过期时间
	expiresAt := time.Now().Add(expiresIn)

	// 创建 Claims
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),  // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()), // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()), // 生效时间
			Issuer:    "zjMall",                       // 签发者
		},
	}

	// 创建 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并获取完整 Token 字符串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// GenerateJWTWithRememberMe 根据 RememberMe 生成 Token（便捷函数）
// userID: 用户 ID
// rememberMe: 是否记住我
// 返回: token 字符串、过期时间戳、错误
func GenerateJWTWithRememberMe(userID string, rememberMe bool) (string, int64, error) {
	var expiresIn time.Duration
	if rememberMe {
		expiresIn = rememberMeExpiresIn
	} else {
		expiresIn = defaultExpiresIn
	}
	return GenerateJWT(userID, expiresIn)
}

// VerifyJWT 验证 JWT Token
// tokenString: JWT Token 字符串
// 返回: 用户 ID 和错误
func VerifyJWT(tokenString string) (string, error) {
	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	// 验证 Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token")
}

// ParseToken 解析 Token（不验证过期，用于刷新 Token 等场景）
// 返回: Claims 和错误
func ParseToken(tokenString string) (*Claims, error) {
	// 解析但不验证过期
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return jwtSecret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
