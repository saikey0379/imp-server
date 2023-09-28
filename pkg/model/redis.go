package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/saikey0379/imp-server/pkg/config"
	"github.com/saikey0379/imp-server/pkg/logger"
)

var ctx = context.Background()
var mutex sync.Mutex

type Redis interface {
	Log() logger.Logger
	Keys(keyReg string) ([]string, error)
	Get(key string) (string, error)
	Set(key string, value string) (string, error)
	SetEx(key string, value string, second int) (string, error)
	Del(key string) (int64, error)
	Exists(key string) (int64, error)
	Lock(key string) bool
	UnLock(key string) int64
}

type RedisClient struct {
	conf   *config.Config
	log    logger.Logger
	client *redis.Client
}

func NewRedis(conf *config.Config, log logger.Logger) (*RedisClient, error) {
	redis_opt := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Server.RedisAddr, conf.Server.RedisPort),
		Password: conf.Server.RedisPasswd,
		DB:       conf.Server.RedisDBNumber,
	}
	// 创建连接池
	c := redis.NewClient(&redis_opt)

	// 判断是否能够链接到数据库
	_, err := c.Ping(ctx).Result()

	var redisClient = &RedisClient{
		conf:   conf,
		log:    log,
		client: c,
	}
	return redisClient, err
}

func (redisClient *RedisClient) Log() logger.Logger {
	return redisClient.log
}

func (redisClient *RedisClient) Keys(keyReg string) ([]string, error) {
	return redisClient.client.Keys(ctx, keyReg).Result()
}

func (redisClient *RedisClient) Get(key string) (string, error) {
	return redisClient.client.Get(ctx, key).Result()
}

func (redisClient *RedisClient) Set(key string, value string) (string, error) {
	return redisClient.client.Set(ctx, key, value, 0).Result()
}

func (redisClient *RedisClient) SetEx(key string, value string, second int) (string, error) {
	return redisClient.client.SetEx(ctx, key, value, time.Duration(second)).Result()
}

func (redisClient *RedisClient) Del(key string) (int64, error) {
	return redisClient.client.Del(ctx, key).Result()
}

func (redisClient *RedisClient) Exists(key string) (int64, error) {
	return redisClient.client.Exists(ctx, key).Result()
}

func (redisClient *RedisClient) Lock(key string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	bool, err := redisClient.client.SetNX(ctx, key, 1, 10*time.Second).Result()
	if err != nil {
		redisClient.log.Error("FAILURE: Lock[%v]", err)
	}
	return bool
}

func (redisClient *RedisClient) UnLock(key string) int64 {
	nums, err := redisClient.client.Del(ctx, key).Result()
	if err != nil {
		redisClient.log.Error("FAILURE: UnLock[%v]", err)
	}
	return nums
}
