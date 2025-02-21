package main

import (
	"context"
	"log"
	"net/http"
	"syscall"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/cloudstorage/app/sm/ioc"
	"github.com/crazyfrankie/cloudstorage/app/sm/rpc"
)

func main() {
	server := ioc.InitServer()

	g := &run.Group{}

	g.Add(func() error {
		return server.Serve()
	}, func(err error) {
		server.Server.GracefulStop()
		server.Server.Stop()
	})

	smServer := &http.Server{Addr: ":9095"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.SmReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		smServer.Handler = mux
		return smServer.ListenAndServe()
	}, func(err error) {
		if err := smServer.Close(); err != nil {
			log.Printf("failed to stop web server, err:%s", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
		return
	}
}
