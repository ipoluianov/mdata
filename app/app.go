package app

import (
	"fmt"

	"github.com/ipoluianov/mdata/httpserver"
	"github.com/ipoluianov/mdata/logger"
)

var httpServer *httpserver.HttpServer

func Start() {
	TuneFDs()
	httpServer = httpserver.NewHttpServer()
	httpServer.Start()
}

func Stop() {
	httpServer.Stop()
}

func RunDesktop() {
	logger.Println("Running as console application")
	Start()
	fmt.Scanln()
	logger.Println("Console application exit")
}

func RunAsService() error {
	Start()
	return nil
}

func StopService() {
	Stop()
}
