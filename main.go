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
	DefaultImageName     = "Ubuntu 16.04"
	DefaultSSHUserName   = "ubuntu"
	DefaultBillingPeriod = "hourly"
	DefaultEndpoint      = "ovh-eu"
)

func main() {
	plugin.RegisterDriver(&Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: DefaultSSHUserName,
			SSHPort: 22,
		}})
}
