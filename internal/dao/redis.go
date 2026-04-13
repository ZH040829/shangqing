package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"shangqing/internal/config"
	"shangqing/internal/model"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg *config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping error: %w", err)
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// ----- Session 缓存 -----

// CacheMessages 缓存对话消息（最近 N 条）
func (r *Redis) CacheMessages(ctx context.Context, convID string, msgs []*model.Message) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("session:%s", convID)
	return r.client.Set(ctx, key, data, 30*time.Minute).Err()
}

// GetCachedMessages 获取缓存的对话消息
func (r *Redis) GetCachedMessages(ctx context.Context, convID string) ([]*model.Message, error) {
	key := fmt.Sprintf("session:%s", convID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var msgs []*model.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}

// InvalidateSession 清除会话缓存
func (r *Redis) InvalidateSession(ctx context.Context, convID string) error {
	key := fmt.Sprintf("session:%s", convID)
	return r.client.Del(ctx, key).Err()
}

// ----- Token 管理 -----

// SetUserToken 存储用户 Token
func (r *Redis) SetUserToken(ctx context.Context, userID int64, token string, expire time.Duration) error {
	key := fmt.Sprintf("token:%d", userID)
	return r.client.Set(ctx, key, token, expire).Err()
}

// GetUserToken 获取用户 Token
func (r *Redis) GetUserToken(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("token:%d", userID)
	return r.client.Get(ctx, key).Result()
}

// DeleteUserToken 删除用户 Token
func (r *Redis) DeleteUserToken(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("token:%d", userID)
	return r.client.Del(ctx, key).Err()
}

// ----- 限流 -----

// RateLimit 限流（滑动窗口）
func (r *Redis) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := r.client.Pipeline()

	// 删除窗口外的记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// 获取窗口内请求数
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 设置过期
	r.client.Expire(ctx, key, window)

	return count <= int64(limit), nil
}

// ----- 熵减排行 -----

// IncrementEntropy 增加熵减
func (r *Redis) IncrementEntropy(ctx context.Context, userID int64, delta float64) error {
	key := "entropy:rank"
	return r.client.ZIncrBy(ctx, key, delta, fmt.Sprintf("%d", userID)).Err()
}

// GetEntropyRank 获取熵减排行
func (r *Redis) GetEntropyRank(ctx context.Context, userID int64) (int64, error) {
	key := "entropy:rank"
	rank, err := r.client.ZRevRank(ctx, key, fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return rank + 1, nil // 1-based
}

// ----- 在线状态 -----

// SetOnline 设置在线
func (r *Redis) SetOnline(ctx context.Context, userID int64) error {
	key := "online"
	return r.client.Set(ctx, fmt.Sprintf("%s:%d", key, userID), time.Now().Unix(), 5*time.Minute).Err()
}
