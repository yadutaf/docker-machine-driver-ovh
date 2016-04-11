package main

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

// Default values for docker-machine-driver-ovh
const (
	DefaultSecurityGroup = "default"
	DefaultProjectName   = "docker-machine"
	DefaultFlavorName    = "vps-ssd-1"
	DefaultRegionName    = "GRA1"
	ImageName            = "Ubuntu 14.04"
	SSHUserName          = "admin"
)

func main() {
	plugin.RegisterDriver(&Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: SSHUserName,
			SSHPort: 22,
		}})
}
