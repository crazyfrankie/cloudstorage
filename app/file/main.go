package main

import (
	"log"
	"net/http"

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

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9098", mux); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
