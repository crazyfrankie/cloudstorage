package repository

import (
	"context"
	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/cache"

	"github.com/crazyfrankie/cloudstorage/app/file/biz/repository/dao"
)

type UploadRepo struct {
	dao   *dao.UploadDao
	cache *cache.FileCache
}

func NewUploadRepo(dao *dao.UploadDao, cache *cache.FileCache) *UploadRepo {
	return &UploadRepo{dao: dao, cache: cache}
}

func (r *UploadRepo) CreateFile(ctx context.Context, file *dao.File) error {
	return r.dao.CreateFile(ctx, file)
}

func (r *UploadRepo) GetFile(ctx context.Context, fid int64, uid int32) (dao.File, error) {
	return r.dao.GetFile(ctx, fid, uid)
}

func (r *UploadRepo) QueryByHash(ctx context.Context, hash string) (dao.File, error) {
	return r.dao.QueryByHash(ctx, hash)
}

func (r *UploadRepo) QueryCapacity(ctx context.Context, uid int32, size int64) (bool, error) {
	return r.dao.QueryCapacity(ctx, uid, size)
}

func (r *UploadRepo) CreateFileStore(ctx context.Context, store *dao.FileStore) (int32, error) {
	return r.dao.CreateFileStore(ctx, store)
}

func (r *UploadRepo) CreateFolder(ctx context.Context, folder *dao.Folder) error {
	return r.dao.CreateFolder(ctx, folder)
}

func (r *UploadRepo) MoveFile(ctx context.Context, fileId, toFolderId int64, uid int32) error {
	return r.dao.MoveFile(ctx, fileId, toFolderId, uid)
}

func (r *UploadRepo) MoveFolder(ctx context.Context, folderId, toFolderId int64, uid int32, name string) error {
	return r.dao.MoveFolder(ctx, folderId, toFolderId, uid, name)
}

func (r *UploadRepo) DeleteFile(ctx context.Context, fileId int64, uid int32) error {
	return r.dao.DeleteFile(ctx, fileId, uid)
}

func (r *UploadRepo) DeleteFolder(ctx context.Context, folderId int64, uid int32) error {
	return r.dao.DeleteFolder(ctx, folderId, uid)
}

func (r *UploadRepo) Search(ctx context.Context, uid int32, query string, page, size int32) ([]dao.File, []dao.Folder, error) {
	return r.dao.Search(ctx, uid, query, page, size)
}

func (r *UploadRepo) ListFolder(ctx context.Context, folderId int64, userId int32) ([]*dao.File, []*dao.Folder, error) {
	return r.dao.ListFolder(ctx, folderId, userId)
}

func (r *UploadRepo) GetNextDownloadTask(ctx context.Context) (string, error) {
	return r.cache.GetNextDownloadTask(ctx)
}

func (r *UploadRepo) GetDownloadTaskInfo(ctx context.Context, taskId string) (*cache.DownloadTask, error) {
	return r.cache.GetDownloadTaskInfo(ctx, taskId)
}

func (r *UploadRepo) CreateDownloadTask(ctx context.Context, taskId string, info *cache.DownloadTask) error {
	return r.cache.CreateDownloadTask(ctx, taskId, info)
}

func (r *UploadRepo) UpdateTaskStatus(ctx context.Context, taskId string, status string, progress int64) error {
	return r.cache.UpdateTaskStatus(ctx, taskId, status, progress)
}
