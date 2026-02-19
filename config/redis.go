package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// ConnectRedis establece la conexión con Redis
func ConnectRedis(cfg *RedisConfig) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Advertencia: No se pudo conectar a Redis: %v (funcionalidades dependientes de Redis estarán deshabilitadas)", err)
		RedisClient = nil
		return
	}

	log.Println("Conexión a Redis establecida exitosamente")
}
