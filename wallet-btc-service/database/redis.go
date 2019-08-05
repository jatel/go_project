package database

import (
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/go-redis/redis"
)

var RedisDb *redis.Client

func InitRedis(address string, dbNum int) error {
	redisdb := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     "",
		DB:           dbNum,
		PoolSize:     1000,
		PoolTimeout:  2 * time.Minute,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	})

	pong, err := redisdb.Ping().Result()
	if nil != err {
		return err
	}

	log.Log.Notice("redis connect result:", pong)

	RedisDb = redisdb
	return nil
}
