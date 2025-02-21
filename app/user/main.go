package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
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

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.UserReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9092", mux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.FileReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9093", mux); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.SmReg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		if err := http.ListenAndServe(":9094", mux); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
