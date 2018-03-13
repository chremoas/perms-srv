package main

import (
	"fmt"
	"github.com/chremoas/services-common/config"
	permsrv "github.com/chremoas/perms-srv/proto"
	"github.com/chremoas/perms-srv/handler"
	"github.com/micro/go-micro"
)

var Version = "1.0.0"
var service micro.Service
var name = "perms"

func main() {
	service = config.NewService(Version, "srv", name, initialize)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(config *config.Configuration) error {
	permsrv.RegisterPermissionsHandler(service.Server(), handler.NewPermissionsHandler(config))
	return nil
}