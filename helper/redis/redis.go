package redis

import (
	"context"
	"encoding/json"
	"errors"
	"gitee.com/DXTeam/idea-go.git/helper/logger"
	"gitee.com/DXTeam/idea-go.git/helper/number"
	"github.com/gomodule/redigo/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type RedisInstance struct {
	DSN  string
	Pool *redis.Pool
}

var redisPoolList = make(map[string][]RedisInstance)

func GetRedisInstance(dbName string) (*RedisInstance, error) {
	if list, ok := redisPoolList[dbName]; ok {
		l := len(redisPoolList[dbName])
		if l > 0 {
			l = l - 1
			i := number.GetRandNum(l)

			return &list[i], nil
		}
	}

	return nil, errors.New(dbName + " redis list is nil")
}

func SetRedisInstance(dbName string, instance []RedisInstance) {
	redisPoolList[dbName] = instance
}
func closePool() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		for _, pools := range redisPoolList {
			for _, instance := range pools {
				instance.Pool.Close()
			}
		}
		redisPoolList = nil

		//os.Exit(0)
	}()
}

func (r *RedisInstance) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	//redis采集先关闭
	if opentracing.IsGlobalTracerRegistered() && ctx != nil {
		span, _ := opentracing.StartSpanFromContext(ctx, "redis::"+commandName)
		defer span.Finish()
		ext.Component.Set(span, args[0].(string))
		//ext.DBInstance.Set(span, "common")
	}

	conn, err := r.getConn(0)
	if conn == nil || err != nil {
		return nil, err
	}

	defer conn.Close()
	reply, e := conn.Do(commandName, args...)

	return reply, e
}

// Set 用法：Set("key", val, 60)，其中 expire 的单位为秒
func (r *RedisInstance) Set(ctx context.Context, key string, val interface{}, expire int) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	case []byte:
		value = string(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}
	if expire > 0 {
		return r.Do(ctx, "SETEX", key, expire, value)
	} else {
		return r.Do(ctx, "SET", key, value)
	}
}

func (r *RedisInstance) Get(ctx context.Context, key string) (interface{}, error) {
	return r.Do(ctx, "GET", key)
}

// SetNxEx exist set value + expires otherwise not do
// cmd: set key value ex 3600 nx
func (r *RedisInstance) SetNxEx(ctx context.Context, key string, val interface{}, expire int) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}
	if expire > 0 {
		return r.Do(ctx, "SET", key, value, "ex", expire, "nx")
	} else {
		return r.Do(ctx, "SET", key, value, "nx")
	}
}

func (r *RedisInstance) Del(ctx context.Context, key string) error {
	_, err := r.Do(ctx, "DEL", key)
	return err
}

func (r *RedisInstance) Incr(ctx context.Context, key string) (val int64, err error) {
	return redis.Int64(r.Do(ctx, "INCR", key))
}

func (r *RedisInstance) IncrBy(ctx context.Context, key string, incrAmount int) (val int64, err error) {
	return redis.Int64(r.Do(ctx, "INCRBY", key, incrAmount))
}

func (r *RedisInstance) Decr(ctx context.Context, key string) (val int64, err error) {
	return redis.Int64(r.Do(ctx, "DECR", key))
}

func (r *RedisInstance) DecrBy(ctx context.Context, key string, decrAmount int) (val int64, err error) {
	return redis.Int64(r.Do(ctx, "DECRBY", key, decrAmount))
}

// Lpop 用法：Lpop("key")
func (r *RedisInstance) Lpop(ctx context.Context, key string) (interface{}, error) {
	return r.Do(ctx, "LPOP", key)
}

// Lpush 用法：Lpush("key", val)
func (r *RedisInstance) Lpush(ctx context.Context, key string, val interface{}) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}

	return r.Do(ctx, "LPUSH", key, value)
}

// Rpop 用法：Rpop("key")
func (r *RedisInstance) Rpop(ctx context.Context, key string) (interface{}, error) {
	return r.Do(ctx, "RPOP", key)
}

func (r *RedisInstance) Lrem(ctx context.Context, key string, val interface{}, count int64) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}

	return r.Do(ctx, "LREM", key, count, value)
}

// Expire 设置键过期时间，expire的单位为秒
func (r *RedisInstance) Expire(ctx context.Context, key string, expire int) error {
	_, err := redis.Bool(r.Do(ctx, "EXPIRE", key, expire))
	return err
}

func (r *RedisInstance) getConn(retryTimes int) (redis.Conn, error) {
	if retryTimes >= 3 {
		err := errors.New("empty conns")
		logger.Error("redis", zap.NamedError("empty_conn", err))
		return nil, err
	}
	conn := r.Pool.Get()
	err := conn.Err()

	if err != nil {
		retryTimes++
		logger.Error("redis", zap.NamedError("conn_err", err))
		conn.Close()

		return r.getConn(retryTimes)
	} else {

		return conn, nil
	}
}

func (r *RedisInstance) GetString(ctx context.Context, key string) (string, error) {
	return redis.String(r.Do(ctx, "GET", key))
}

func (r *RedisInstance) SetString(ctx context.Context, key string, value string) (string, error) {
	return redis.String(r.Do(ctx, "SET", key, value))
}

func (r *RedisInstance) Ttl(ctx context.Context, key string) (int64, error) {
	return redis.Int64(r.Do(ctx, "TTL", key))
}
