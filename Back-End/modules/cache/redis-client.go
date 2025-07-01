package cache

import (
	"TaleHaven-Backend/modules/utils"
	"context"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
)

var (
	// Client Global Redis client
	Client *redis.Client
	// Ctx Context that can be used throughout the project
	Ctx = context.Background()
)

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PWD"),
		DB:       0,
	})

	pong, err := Client.Ping(Ctx).Result()
	if err != nil {
		utils.Logger.Fatal("Could not connect to Redis", zap.Error(err))
	}

	utils.Logger.Info("Connected to Redis", zap.String("pong", pong))
}
