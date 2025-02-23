package mws

import (
	"bytes"
	"context"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloudstorage/app/user/config"
)

type MinioServer struct {
	client     *minio.Client
	BucketName string
}

func NewMinioServer(client *minio.Client) *MinioServer {
	name := config.GetConf().Minio.BucketName
	server := &MinioServer{client: client, BucketName: name}

	server.MakeBucket(context.Background(), name)

	return server
}

// MakeBucket 创建 Bucket
func (m *MinioServer) MakeBucket(ctx context.Context, name string) {
	exists, err := m.client.BucketExists(ctx, name)
	if err != nil {
		return
	}
	if !exists {
		err = m.client.MakeBucket(ctx, name, minio.MakeBucketOptions{})
		if err != nil {
			return
		}
	} else {
		log.Println("bucket exists")
	}
}

// PutToBucket 放入对象
func (m *MinioServer) PutToBucket(ctx context.Context, bucketName, filename string, filesize int64, data []byte) (minio.UploadInfo, error) {
	info, err := m.client.PutObject(ctx, bucketName, filename, bytes.NewReader(data), filesize, minio.PutObjectOptions{})
	return info, err
}

func (m *MinioServer) PresignedGetObject(ctx context.Context, bucketName, filename string, expiration time.Duration) (*url.URL, error) {
	reqParams := make(url.Values)
	return m.client.PresignedGetObject(ctx, bucketName, filename, expiration, reqParams)
}
