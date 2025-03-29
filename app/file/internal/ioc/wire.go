//go:build wireinject

package ioc

import (
	"fmt"
	"os"
	"time"

	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/config"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/mws"
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

	db.AutoMigrate(&dao.File{}, &dao.FileStore{}, &dao.Folder{})

	return db
}

func InitCache() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.GetConf().Redis.Addr,
	})
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

func InitServer() *service.FileServer {
	wire.Build(
		InitDB,
		InitMinio,
		InitCache,
		dao.NewUploadDao,
		cache.NewFileCache,
		repository.NewUploadRepo,
		mws.NewMinioServer,
		mws.NewKafkaProducer,
		service.NewRedisWorker,
		service.NewFileServer,
	)
	return new(service.FileServer)
}
