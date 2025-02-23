//go:build wireinject

package ioc

import (
	"fmt"
	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/user/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/user/mws"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/user/rpc"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/crazyfrankie/cloudstorage/app/user/config"
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

func InitServer() *rpc.Server {
	wire.Build(
		InitDB,
		InitRegistry,
		InitMinio,
		mws.NewMinioServer,
		dao.NewUserDao,
		repository.NewUserRepo,
		service.NewUserServer,
		rpc.InitSmClient,
		rpc.InitFileClient,
		rpc.NewServer,
	)
	return new(rpc.Server)
}
