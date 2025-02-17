package service

import (
	"context"
	"os"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

type FileServer struct {
	repo  *repository.UploadRepo
	minio *mws.MinioServer
	file.UnimplementedFileServiceServer
}

func NewFileServer(repo *repository.UploadRepo, minio *mws.MinioServer) *FileServer {
	return &FileServer{repo: repo, minio: minio}
}

func (s *FileServer) Upload(ctx context.Context, req *file.UploadRequest) (*file.UploadResponse, error) {
	meta, data := req.GetMetadata(), req.GetData()

	// 存到本地
	newFile, err := os.Create(meta.Path)
	if err != nil {
		return nil, err
	}
	_, err = newFile.Write(data)
	newFile.Close()
	if err != nil {
		return nil, err
	}

	// 存 OSS
	info, err := s.minio.PutToBucket(ctx, meta.Name, meta.Size, data)
	if err != nil {
		return nil, err
	}

	// 存数据库
	f := &dao.File{
		Name: meta.Name,
		Hash: meta.Hash,
		Type: meta.ContentType,
		Path: meta.Path,
		Size: meta.Size,
		URL:  info.Location,
	}

	err = s.repo.CreateFile(ctx, f)
	if err != nil {
		return nil, err
	}

	return &file.UploadResponse{
		Id:  int32(f.Id),
		Url: info.Location,
	}, nil
}
