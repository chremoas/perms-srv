package main

import (
	"fmt"
	"github.com/chremoas/services-common/config"
	permsrv "github.com/chremoas/perms-srv/proto"
	"github.com/chremoas/perms-srv/handler"
)

var Version = "1.0.0"
var name = "perms"

func main() {
	service := config.NewService(Version, "srv", name, config.NilInit)

	permsrv.RegisterPermissionsHandler(service.Server(), handler.NewPermissionsHandler())

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
