//go:build wireinject

package ioc

import (
	"github.com/gin-contrib/cors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/crazyfrankie/cloudstorage/app/gateway/api"
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
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:8081"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length", "x-jwt-token"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
	}
}

func InitGin(mws []gin.HandlerFunc, user *api.UserHandler, file *api.FileHandler) *gin.Engine {
	server := gin.Default()
	server.MaxMultipartMemory = 100 * 1024 * 1024
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
