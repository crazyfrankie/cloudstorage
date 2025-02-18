package ioc

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
)

func InitFileClient(cli *clientv3.Client) file.FileServiceClient {
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/file",
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallRecvMsgSize(10*1024*1024),
		),
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	return file.NewFileServiceClient(conn)
}

func InitUserClient(cli *clientv3.Client) user.UserServiceClient {
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return user.NewUserServiceClient(conn)
}
