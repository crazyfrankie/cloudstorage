//go:build wireinject

package ioc

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/crazyfrankie/cloudstorage/app/gateway/api"
	"github.com/crazyfrankie/cloudstorage/app/gateway/mws"
)

func InitRegistry() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	return cli
}

func InitMws() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		mws.NewAuthBuilder().
			IgnorePath("/api/user/send-code").
			IgnorePath("/api/user/verify-code").
			Auth(),
	}
}

func InitGin(mws []gin.HandlerFunc, user *api.UserHandler, file *api.FileHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mws...)

	user.RegisterRoute(server)
	file.RegisterRoute(server)

	return server
}

func InitServer() *gin.Engine {
	wire.Build(
		InitRegistry,
		InitUserClient,
		InitFileClient,
		api.NewUserHandler,
		api.NewFileHandler,
		InitMws,
		InitGin,
	)
	return new(gin.Engine)
}
