package mws

import (
	"bytes"
	"context"
	"log"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloudstorage/app/file/config"
)

type MinioServer struct {
	client     *minio.Client
	bucketName string
}

func NewMinioServer(client *minio.Client) *MinioServer {
	name := config.GetConf().Minio.BucketName[0]
	server := &MinioServer{client: client, bucketName: name}

	server.MakeBucket(context.Background(), name)

	return server
}

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

func (m *MinioServer) PutToBucket(ctx context.Context, filename string, filesize int64, data []byte) (minio.UploadInfo, error) {
	info, err := m.client.PutObject(ctx, m.bucketName, filename, bytes.NewReader(data), filesize, minio.PutObjectOptions{})
	return info, err
}
