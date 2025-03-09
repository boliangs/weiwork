package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"sendswork/config"
)

var RedisClient *redis.Client

func InitRDB() {
	rConfig := config.Conf.Redis["user"]
	client := redis.NewClient(&redis.Options{
		Addr:     rConfig.Address,  // Redis 地址
		Password: rConfig.Password, // Redis 密码
		DB:       0,                // Redis 数据库编号
	})
	_, err := client.Ping(context.Background()).Result() // 测试连接
	if err != nil {
		logrus.Info(err)
		panic(err) // 如果连接失败，抛出异常
	}
	RedisClient = client
}

func NewRDBClient(ctx context.Context) *redis.Client {
	return RedisClient.WithContext(ctx) // 返回带上下文的 Redis 客户端
}
