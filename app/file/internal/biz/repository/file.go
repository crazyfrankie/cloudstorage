package repository

import (
	"context"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository/cache"
	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository/dao"
)

type UploadRepo struct {
	dao   *dao.UploadDao
	cache *cache.FileCache
}

func NewUploadRepo(dao *dao.UploadDao, cache *cache.FileCache) *UploadRepo {
	return &UploadRepo{dao: dao, cache: cache}
}

// CreateFile 创建文件记录
func (r *UploadRepo) CreateFile(ctx context.Context, file *dao.File) error {
	return r.dao.CreateFile(ctx, file)
}

// GetFile 获取文件信息
func (r *UploadRepo) GetFile(ctx context.Context, fid int64, uid int32) (dao.File, error) {
	return r.dao.GetFile(ctx, fid, uid)
}

// QueryByHash 根据文件 hash 查询文件是否存在
func (r *UploadRepo) QueryByHash(ctx context.Context, hash string) (dao.File, error) {
	return r.dao.QueryByHash(ctx, hash)
}

// QueryCapacity 查询用户空间容量
func (r *UploadRepo) QueryCapacity(ctx context.Context, uid int32, size int64) (bool, error) {
	return r.dao.QueryCapacity(ctx, uid, size)
}

// CreateFileStore 创建用户资源空间
func (r *UploadRepo) CreateFileStore(ctx context.Context, store *dao.FileStore) (int32, error) {
	return r.dao.CreateFileStore(ctx, store)
}

// CreateFolder 创建文件夹
func (r *UploadRepo) CreateFolder(ctx context.Context, folder *dao.Folder) error {
	return r.dao.CreateFolder(ctx, folder)
}

// MoveFile 移动文件
func (r *UploadRepo) MoveFile(ctx context.Context, fileId, toFolderId int64, uid int32) error {
	return r.dao.MoveFile(ctx, fileId, toFolderId, uid)
}

// MoveFolder 移动文件夹
func (r *UploadRepo) MoveFolder(ctx context.Context, folderId, toFolderId int64, uid int32, name string) error {
	return r.dao.MoveFolder(ctx, folderId, toFolderId, uid, name)
}

// DeleteFile 删除文件
func (r *UploadRepo) DeleteFile(ctx context.Context, fileId int64, uid int32) error {
	return r.dao.DeleteFile(ctx, fileId, uid)
}

// DeleteFolder 删除文件夹
func (r *UploadRepo) DeleteFolder(ctx context.Context, folderId int64, uid int32) error {
	return r.dao.DeleteFolder(ctx, folderId, uid)
}

// Search 查找文件（文件夹)
func (r *UploadRepo) Search(ctx context.Context, uid int32, query string, page, size int32) ([]dao.File, []dao.Folder, error) {
	return r.dao.Search(ctx, uid, query, page, size)
}

// ListFolder 展示文件夹及文件
func (r *UploadRepo) ListFolder(ctx context.Context, folderId int64, userId int32) ([]*dao.File, []*dao.Folder, error) {
	return r.dao.ListFolder(ctx, folderId, userId)
}

// GetNextDownloadTask 获取下一个下载任务
func (r *UploadRepo) GetNextDownloadTask(ctx context.Context) (string, error) {
	return r.cache.GetNextDownloadTask(ctx)
}

// GetDownloadTaskInfo 获取下载任务信息
func (r *UploadRepo) GetDownloadTaskInfo(ctx context.Context, taskId string) (*cache.DownloadTask, error) {
	return r.cache.GetDownloadTaskInfo(ctx, taskId)
}

// CreateDownloadTask 创建下载任务
func (r *UploadRepo) CreateDownloadTask(ctx context.Context, taskId string, info *cache.DownloadTask) error {
	return r.cache.CreateDownloadTask(ctx, taskId, info)
}

// UpdateTaskStatus 更新下载任务状态
func (r *UploadRepo) UpdateTaskStatus(ctx context.Context, taskId string, status string, progress int64) error {
	return r.cache.UpdateTaskStatus(ctx, taskId, status, progress)
}

// SavePartETag 保存分片标签
func (r *UploadRepo) SavePartETag(ctx context.Context, uploadId string, part int, etag string) error {
	return r.cache.SavePartETag(ctx, uploadId, part, etag)
}

// GetPartETags 获取分片标签
func (r *UploadRepo) GetPartETags(ctx context.Context, uploadId string) (map[int]string, error) {
	return r.cache.GetPartETags(ctx, uploadId)
}

func (r *UploadRepo) GetFilesByIds(ctx context.Context, files []int64) ([]*dao.File, error) {
	return nil, nil
}

func (r *UploadRepo) CreateShareLink(ctx context.Context, share *dao.ShareLink) error {
	return nil
}

func (r *UploadRepo) CreateShareFile(ctx context.Context, share *dao.ShareFile) error {
	return nil
}

func (r *UploadRepo) GetShareLink(ctx context.Context, shareId string) (dao.ShareLink, error) {
	return dao.ShareLink{}, nil
}

func (r *UploadRepo) FindFileStoreById(ctx context.Context, uid int32) (dao.FileStore, error) {
	return r.dao.FindFileStoreById(ctx, uid)
}
