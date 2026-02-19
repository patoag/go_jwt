package repositories

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenBlacklist implementa TokenBlacklist usando Redis
type RedisTokenBlacklist struct {
	client *redis.Client
}

// NewRedisTokenBlacklist crea una nueva instancia de RedisTokenBlacklist
func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) Add(token string, remainingTTL int64) error {
	if b.client == nil {
		return nil
	}

	key := blacklistKey(token)
	ttl := time.Duration(remainingTTL) * time.Second
	if ttl <= 0 {
		ttl = time.Hour // fallback mínimo
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return b.client.Set(ctx, key, "1", ttl).Err()
}

func (b *RedisTokenBlacklist) IsBlacklisted(token string) (bool, error) {
	if b.client == nil {
		return false, nil
	}

	key := blacklistKey(token)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func blacklistKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("blacklist:%x", hash)
}
