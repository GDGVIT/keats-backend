package redisclient

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var rdb *redis.Client = nil

func GetRedisClient() (*redis.Client, error) {
	if rdb != nil {
		return rdb, nil
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_ADDRESS") + ":" + viper.GetString("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil

}
