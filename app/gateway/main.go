package main

import (
	"context"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"syscall"

	"github.com/crazyfrankie/cloudstorage/app/gateway/ioc"
)

func main() {
	handler := ioc.InitServer()

	g := &run.Group{}

	server := &http.Server{
		Addr:    "localhost:9091",
		Handler: handler,
	}
	g.Add(func() error {
		return server.ListenAndServe()
	}, func(err error) {
		if err := server.Close(); err != nil {
			log.Printf("failed to stop web server, err:%s", err)
		}
	})

	fileServer := &http.Server{Addr: ":9096"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			ioc.FileReg,
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

	userServer := &http.Server{Addr: ":9097"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			ioc.UserReg,
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

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
		return
	}
}
