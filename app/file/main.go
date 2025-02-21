package main

import (
	"context"
	"github.com/crazyfrankie/cloudstorage/app/file/rpc"
	"github.com/oklog/run"
	"log"
	"net/http"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/cloudstorage/app/file/ioc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	server := ioc.InitServer()

	g := &run.Group{}

	g.Add(func() error {
		return server.Serve()
	}, func(err error) {
		server.Server.GracefulStop()
		server.Server.Stop()
	})

	fileServer := &http.Server{Addr: ":9098"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.FileReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		fileServer.Handler = mux
		return fileServer.ListenAndServe()
	}, func(err error) {
		if err := fileServer.Close(); err != nil {
			log.Printf("failed to stop web server, err:%s", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
		return
	}
}
