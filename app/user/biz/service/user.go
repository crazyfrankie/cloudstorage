package service

import (
	"context"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/user/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
)

var (
	defaultAvatar = "github.com/crazyfrankie/cloud/default.png"
)

type UserServer struct {
	repo *repository.UserRepo
	sm   sm.ShortMsgServiceClient
	file file.FileServiceClient
	user.UnimplementedUserServiceServer
}

func NewUserServer(repo *repository.UserRepo, sm sm.ShortMsgServiceClient, file file.FileServiceClient) *UserServer {
	return &UserServer{repo: repo, sm: sm, file: file}
}

func (s *UserServer) SendCode(ctx context.Context, req *user.SendCodeRequest) (*user.SendCodeResponse, error) {
	phone := req.GetPhone()
	u, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	var biz string
	if u.Id == 0 {
		biz = "register"
	} else {
		biz = "login"
	}

	_, err = s.sm.SendSm(ctx, &sm.SendSmRequest{Biz: biz, Phone: phone})
	if err != nil {
		return nil, err
	}

	return &user.SendCodeResponse{Biz: biz}, nil
}

func (s *UserServer) VerifyCode(ctx context.Context, req *user.VerifyCodeRequest) (*user.VerifyCodeResponse, error) {
	phone, code, biz := req.GetPhone(), req.GetCode(), req.GetBiz()
	_, err := s.sm.VerifySm(ctx, &sm.VerifySmRequest{Biz: biz, Phone: phone, Code: code})
	if err != nil {
		return nil, err
	}

	var uid int
	if biz == "register" {
		u := &dao.User{
			Phone:  phone,
			Name:   phone,
			Avatar: defaultAvatar,
		}
		err = s.repo.Create(ctx, u)
		if err != nil {
			return nil, err
		}

		_, err = s.file.CreateFileStore(ctx, &file.CreateFileStoreRequest{UserId: int32(u.Id)})
		if err != nil {
			return nil, err
		}
		uid = u.Id
	} else {
		u, err := s.repo.FindByPhone(ctx, phone)
		if err != nil {
			return nil, err
		}
		uid = u.Id
	}

	var token string
	token, err = mws.GenerateToken(int32(uid))
	if err != nil {
		return nil, err
	}

	return &user.VerifyCodeResponse{Token: token}, nil
}
