package kvdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient implements Client
type RedisClient struct {
	Conf *Conf

	// internal fields are implementation details, not exported
	client *redis.Client
}

func (c *RedisClient) Init() error {
	c.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Conf.Host, c.Conf.Port),
		Password: c.Conf.PW,
		DB:       c.Conf.DB,
	})
	log.Println("[INFO] redis client initialized")
	return nil
}

func (c *RedisClient) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *RedisClient) DBHandle() any { // use with runtime type assertion
	return c.client
}

//--- Group Ops ----

func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *RedisClient) Delete(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// Redis EXPIRE returns true if key existed and TTL was set, false if key does not exist
	return c.client.Expire(ctx, key, expiration).Result()
}

//---- Single-value Ops ----

func (c *RedisClient) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil // redis.Nil -> ok: false, err: nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

//---- List Ops ----

func (c *RedisClient) Push(ctx context.Context, key, value string) error {
	// Add to the tail (right) of the list
	return c.client.RPush(ctx, key, value).Err()
}

func (c *RedisClient) Pop(ctx context.Context, key string) (string, bool, error) { // val, found, err
	// Pop from the head (left) of the list (FIFO)
	val, err := c.client.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil // redis.Nil -> ok: false, err: nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *RedisClient) Len(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

func (c *RedisClient) Range(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

func (c *RedisClient) Remove(ctx context.Context, key string, cnt int64, value any) (int64, error) {
	return c.client.LRem(ctx, key, cnt, value).Result()
}

func (c *RedisClient) Trim(ctx context.Context, key string, start, stop int64) error {
	return c.client.LTrim(ctx, key, start, stop).Err()
}

//---- Hash Ops ----

func (c *RedisClient) SetField(ctx context.Context, key string, field string, value any) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

func (c *RedisClient) GetField(ctx context.Context, key string, field string) (string, bool, error) { // val, found, err
	val, err := c.client.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil // key or field missing
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *RedisClient) SetFields(ctx context.Context, key string, fields map[string]any) error {
	return c.client.HSet(ctx, key, fields).Err()
}

// GetFields returns a map {field:value} from a hash data, which contains only found fields
// so, if len(rtnMap) < len(fields), some fields are missing
// [NOTE] returns an empty map even if key is not found. not error
func (c *RedisClient) GetFields(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	resultSlice, err := c.client.HMGet(ctx, key, fields...).Result() // []any
	if err != nil {
		return nil, err
	}
	rtnMap := make(map[string]string, len(fields)) // capacity = max len = when all fields found
	for i, v := range resultSlice {
		if v != nil {
			rtnMap[fields[i]] = fmt.Sprint(v)
		}
		// if v is nil, field missing → omitted in the return map
	}
	return rtnMap, nil
}

func (c *RedisClient) RemoveFields(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.client.HDel(ctx, key, fields...).Result()
}

// GetAllFields returns a map {field:value} from a hash data with all fields in it
// [NOTE] returns an empty map even if key is not found. not error
func (c *RedisClient) GetAllFields(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}
