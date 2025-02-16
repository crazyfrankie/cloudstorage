package rpc

import (
	"net"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/file/config"
)

type Server struct {
	*grpc.Server
	Addr   string
	client *clientv3.Client
}

func NewServer(client *clientv3.Client) *Server {
	s := grpc.NewServer()

	return &Server{
		Server: s,
		Addr:   config.GetConf().Server.Addr,
		client: client,
	}
}

func (s *Server) Serve() error {
	conn, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic(err)
	}

	err = registerService(s.client, s.Addr)
	if err != nil {
		panic(err)
	}

	return s.Server.Serve(conn)
}

func registerService(cli *clientv3.Client, port string) error {
	return nil
}
