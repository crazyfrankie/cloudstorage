package main

import (
	"log"
	"net/http"

	"github.com/crazyfrankie/cloudstorage/app/sm/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	server := ioc.InitServer()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9095", mux); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
