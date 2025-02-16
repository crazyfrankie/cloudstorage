package repository

import (
	"context"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
)

type UploadRepo struct {
	dao *dao.UploadDao
}

func NewUploadRepo(dao *dao.UploadDao) *UploadRepo {
	return &UploadRepo{dao: dao}
}

func (r *UploadRepo) CreateFile(ctx context.Context, file *dao.File) error {
	return r.dao.CreateFile(ctx, file)
}

func (r *UploadRepo) QueryByName(ctx context.Context, name string) (bool, error) {
	return r.dao.QueryByName(ctx, name)
}

func (r *UploadRepo) QueryByHash(ctx context.Context, hash string) (bool, error) {
	return r.dao.QueryByHash(ctx, hash)
}
