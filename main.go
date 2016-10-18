package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/yunify/docker-machine-driver-qingcloud/qingcloud"
)

func main() {
	plugin.RegisterDriver(qingcloud.NewDriver("", ""))
}
