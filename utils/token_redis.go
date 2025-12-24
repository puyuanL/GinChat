package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/google/uuid"
)

// 常量定义
const (
	// RedisAccessTokenPrefix Redis Key前缀
	RedisAccessTokenPrefix = "token:" // 访问Token Key: access_token:xxx-xxx-xxx
	// AccessTokenExpire Token过期时间
	AccessTokenExpire = 2 * time.Hour // 访问Token有效期2小时
)

// UserInfo Redis中存储的用户信息（可根据业务扩展）
type UserInfo struct {
	UserID   uint   `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	Role     string `json:"role"`     // 用户角色（admin/user）
	LoginAt  int64  `json:"login_at"` // 登录时间戳
}

// GenerateTokens 生成访问Token和刷新Token，并存储到Redis
func GenerateTokens(userID uint, username string) (accessToken string, err error) {
	// 1. 生成唯一的Token（UUIDv4保证唯一性）
	accessToken = uuid.NewString()

	// 2. 构造用户信息
	userInfo := UserInfo{
		UserID:   userID,
		Username: username,
		LoginAt:  time.Now().Unix(),
	}
	// 序列化用户信息（JSON格式存储）
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		return "", fmt.Errorf("序列化用户信息失败: %w", err)
	}

	ctx := context.Background()
	// 3. 存储访问Token到Redis（设置过期时间）
	accessTokenKey := fmt.Sprintf("%s%s", RedisAccessTokenPrefix, accessToken)
	if err := RedisClient.SetEX(ctx, accessTokenKey, userInfoJSON, AccessTokenExpire).Err(); err != nil {
		return "", fmt.Errorf("存储访问Token失败: %w", err)
	}

	return accessToken, nil
}

// VerifyAccessToken 验证访问Token并获取用户信息
func VerifyAccessToken(token string) (*UserInfo, error) {
	ctx := context.Background()

	tokenKey := fmt.Sprintf("%s%s", RedisAccessTokenPrefix, token)
	userInfoJSON, err := RedisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) { // Token不存在（过期/已登出/无效）
			return nil, errors.New("token invalid (Logout/Expire)")
		}
		return nil, fmt.Errorf("token: search Redis fail: %w", err)
	}

	// 3. 反序列化用户信息
	var userInfo UserInfo
	if err := json.Unmarshal([]byte(userInfoJSON), &userInfo); err != nil {
		return nil, fmt.Errorf("反序列化用户信息失败: %w", err)
	}

	return &userInfo, nil
}

// InvalidateToken 主动失效Token（登出/改密码/封禁时调用）
func InvalidateToken(token string) error {
	ctx := context.Background()
	tokenKey := fmt.Sprintf("%s%s", RedisAccessTokenPrefix, token)
	if err := RedisClient.Del(ctx, tokenKey).Err(); err != nil {
		return fmt.Errorf("invalid Token: %w", err)
	}
	return nil
}
