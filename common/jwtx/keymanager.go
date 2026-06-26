package jwtx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	JWT_SECRET_KEY_PREFIX = "jwt:secret:key"
	DEFAULT_SECRET_LENGTH = 32
)

type KeyManager struct {
	redisClient *redis.Redis
	mu          sync.RWMutex
	cachedKey   string
}

func NewKeyManager(redisClient *redis.Redis) *KeyManager {
	return &KeyManager{
		redisClient: redisClient,
	}
}

// GetOrCreateSecretKey 从Redis获取或创建JWT密钥
func (km *KeyManager) GetOrCreateSecretKey(ctx context.Context) (string, error) {
	km.mu.RLock()
	if km.cachedKey != "" {
		km.mu.RUnlock()
		return km.cachedKey, nil
	}
	km.mu.RUnlock()

	km.mu.Lock()
	defer km.mu.Unlock()

	// 双重检查
	if km.cachedKey != "" {
		return km.cachedKey, nil
	}

	// 尝试从Redis获取密钥
	key, err := km.redisClient.Get(JWT_SECRET_KEY_PREFIX)
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("failed to get secret key from redis: %v", err)
	}

	// 如果Redis中没有密钥，则生成新的密钥
	if err == redis.Nil || key == "" {
		key, err = km.generateSecretKey()
		if err != nil {
			return "", fmt.Errorf("failed to generate secret key: %v", err)
		}

		// 将新密钥存储到Redis（永不过期）
		err = km.redisClient.Set(JWT_SECRET_KEY_PREFIX, key)
		if err != nil {
			return "", fmt.Errorf("failed to store secret key to redis: %v", err)
		}
	}

	// 缓存密钥到内存
	km.cachedKey = key
	return key, nil
}

// generateSecretKey 生成随机密钥
func (km *KeyManager) generateSecretKey() (string, error) {
	bytes := make([]byte, DEFAULT_SECRET_LENGTH)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RefreshSecretKey 强制刷新密钥（慎用，会导致所有现有token失效）
func (km *KeyManager) RefreshSecretKey(ctx context.Context) (string, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	// 生成新密钥
	key, err := km.generateSecretKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate secret key: %v", err)
	}

	// 存储到Redis
	err = km.redisClient.Set(JWT_SECRET_KEY_PREFIX, key)
	if err != nil {
		return "", fmt.Errorf("failed to store secret key to redis: %v", err)
	}

	// 更新缓存
	km.cachedKey = key
	return key, nil
}

// GetSecretKey 获取当前密钥（仅从缓存获取）
func (km *KeyManager) GetSecretKey() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.cachedKey
}