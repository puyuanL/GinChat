// Package utils
/**
jwt method of token generate
暂时未启用
*/
package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 1. 定义常量（生产环境建议通过配置文件读取）
const (
	// JWT密钥（必须保密，生产环境建议用环境变量/密钥管理服务）
	SecretKey = "your-strong-secret-key-32bytes-long-12345678"
	// Token过期时间（例如2小时）
	TokenExpireDuration = 2 * time.Hour
	// 刷新Token过期时间（例如7天，可选）
	RefreshTokenExpireDuration = 7 * 24 * time.Hour
)

// 2. 自定义Claims（包含用户核心信息，避免敏感数据）
type CustomClaims struct {
	UserID               uint64 `json:"user_id"`  // 用户ID
	Username             string `json:"username"` // 用户名（非敏感）
	Role                 string `json:"role"`     // 用户角色（用于权限控制）
	jwt.RegisteredClaims        // 嵌入JWT标准Claims（包含过期时间等）
}

// 3. 生成访问Token
func GenerateToken(userID uint64, username, role string) (accessToken string, err error) {
	// 构造Claims
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			// 签发时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// 签发者
			Issuer: "your-app-name",
			// 受众（可选）
			Audience: jwt.ClaimStrings{"web", "mobile"},
		},
	}

	// 创建Token对象（使用HS256算法）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名生成Token字符串
	accessToken, err = token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("生成Token失败: %w", err)
	}
	return accessToken, nil
}

// 4. 生成刷新Token（可选，用于无感刷新）
func GenerateRefreshToken(userID uint64) (refreshToken string, err error) {
	// 刷新Token只保留核心用户ID，减少数据量
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpireDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "your-app-name",
		Subject:   fmt.Sprintf("%d", userID), // 用Subject存储用户ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err = token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("生成刷新Token失败: %w", err)
	}
	return refreshToken, nil
}

// 5. 解析并验证Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析Token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{}, // 传入自定义Claims的指针
		func(token *jwt.Token) (interface{}, error) {
			// 验证算法是否匹配
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
			}
			return []byte(SecretKey), nil
		},
	)

	// 处理解析错误
	if err != nil {
		// 区分过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Token已过期")
		}
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}

	// 验证Token有效并提取Claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// 验证刷新Token并获取用户ID（可选）
func ParseRefreshToken(refreshToken string) (userID uint64, err error) {
	token, err := jwt.ParseWithClaims(
		refreshToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
			}
			return []byte(SecretKey), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, errors.New("刷新Token已过期")
		}
		return 0, fmt.Errorf("解析刷新Token失败: %w", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// 解析Subject中的用户ID
		_, err = fmt.Sscanf(claims.Subject, "%d", &userID)
		if err != nil {
			return 0, errors.New("刷新Token中用户ID格式错误")
		}
		return userID, nil
	}

	return 0, errors.New("无效的刷新Token")
}
