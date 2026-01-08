package models

import (
	"GinChat/utils"
	"context"
	"time"
)

// SetUserOnlineInfo 设置在线用户到redis缓存
func SetUserOnlineInfo(key string, val []byte, timeTTL time.Duration) {
	ctx := context.Background()
	utils.RedisCluster.Set(ctx, key, val, timeTTL)
}
