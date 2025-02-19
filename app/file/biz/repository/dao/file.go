package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type File struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	UserId   int32  `gorm:"index:uid_fld_status"`
	Name     string `gorm:"unique"` // 修改为同文件夹下唯一
	FolderId int64  `gorm:"index:uid_fld_status"`
	Path     string
	Hash     string `gorm:"index"` // 添加索引以优化秒传查询
	Type     string
	Size     int64
	Status   int8 `gorm:"index:uid_fld_status"`
	Ctime    int64
	Utime    int64
}

type Folder struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Name     string
	ParentId int64  `gorm:"index:uid_pid_status"` // 添加联合索引
	UserId   int32  `gorm:"index:uid_pid_status"`
	Path     string `gorm:"index"` // 添加索引支持路径搜索
	Status   int    `gorm:"index:uid_pid_status"`
	Ctime    int64
	Utime    int64
}

type FileStore struct {
	Id          int   `gorm:"primaryKey,autoIncrement"`
	UserId      int32 `gorm:"unique"`
	Capacity    int64 `gorm:"default:10737418240"`
	CurrentSize int64
	Ctime       int64
	Utime       int64
}

type UploadDao struct {
	db *gorm.DB
}

func NewUploadDao(db *gorm.DB) *UploadDao {
	return &UploadDao{db: db}
}

func (d *UploadDao) CreateFile(ctx context.Context, file *File) error {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		file.Ctime = now
		file.Utime = now
		err := tx.WithContext(ctx).Model(&File{}).Create(file).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Model(&FileStore{}).Where("user_id = ?", file.UserId).Update("current_size", gorm.Expr(fmt.Sprintf("current_size + %d", file.Size))).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *UploadDao) GetFile(ctx context.Context, fid int64, uid int32) (File, error) {
	var file File
	err := d.db.WithContext(ctx).Model(&File{}).Where("id = ? AND user_id = ?", fid, uid).Find(&file).Error
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func (d *UploadDao) QueryByHash(ctx context.Context, hash string) (File, error) {
	var file File
	err := d.db.WithContext(ctx).Model(&File{}).Where("hash = ?", hash).Find(&file).Error
	if err != nil {
		return File{}, err
	}

	return file, nil
}

func (d *UploadDao) QueryCapacity(ctx context.Context, uid int32, size int64) (bool, error) {
	var store FileStore
	err := d.db.WithContext(ctx).Model(&FileStore{}).Where("user_id = ?", uid).Find(&store).Error
	if err != nil {
		return false, err
	}
	if store.Capacity < size+store.CurrentSize {
		return false, nil
	}

	return true, nil
}

func (d *UploadDao) CreateFileStore(ctx context.Context, store *FileStore) (int32, error) {
	now := time.Now().Unix()
	store.Ctime = now
	store.Utime = now
	err := d.db.WithContext(ctx).Model(&FileStore{}).Create(store).Error
	if err != nil {
		return 0, err
	}

	return int32(store.Id), nil
}

func (d *UploadDao) CreateFolder(ctx context.Context, folder *Folder) error {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		folder.Ctime = now
		folder.Utime = now

		var parent Folder
		err := tx.WithContext(ctx).Model(&Folder{}).Where("id = ? AND user_id = ?", folder.ParentId, folder.UserId).Find(&parent).Error
		if err != nil {
			return err
		}

		if parent.Path == "" {
			folder.Path = folder.Name
		}
		folder.Path = parent.Path + "/" + folder.Name
		err = tx.WithContext(ctx).Model(&Folder{}).Create(folder).Error

		return err
	})

	return err
}

func (d *UploadDao) MoveFile(ctx context.Context, fileId, toFolderId int64, uid int32) error {
	return d.db.WithContext(ctx).Model(&File{}).
		Where("user_id = ? AND id = ?", uid, fileId).
		Update("folder_id", toFolderId).Error
}

