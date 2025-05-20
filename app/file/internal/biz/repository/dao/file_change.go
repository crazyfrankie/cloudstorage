package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SaveFileChange 保存文件变更记录
func (d *UploadDao) SaveFileChange(ctx context.Context, change *FileChange) error {
	return d.db.WithContext(ctx).Create(change).Error
}

// GetFileChanges 获取文件变更记录
func (d *UploadDao) GetFileChanges(ctx context.Context, fileID int64, fromVersion int64) ([]FileChange, error) {
	var changes []FileChange
	err := d.db.WithContext(ctx).
		Where("file_id = ? AND version > ?", fileID, fromVersion).
		Order("version ASC").
		Find(&changes).Error
	return changes, err
}

// ApplyFileChanges 在服务端应用文件变更
func (d *UploadDao) ApplyFileChanges(ctx context.Context, fileID int64, changes []FileChange) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取当前文件
		var file File
		if err := tx.WithContext(ctx).Where("id = ?", fileID).First(&file).Error; err != nil {
			return err
		}

		// 记录原始大小用于计算变化
		originalSize := file.Size

		// 应用变更到文件
		for _, change := range changes {
			// 更新文件版本号
			if change.Version > int64(file.Version) {
				file.Version = int32(change.Version)
			}

			// 更新设备ID和最后修改者
			file.DeviceId = change.DeviceID
			file.LastModifiedBy = change.DeviceID

			// 更新文件大小（简化处理，实际上应该根据变更内容计算）
			// 这里仅示例，实际实现需要根据变更内容调整文件大小
			switch change.Operation {
			case Insert:
				file.Size += int64(len(change.Content))
			case Delete:
				file.Size -= change.Length
			case Update:
				file.Size = file.Size - change.Length + int64(len(change.Content))
			}
		}

		// 更新文件记录
		file.Utime = time.Now().Unix()
		if err := tx.WithContext(ctx).Save(&file).Error; err != nil {
			return err
		}

		// 更新用户存储空间使用量
		sizeDiff := file.Size - originalSize
		if sizeDiff != 0 {
			expr := "current_size + ?"
			if sizeDiff < 0 {
				expr = "current_size - ?"
				sizeDiff = -sizeDiff
			}

			if err := tx.Model(&FileStore{}).Where("user_id = ?", file.UserId).
				Update("current_size", gorm.Expr(expr, sizeDiff)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
