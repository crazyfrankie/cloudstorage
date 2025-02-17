package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type File struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"unique"`
	Path  string
	URL   string
	Hash  string
	Type  string
	Size  int64
	Ctime int64 `gorm:"index"`
	Utime int64
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
	err := d.db.WithContext(ctx).Model(&File{}).Where("name = ?", name).First(&file).Error
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
	err := d.db.WithContext(ctx).Model(&File{}).Where("hash = ?", hash).First(&file).Error
	if err != nil {
		return false, err
	}

	if file.Id == 0 {
		return false, nil
	}

	return true, nil
}
