package user

import (
	"context"
	"errors"
	"hahacking/common/redis"
	"log"
	"strings"
	"time"

	redisClient "github.com/go-redis/redis/v8"
)

// 延迟初始化，避免在测试时自动创建客户端
var tokenStore *TokenStore

func GetTokenStore() *TokenStore {
	if tokenStore == nil {
		tokenStore = NewTokenStore()
	}
	return tokenStore
}

type TokenStore struct {
	client *redisClient.Client
	ctx    context.Context
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		client: redis.NewClient(),
		ctx:    context.Background(),
	}
}

type UserFeature struct {
	RefreshToken string
	UserAgent    string
	ExpiresAt    time.Time
}

func (t *TokenStore) Revoke(userID string) error {
	// 获取所有refresh_token开头的键
	pattern := "refresh_token:*"
	keys, err := t.client.Keys(t.ctx, pattern).Result()
	if err != nil {
		return errors.New("无法获取令牌列表: " + err.Error())
	}

	// 依次查找全部元素检查，如果id是userID，则删除
	for _, key := range keys {
		// 获取令牌信息
		result, err := t.client.HGetAll(t.ctx, key).Result()
		if err != nil {
			continue // 跳过有问题的键
		}

		// 检查是否包含用户ID信息（这里假设用户ID存储在某个字段中）
		// 如果令牌的某个字段包含userID，则删除该令牌
		if containsUserID(result, userID) {
			err := t.client.Del(t.ctx, key).Err()
			if err != nil {
				// 记录错误但继续处理其他令牌
				continue
			}
		}
	}

	return nil
}

// containsUserID 检查结果中是否包含指定的用户ID
func containsUserID(result map[string]string, userID string) bool {
	for _, value := range result {
		if strings.Contains(value, userID) {
			return true
		}
	}
	return false
}

func (t *TokenStore) Add(refreshToken string, feature UserFeature) error {
	key := "refresh_token:" + refreshToken
	pipe := t.client.Pipeline()
	ctx := t.ctx
	pipe.HSet(ctx, key, "userAgent", feature.UserAgent, "expiresAt", feature.ExpiresAt)
	pipe.PExpire(ctx, key, time.Duration(feature.ExpiresAt.Unix()-time.Now().Unix())*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.New("无法保存令牌到redis")
	}
	log.Printf("尝试添加Redis键: %s", key)
	return nil
}
func (t *TokenStore) Get(refreshToken string) (UserFeature, error) {
	key := "refresh_token:" + refreshToken

	log.Printf("尝试获取Redis键: %s", key)

	result, err := t.client.HGetAll(t.ctx, key).Result()
	if err != nil {
		log.Printf("HGetAll失败: %v", err)
		return UserFeature{}, errors.New("无法从 Redis 加载令牌: " + err.Error())
	}

	log.Printf("Redis返回结果: %+v", result)
	log.Printf("结果长度: %d", len(result))

	if len(result) == 0 {
		log.Printf("结果为空")
		return UserFeature{}, errors.New("令牌不存在或已过期")
	}

	// 获取expiresAt字段
	expiresAtStr := result["expiresAt"]
	if expiresAtStr == "" {
		log.Printf("expiresAt字段为空")
		return UserFeature{}, errors.New("令牌数据不完整：缺少expiresAt字段")
	}

	log.Printf("expiresAt字符串: %s", expiresAtStr)

	// 尝试解析时间
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		log.Printf("RFC3339时间解析失败: %v", err)
		// 尝试其他格式
		expiresAt, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", expiresAtStr)
		if err != nil {
			log.Printf("标准格式时间解析也失败: %v", err)
			return UserFeature{}, errors.New("无法解析过期时间: " + err.Error())
		}
	}

	feature := UserFeature{
		UserAgent: result["userAgent"],
		ExpiresAt: expiresAt,
	}

	log.Printf("成功构建UserFeature: %+v", feature)
	return feature, nil
}
func (t *TokenStore) Delete(refreshToken string) error {
	key := "refresh_token:" + refreshToken
	err := t.client.Del(t.ctx, key).Err()
	if err != nil {
		return errors.New("无法删除令牌: " + err.Error())
	}
	return nil
}
