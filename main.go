package main

import (
	"github.com/ipoluianov/mdata/app"
	"github.com/ipoluianov/mdata/application"
	"github.com/ipoluianov/mdata/bybit"
	"github.com/ipoluianov/mdata/logger"
)

func main() {
	bybit.Start()

	application.Name = "mdata"
	application.ServiceName = "mdata"
	application.ServiceDisplayName = "mdata"
	application.ServiceDescription = "mdata"
	application.ServiceRunFunc = app.RunAsService
	application.ServiceStopFunc = app.StopService

	logger.Init(logger.CurrentExePath() + "/logs")

	if !application.TryService() {
		app.RunDesktop()
	}

}
