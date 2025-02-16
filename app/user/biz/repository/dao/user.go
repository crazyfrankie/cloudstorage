package dao

import (
	"context"

	"gorm.io/gorm"
)

type User struct {
	Id     int    `gorm:"primaryKey,autoIncrement"`
	Name   string `gorm:"varchar(255)"`
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
	return u.db.WithContext(ctx).Create(user).Error
}

func (u *UserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (u *UserDao) FindById(ctx context.Context, id int) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return User{}, err
	}

	return user, nil
}
