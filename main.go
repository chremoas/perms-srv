package main

import (
	"fmt"

	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"go.uber.org/zap"

	chremoasPrometheus "github.com/chremoas/services-common/prometheus"

	"github.com/chremoas/perms-srv/handler"
	permsrv "github.com/chremoas/perms-srv/proto"
)

var (
	Version = "SET ME YOU KNOB"
	service micro.Service
	name    = "perms"
	logger  *zap.Logger
)

func main() {
	var err error

	// TODO pick stuff up from the config
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Initialized logger")

	go chremoasPrometheus.PrometheusExporter(logger)

	service = config.NewService(Version, "srv", name, initialize)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(config *config.Configuration) error {
	h, err := handler.NewPermissionsHandler(config, logger)
	if err != nil {
		return err
	}

	permsrv.RegisterPermissionsHandler(service.Server(), h)
	return nil
}
