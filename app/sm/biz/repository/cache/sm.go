package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type SmCache struct {
	cmd redis.Cmdable
}

func NewSmCache(cmd redis.Cmdable) *SmCache {
	return &SmCache{cmd: cmd}
}

func (s *SmCache) Store(ctx context.Context) error {
	return nil
}
