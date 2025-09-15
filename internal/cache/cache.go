package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var TTL = 5 * time.Minute

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func GetPost(ctx context.Context, rdb *redis.Client, id string) (string, error) {
	return rdb.Get(ctx, "post:"+id).Result()
}

func SetPost(ctx context.Context, rdb *redis.Client, id string, data string) error {
	return rdb.Set(ctx, "post:"+id, data, TTL).Err()
}

func InvalidatePost(ctx context.Context, rdb *redis.Client, id string) error {
	return rdb.Del(ctx, "post:"+id).Err()
}
