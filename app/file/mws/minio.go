package mws

import (
	"bytes"
	"context"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloudstorage/app/file/config"
)

type MinioServer struct {
	client     *minio.Client
	core       *minio.Core
	BucketName string
}

func NewMinioServer(client *minio.Client) *MinioServer {
	name := config.GetConf().Minio.BucketName[0]
	// 从 client 获取 core
	core := &minio.Core{Client: client}
	server := &MinioServer{client: client, BucketName: name, core: core}

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

// GetObject 获取对象信息
func (m *MinioServer) GetObject(ctx context.Context, bucketName, filename string) (*minio.Object, error) {
	info, err := m.client.GetObject(ctx, bucketName, filename, minio.GetObjectOptions{})
	return info, err
}

// PresignedGetObject 获取预览 URL
func (m *MinioServer) PresignedGetObject(ctx context.Context, bucketName, filename string, expiration time.Duration) (*url.URL, error) {
	reqParams := make(url.Values)
	return m.client.PresignedGetObject(ctx, bucketName, filename, expiration, reqParams)
}

// CreateMultipartUpload 初始化分片上传
func (m *MinioServer) CreateMultipartUpload(ctx context.Context, bucketName, filename string) (string, error) {
	uploadID, err := m.core.NewMultipartUpload(ctx, bucketName, filename, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return uploadID, nil
}

// PutObjectPart 上传分片
func (m *MinioServer) PutObjectPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int,
	data []byte, size int64) (minio.ObjectPart, error) {
	return m.core.PutObjectPart(ctx, bucketName, objectName, uploadID, partNumber, bytes.NewReader(data), size, minio.PutObjectPartOptions{})
}

// CompleteMultipartUpload 完成分片上传
func (m *MinioServer) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string,
	parts []minio.CompletePart) (minio.UploadInfo, error) {
	return m.core.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, parts, minio.PutObjectOptions{})
}
