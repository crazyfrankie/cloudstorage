//go:build wireinject

package ioc

import (
	"fmt"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/biz/infra/rpc"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/config"
	"github.com/crazyfrankie/cloudstorage/app/user/internal/mws"
	"os"
	"time"

	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(config.GetConf().MySQL.DSN,
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&dao.User{})

	return db
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

func InitMinio() *minio.Client {
	client, err := minio.New(config.GetConf().Minio.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetConf().Minio.AccessKey, config.GetConf().Minio.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}

	return client
}

func InitClientLog() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func InitServerLog() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return logger
}

func InitUserServer() *service.UserServer {
	wire.Build(
		InitDB,
		InitRegistry,
		InitMinio,
		InitClientLog,
		mws.NewMinioServer,
		dao.NewUserDao,
		repository.NewUserRepo,
		service.NewUserServer,
		rpc.InitSmClient,
		rpc.InitFileClient,
	)

	return new(service.UserServer)
}
