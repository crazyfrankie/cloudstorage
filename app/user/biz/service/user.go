package service

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/repository"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
)

var (
	defaultAvatar = "github.com/crazyfrankie/cloud/default.png"
)

type UserServer struct {
	repo *repository.UserRepo
	sm   sm.ShortMsgServiceClient
	user.UnimplementedUserServiceServer
}

func NewUserServer(repo *repository.UserRepo, sm sm.ShortMsgServiceClient) *UserServer {
	return &UserServer{repo: repo, sm: sm}
}

func (s *UserServer) RegisterService(server *grpc.Server) {
	user.RegisterUserServiceServer(server, s)
}

func (s *UserServer) SendCode(ctx context.Context, req *user.SendCodeRequest) (*user.SendCodeResponse, error) {
	return nil, nil
}

func (s *UserServer) VerifyCode(ctx context.Context, req *user.VerifyCodeRequest) (*user.VerifyCodeResponse, error) {
	return nil, nil
}
