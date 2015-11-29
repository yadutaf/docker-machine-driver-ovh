package main

import (
	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

const (
	AuthURL              = "https://auth.cloud.ovh.net/v2.0/"
	DefaultSecurityGroup = "default"
	DefaultProjectName   = "docker-machine"
	DefaultFlavorName    = "vps-ssd-1"
	DefaultRegionName    = "GRA1"
	NetworkName          = "Ext-Net"
	ImageName            = "Ubuntu 14.04"
	SshUserName          = "admin"
)

func main() {
	plugin.RegisterDriver(&Driver{
		Driver: openstack.NewDerivedDriver("", ""),
	})
}
