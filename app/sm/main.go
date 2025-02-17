package main

import (
	"github.com/crazyfrankie/cloudstorage/app/sm/ioc"
)

func main() {
	server := ioc.InitServer()

	err := server.Serve()
	if err != nil {
		panic(err)
	}
}
