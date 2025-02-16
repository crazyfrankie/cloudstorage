//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/service/sms/memory"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/crazyfrankie/cloudstorage/app/sm/config"
	"github.com/crazyfrankie/cloudstorage/app/sm/rpc"
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

func InitServer() *rpc.Server {
	wire.Build(
		InitCache,
		InitRegistry,
		cache.NewSmCache,
		repository.NewSmRepo,
		memory.NewMemorySmService,
		service.NewSmServer,
		rpc.NewServer,
	)
	return new(rpc.Server)
}
