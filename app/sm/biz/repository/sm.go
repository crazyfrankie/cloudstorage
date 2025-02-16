package repository

import "github.com/crazyfrankie/cloudstorage/app/sm/biz/repository/cache"

type SmRepo struct {
	cache *cache.SmCache
}

func NewSmRepo(c *cache.SmCache) *SmRepo {
	return &SmRepo{cache: c}
}
