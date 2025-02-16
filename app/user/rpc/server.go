package rpc

import (
	"net"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/user/config"
)

type Server struct {
	*grpc.Server
	Addr   string
	client *clientv3.Client
}

func NewServer(user *service.UserServer, client *clientv3.Client) *Server {
	s := grpc.NewServer()
	user.RegisterService(s)

	return &Server{
		Server: s,
		Addr:   config.GetConf().Server.Addr,
		client: client,
	}
}

func (s *Server) Serve() error {
	_, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	err = registerService(s.client, s.Addr)
	if err != nil {
		return err
	}

	return nil
}

func registerService(cli *clientv3.Client, port string) error {
	return nil
}
