package service

import (
	"context"
	"log"
	"sync"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository/dao"
	"github.com/crazyfrankie/cloudstorage/app/user/mws"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
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

func (s *UserServer) GetUserInfo(ctx context.Context, req *user.GetUserInfoRequest) (*user.GetUserInfoResponse, error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var err error
	var u dao.User
	var resp *file.GetUserFileStoreResponse
	go func() {
		u, err = s.repo.FindById(ctx, int(req.GetUserId()))
		if err != nil {
			log.Printf("failed get user info, %s", err)
		}
		wg.Done()
	}()

	go func() {
		resp, err = s.file.GetUserFileStore(ctx, &file.GetUserFileStoreRequest{UserId: req.GetUserId()})
		if err != nil {
			log.Printf("failed get user file store, %s", err)
		}
		wg.Done()
	}()

	return &user.GetUserInfoResponse{
		User: &user.User{
			Id:     int32(u.Id),
			Name:   u.Name,
			Phone:  u.Phone,
			Avatar: u.Avatar,
		},
		FileStore: resp.GetFileStore(),
	}, nil
}

func (s *UserServer) UpdateInfo(ctx context.Context, req *user.UpdateInfoRequest) (*user.UpdateInfoResponse, error) {
	err := s.repo.UpdateInfo(ctx, &dao.User{
		Id:     int(req.GetUserId()),
		Name:   req.GetName(),
		Avatar: req.GetAvatar(),
	})
	if err != nil {
		return nil, err
	}

	return &user.UpdateInfoResponse{}, nil
}
