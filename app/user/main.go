package main

import (
	"context"
	"log"
	"net/http"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/cloudstorage/app/user/ioc"
	"github.com/crazyfrankie/cloudstorage/app/user/rpc"
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

	userServer := &http.Server{Addr: ":9092"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.UserReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		userServer.Handler = mux
		return userServer.ListenAndServe()
	}, func(err error) {
		if err := userServer.Close(); err != nil {
			log.Printf("failed to stop web server, err:%s", err)
		}
	})

	fileServer := &http.Server{Addr: ":9093"}
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

	smServer := &http.Server{Addr: ":9094"}
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
