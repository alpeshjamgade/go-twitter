package data

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

type Cache struct {
	Addr    string
	Client  *redis.Client
	Context context.Context
}

func NewCache(addr string) *Cache {
	return &Cache{Addr: addr}
}
func (c *Cache) Connect() error {
	c.Client = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: "password", // Optional, remove this line if Redis doesn't require authentication
		DB:       0,          // Use the default Redis database (0 by default)
	})
	c.Context = c.Client.Context()
	_, err := c.Client.Ping(c.Context).Result()
	if err != nil {
		log.Printf("Unable to connect to redis %v", err)
		return err
	}
	return nil
}

func (c *Cache) Get(key string) string {
	return c.Client.Get(c.Context, key).Val()
}

func (c *Cache) HGet(key string, field string) string {
	return c.Client.HGet(c.Context, key, field).Val()
}

func (c *Cache) Set(key string, value string, duration time.Duration) (string, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Printf("error while updating value in redis %v", err)
		return "", err
	}

	cmd := c.Client.Set(c.Context, key, jsonData, duration)
	return cmd.Result()
}

func (c *Cache) HSetNX(key string, field string, value string, duration time.Duration) (bool, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		log.Printf("error while updating value in redis %v", err)
		return false, err
	}

	cmd := c.Client.HSetNX(c.Context, key, field, jsonData)
	return cmd.Result()
}

func (c *Cache) Del(key string) {
	c.Client.Del(c.Context, key)
}

func (c *Cache) HDel(key string, field string) {
	c.Client.HDel(c.Context, key, field)
}
