package rpc

import (
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
)

func InitSmClient(cli *clientv3.Client) sm.ShortMsgServiceClient {
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/sm",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return sm.NewShortMsgServiceClient(conn)
}

func InitFileClient(cli *clientv3.Client) file.FileServiceClient {
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/file",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return file.NewFileServiceClient(conn)
}
