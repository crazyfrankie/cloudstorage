package repository

import (
	"context"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/repository/cache"
)

type SmRepo struct {
	cache *cache.SmCache
}

func NewSmRepo(c *cache.SmCache) *SmRepo {
	return &SmRepo{cache: c}
}

func (r *SmRepo) Store(ctx context.Context, biz, phone, code string) error {
	return r.cache.Store(ctx, biz, phone, code)
}

func (r *SmRepo) Verify(ctx context.Context, biz, phone, code string) error {
	return r.cache.Verify(ctx, biz, phone, code)
}
