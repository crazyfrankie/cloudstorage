package service

import (
	"context"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/sm/biz/repository"
	"github.com/crazyfrankie/cloudstorage/app/sm/biz/service/sms"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
)

type SmServer struct {
	repo *repository.SmRepo
	sms  sms.Service
	sm.UnimplementedShortMsgServiceServer
}

func NewSmServer(repo *repository.SmRepo, sms sms.Service) *SmServer {
	return &SmServer{repo: repo, sms: sms}
}

func (s *SmServer) RegisterService(server *grpc.Server) {
	sm.RegisterShortMsgServiceServer(server, s)
}

func (s *SmServer) SendSm(ctx context.Context, request *sm.SendSmRequest) (*sm.SendSmResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SmServer) VerifySm(ctx context.Context, request *sm.VerifySmRequest) (*sm.VerifySmResponse, error) {
	//TODO implement me
	panic("implement me")
}
