package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"path"
	"strconv"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/file/mws"
)

type FileService struct {
	repo  *repository.UploadRepo
	minio *mws.MinioServer
}

func NewFileService(repo *repository.UploadRepo, minio *mws.MinioServer) *FileService {
	return &FileService{repo: repo, minio: minio}
}

func (s *FileService) UploadFile(ctx context.Context, header *multipart.FileHeader, parentFold, storeId int64) (*dao.File, error) {
	fileName := header.Filename
	// 先查询文件名是否重复
	exists, err := s.repo.QueryByName(ctx, fileName)
	if err != nil {
		return &dao.File{}, err
	}
	if exists {
		return &dao.File{}, err
	}

	file, err := header.Open()
	if err != nil {
		return &dao.File{}, err
	}
	defer file.Close()

	hash := s.getHashFromFile(file)

	exists, err = s.repo.QueryByHash(ctx, hash)

	// 文件基本信息
	suffix := path.Ext(fileName)
	filePrefix := fileName[:len(fileName)-len(suffix)]
	fileSize := header.Size

	var sizeStr string
	if fileSize < 1048576 {
		sizeStr = strconv.FormatInt(fileSize/1024, 10) + "KB"
	} else {
		sizeStr = strconv.FormatInt(fileSize/102400, 10) + "MB"
	}

	newFile := &dao.File{
		FileName:       filePrefix,
		FileHash:       hash,
		FileStoreId:    storeId,
		ParentFolderId: parentFold,
		Type:           suffix,
		Size:           fileSize,
		SizeStr:        sizeStr,
	}
	err = s.repo.CreateFile(ctx, newFile)
	if err != nil {
		return &dao.File{}, err
	}

	//// TODO
	//// 上传 OSS

	return newFile, nil
}

func (s *FileService) getHashFromFile(file multipart.File) string {
	hash := sha256.New()
	_, _ = io.Copy(hash, file)

	bytes := hash.Sum(nil)
	hashCode := hex.EncodeToString(bytes)

	return hashCode
}
