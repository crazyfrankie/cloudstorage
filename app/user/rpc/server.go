package rpc

import (
	"context"
	"fmt"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"log"
	"net"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/user/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/user/config"
)

var (
	serviceName = "service/user"
)

type Server struct {
	*grpc.Server
	Addr   string
	client *clientv3.Client
}

func NewServer(u *service.UserServer, client *clientv3.Client) *Server {
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, u)

	return &Server{
		Server: s,
		Addr:   config.GetConf().Server.Addr,
		client: client,
	}
}

func (s *Server) Serve() error {
	conn, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	err = registerService(s.client, s.Addr)
	if err != nil {
		return err
	}

	return s.Server.Serve(conn)
}

func registerService(cli *clientv3.Client, port string) error {
	em, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		return err
	}

	addr := "127.0.0.1" + port
	serviceKey := serviceName + "/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leaseResp, err := cli.Grant(ctx, 180)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, serviceKey, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch, err := cli.KeepAlive(ctx, leaseResp.ID)
		if err != nil {
			log.Printf("keep alive failed lease id:%d", leaseResp.ID)
			return
		}
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					log.Println("KeepAlive channel closed")
					return
				}
				fmt.Println("Lease renewed")
			case <-ctx.Done():
				return
			}
		}
	}()

	return err
}
