// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/biz/service/sms/memory"
	"github.com/crazyfrankie/cloudstorage/app/sm/internal/config"
	"github.com/redis/go-redis/v9"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"time"
)

// Injectors from wire.go:

func InitSmServer() *service.SmServer {
	cmdable := InitCache()
	smCache := cache.NewSmCache(cmdable)
	smRepo := repository.NewSmRepo(smCache)
	smsService := memory.NewMemorySmService()
	smServer := service.NewSmServer(smRepo, smsService)
	return smServer
}

// wire.go:

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
