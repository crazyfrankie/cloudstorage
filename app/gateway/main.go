package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/cloudstorage/app/gateway/ioc"
)

func main() {
	handler := ioc.InitServer()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			ioc.FileReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9096", mux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			ioc.UserReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9097", mux); err != nil {
			log.Fatal(err)
		}
	}()

	server := &http.Server{
		Addr:    "localhost:9091",
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server start failed %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutting down err:%s\n", err)
	}

	log.Printf("Server exited gracefully")
}
