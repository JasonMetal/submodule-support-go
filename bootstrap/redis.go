package bootstrap

import (
	"fmt"
	"gitee.com/DXTeam/idea-go.git/helper/config"
	redisHelper "gitee.com/DXTeam/idea-go.git/helper/redis"
	"github.com/gomodule/redigo/redis"
	yCfg "github.com/olebedev/config"
	"os"
	"time"
)

func InitRedis() {
	dbList := getDbNames("redis")
	for _, dbname := range dbList {
		instances, err := initRedisPool(dbname)
		if err == nil {
			redisHelper.SetRedisInstance(dbname, instances)
		}
	}

	//go closePool()
}

func initRedisPool(dbName string) ([]redisHelper.RedisInstance, error) {
	path := fmt.Sprintf("%sconfig/%s/redis.yml", ProjectPath(), DevEnv)

	cfg, err := config.GetConfig(path)
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
	redisPools := make([]redisHelper.RedisInstance, len(servers))
	for k, v := range servers {
		address, _ := yCfg.Get(v, "address")
		passwd, _ := yCfg.Get(v, "passwd")

		if address == nil {
			os.Exit(1)
		}

		// 建立连接池
		redisPools[k] = redisHelper.RedisInstance{
			DSN: address.(string),
			Pool: &redis.Pool{
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
