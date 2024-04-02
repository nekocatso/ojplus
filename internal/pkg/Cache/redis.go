package Cache

import (
	"fmt"
	"time"
	"github.com/gomodule/redigo/redis"
)

type RedisClientPool struct {
	pool *redis.Pool
}

func NewRedisClientPool() *RedisClientPool {
	var rcp RedisClientPool
	rcp.pool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   12000,
		IdleTimeout: time.Duration(30),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379", redis.DialPassword(""), redis.DialReadTimeout(time.Second), redis.DialWriteTimeout(time.Second))
			if err != nil {
				fmt.Printf("redisClient dial host: %s, auth: %s err: %s\n", "127.0.0.1:6379", " ", err.Error())
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				fmt.Println("redisClient ping err: ", err.Error())
			}
			return err
		},
	}
	fmt.Println("init RedisClientPool ok")
	return &rcp
}

func (rcp *RedisClientPool) Close() {
	if rcp.pool != nil {
		err := rcp.pool.Close()
		if err != nil {
			fmt.Println("do Close RedisClientPool error:", err.Error())
		}
	}
}
func (rcp *RedisClientPool) GetConn() redis.Conn {
	return rcp.pool.Get()
}
