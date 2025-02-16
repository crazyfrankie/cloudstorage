package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type File struct {
	Id             int64  `gorm:"primaryKey,autoIncrement"`
	FileName       string `gorm:"unique"`
	FileHash       string
	FileStoreId    int64
	DownloadCount  int64
	ParentFolderId int64
	Type           string
	Size           int64
	SizeStr        string
	Ctime          int64 `gorm:"index"`
}

type UploadDao struct {
	db *gorm.DB
}

func NewUploadDao(db *gorm.DB) *UploadDao {
	return &UploadDao{db: db}
}

func (d *UploadDao) CreateFile(ctx context.Context, file *File) error {
	now := time.Now().Unix()
	file.Ctime = now

	return d.db.WithContext(ctx).Create(file).Error
}

func (d *UploadDao) QueryByName(ctx context.Context, name string) (bool, error) {
	var file File
	err := d.db.WithContext(ctx).Model(&File{}).Where("file_name = ?", name).First(&file).Error
	if err != nil {
		return false, err
	}

	if file.Id == 0 {
		return false, nil
	}

	return true, nil
}

func (d *UploadDao) QueryByHash(ctx context.Context, hash string) (bool, error) {
	var file File
	err := d.db.WithContext(ctx).Model(&File{}).Where("file_hash = ?", hash).First(&file).Error
	if err != nil {
		return false, err
	}

	if file.Id == 0 {
		return false, nil
	}

	return true, nil
}
