//go:build wireinject

package ioc

import (
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/service/sms/memory"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/config"
)

func InitCache() redis.Cmdable {
	cli := redis.NewClient(&redis.Options{
		Addr: config.GetConf().Redis.Addr,
	})

	return cli
}

func InitRegistry() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.GetConf().ETCD.Addr},
		DialTimeout: time.Second * 2,
	})
	if err != nil {
		panic(err)
	}

	return cli
}

func InitServerLog() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return logger
}

func InitSmServer() *service.SmServer {
	wire.Build(
		InitCache,
		cache.NewSmCache,
		repository.NewSmRepo,
		memory.NewMemorySmService,
		service.NewSmServer,
	)
	return new(service.SmServer)
}
