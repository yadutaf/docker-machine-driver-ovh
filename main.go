package main

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

const (
	DefaultSecurityGroup = "default"
	DefaultProjectName   = "docker-machine"
	DefaultFlavorName    = "vps-ssd-1"
	DefaultRegionName    = "GRA1"
	ImageName            = "Ubuntu 14.04"
	SshUserName          = "admin"
)

func main() {
	plugin.RegisterDriver(&Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: SshUserName,
			SSHPort: 22,
		}})
}
