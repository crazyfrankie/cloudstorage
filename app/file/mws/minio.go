package mws

import "github.com/minio/minio-go/v7"

type MinioServer struct {
	client *minio.Client
}

func NewMinioServer(client *minio.Client) *MinioServer {
	return &MinioServer{client: client}
}