func (d *UploadDao) MoveFolder(ctx context.Context, folderId, toFolderId int64, uid int32, name string) error {
	if folderId == toFolderId {
		return errors.New("cannot move folder to itself")
	}

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var toFolder Folder
		if err := tx.WithContext(ctx).Model(&Folder{}).
			Where("id = ?", toFolderId).
			First(&toFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("target folder not found")
			}
		}

		// 检查源文件夹是否存在
		var sourceFolder Folder
		if err := tx.Model(&Folder{}).
			Where("id = ? AND user_id = ?", folderId, uid).
			First(&sourceFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("source folder not found")
			}
			return err
		}

		if toFolder.ParentId == folderId {
			return errors.New("cannot move folder to its subfolder")
		}

		err := tx.WithContext(ctx).Model(&File{}).Where("folder_id = ? AND user_id = ?", folderId, uid).
			Update("folder_id", toFolderId).Error
		if err != nil {
			return err
		}

		// 更新文件夹路径
		newPath := "/"
		if toFolderId != 0 {
			newPath = toFolder.Path + "/" + name
		} else {
			newPath = "/" + name
		}

		// 更新所有子文件夹的路径
		return tx.Model(&Folder{}).
			Where("path LIKE ? AND user_id = ?", sourceFolder.Path+"/%", uid).
			Update("path", gorm.Expr(
				"CONCAT(?, SUBSTR(path, ?))",
				newPath,
				len(sourceFolder.Path)+1,
			)).Error
		//path = toFolder.Path + "/" + name
		//err = tx.WithContext(ctx).Model(&Folder{}).
		//	Where("user_id = ? AND id = ?", uid, folderId).
		//	Updates(map[string]any{
		//		"parent_id": toFolderId,
		//		"path":      path,
		//	}).Error
		//return err
	})

	return err
}

func (d *UploadDao) DeleteFile(ctx context.Context, fileId int64, uid int32) error {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var file File
		err := tx.WithContext(ctx).Model(&File{}).Where("id = ? AND user_id = ?", fileId, uid).Find(&file).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Model(&File{}).
			Where("id = ? AND user_id = ?", fileId, uid).Update("status", 1).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Model(&FileStore{}).Where("user_id = ?", uid).Update("current_size", gorm.Expr(fmt.Sprintf("current_size - %d", file.Size))).Error
		return err
	})

	return err
}

func (d *UploadDao) DeleteFolder(ctx context.Context, folderId int64, uid int32) error {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var totalSize struct {
			Size int64
		}
		if err := tx.Model(&File{}).
			Select("COALESCE(SUM(size), 0) as size").
			Where("user_id = ? AND folder_id = ? AND status = 0", uid, folderId).
			Scan(&totalSize).Error; err != nil {
			return err
		}

		// 批量更新文件状态
		if err := tx.Model(&File{}).
			Where("user_id = ? AND folder_id = ?", uid, folderId).
			Update("status", 1).Error; err != nil {
			return err
		}

		// 更新存储空间
		if totalSize.Size > 0 {
			if err := tx.Model(&FileStore{}).
				Where("user_id = ?", uid).
				Update("current_size", gorm.Expr("current_size - ?", totalSize.Size)).
				Error; err != nil {
				return err
			}
		}

		// 更新文件夹状态
		return tx.Model(&Folder{}).
			Where("id = ? AND user_id = ?", folderId, uid).
			Update("status", 1).Error
	})

	return err
}

func (d *UploadDao) Search(ctx context.Context, uid int32, query string, page, size int32) ([]File, []Folder, error) {
	var files []File
	var folders []Folder
	offset := (page - 1) * size

	// 搜索文件
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND status = 0 AND name LIKE ?", uid, "%"+query+"%").
		Order("ctime DESC").
		Offset(int(offset)).
		Limit(int(size)).
		Find(&files).Error
	if err != nil {
		return nil, nil, err
	}

	// 搜索文件夹
	err = d.db.WithContext(ctx).
		Where("user_id = ? AND status = 0 AND name LIKE ?", uid, "%"+query+"%").
		Order("ctime DESC").
		Find(&folders).Error
	if err != nil {
		return nil, nil, err
	}

	return files, folders, nil
}

func (d *UploadDao) ListFolder(ctx context.Context, folderId int64, userId int32) ([]*File, []*Folder, error) {
	var files []*File
	var folders []*Folder

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&File{}).Where("folder_id = ? AND user_id = ? AND status = ?", folderId, userId, 0).Find(&files).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Model(&Folder{}).Where("parent_id = ? AND user_id = ? AND status = ?", folderId, userId, 0).Find(&folders).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return []*File{}, []*Folder{}, err
	}

	return files, folders, nil
}
