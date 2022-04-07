package bootstrap

import (
	"github.com/gomodule/redigo/redis"
	yCfg "github.com/olebedev/config"
	"idea-go/helpers/config"
	"os"
	"time"
)

type RedisInstance struct {
	DSN  string
	Pool *redis.Pool
}

var redisPoolList = make(map[string][]RedisInstance)

func InitRedis() {
	dbList := getDbNames("redis")
	for _, dbname := range dbList {
		instance, err := initRedisPool(dbname)
		if err == nil {
			redisPoolList[dbname] = instance
		}
	}

	//go closePool()
}

func initRedisPool(dbName string) ([]RedisInstance, error) {

	cfg, err := config.GetConfig("redis")
	if err != nil {
		return nil, err
	}

	maxIdle, _ := cfg.Int("redis." + dbName + ".maxIdle")
	maxActive, _ := cfg.Int("redis." + dbName + ".maxActive")
	idleTimeout, _ := cfg.Int("redis." + dbName + ".idleTimeout")

	servers, err := cfg.List("redis." + dbName + ".servers")
	if err != nil {
		return nil, err
	}
	redisPools := make([]RedisInstance, len(servers))
	for k, v := range servers {
		address, _ := yCfg.Get(v, "address")
		passwd, _ := yCfg.Get(v, "passwd")

		if address == nil {
			os.Exit(1)
		}

		// 建立连接池
		redisPools[k] = RedisInstance{
			address.(string),
			&redis.Pool{
				MaxIdle:     maxIdle,
				MaxActive:   maxActive,
				IdleTimeout: time.Duration(idleTimeout) * time.Second,
				Dial: func() (redis.Conn, error) {
					c, err := redis.Dial("tcp", address.(string),
						redis.DialReadTimeout(time.Duration(100)*time.Millisecond),
						redis.DialWriteTimeout(time.Duration(100)*time.Millisecond),
						redis.DialConnectTimeout(time.Duration(1000)*time.Millisecond),
					)
					if err != nil {

						return nil, err
					}

					if passwd != nil {
						if _, err := c.Do("AUTH", passwd); err != nil {

							c.Close()

							return nil, err
						}
					}

					return c, nil
				},
			}}
	}

	return redisPools, nil
}
