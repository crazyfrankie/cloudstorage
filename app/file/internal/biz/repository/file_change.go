package repository

import (
	"context"
	"time"

	"github.com/crazyfrankie/cloudstorage/app/file/internal/biz/repository/dao"
)

// SaveFileChange 保存文件变更记录
func (r *UploadRepo) SaveFileChange(ctx context.Context, change *dao.FileChange) error {
	change.CreatedAt = time.Now().Unix()
	return r.dao.SaveFileChange(ctx, change)
}

// GetFileChanges 获取文件变更记录
func (r *UploadRepo) GetFileChanges(ctx context.Context, fileID int64, fromVersion int64) ([]dao.FileChange, error) {
	return r.dao.GetFileChanges(ctx, fileID, fromVersion)
}

// ApplyFileChanges 在服务端应用文件变更
func (r *UploadRepo) ApplyFileChanges(ctx context.Context, fileID int64, changes []dao.FileChange) error {
	return r.dao.ApplyFileChanges(ctx, fileID, changes)
}
