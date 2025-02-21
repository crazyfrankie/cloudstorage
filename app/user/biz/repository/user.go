package repository

import (
	"context"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository/dao"
)

type UserRepo struct {
	dao *dao.UserDao
}

func NewUserRepo(dao *dao.UserDao) *UserRepo {
	return &UserRepo{dao: dao}
}

func (r *UserRepo) Create(ctx context.Context, user *dao.User) error {
	return r.dao.Create(ctx, user)
}

func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (dao.User, error) {
	return r.dao.FindByPhone(ctx, phone)
}

func (r *UserRepo) FindById(ctx context.Context, id int) (dao.User, error) {
	return r.dao.FindById(ctx, id)
}

func (r *UserRepo) UpdateInfo(ctx context.Context, u *dao.User) error {
	return r.dao.UpdateInfo(ctx, u)
}
