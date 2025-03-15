package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id     int    `gorm:"primaryKey,autoIncrement"`
	Name   string `gorm:"unique;type:varchar(255)"`
	Phone  string `gorm:"unique"`
	Avatar string
	Ctime  int64
	Utime  int64
}

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

func (u *UserDao) Create(ctx context.Context, user *User) error {
	now := time.Now().Unix()
	user.Ctime = now
	user.Utime = now
	return u.db.WithContext(ctx).Create(user).Error
}

func (u *UserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("phone = ?", phone).Find(&user).Error
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (u *UserDao) FindById(ctx context.Context, id int) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("id = ?", id).Find(&user).Error
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (u *UserDao) UpdateInfo(ctx context.Context, user *User) error {
	updates := make(map[string]any, 3)
	if user.Name != "" {
		updates["name"] = user.Name
	}
	if user.Avatar != "" {
		updates["avatar"] = user.Avatar
	}

	if len(updates) > 0 {
		updates["utime"] = time.Now().Unix()
	}

	return u.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.Id).Updates(updates).Error
}
