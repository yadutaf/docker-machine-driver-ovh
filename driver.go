package main

import (
	"fmt"
	"strings"

	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
)

// Driver is a machine driver for OVH. It's a specialization of the generic OpenStack one.
type Driver struct {
	*openstack.Driver
}

// GetCreateFlags registers the "machine create" flags recognized by this driver, including
// their help text and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "OS_USERNAME",
			Name:   "ovh-username",
			Usage:  "OVH Cloud username",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PASSWORD",
			Name:   "ovh-password",
			Usage:  "OVH Cloud password",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_TENANT_NAME",
			Name:   "ovh-tenant-name",
			Usage:  "OVH Cloud project name",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_TENANT_ID",
			Name:   "ovh-tenant-id",
			Usage:  "OVH Cloud project id",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_REGION_NAME",
			Name:   "ovh-region",
			Usage:  "OVH Cloud region name",
			Value:  DefaultRegionName,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLAVOR_NAME",
			Name:   "ovh-flavor",
			Usage:  "OVH Cloud flavor name. Default: VPS SSD 2GB",
			Value:  DefaultFlavorName,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_SECURITY_GROUPS",
			Name:   "ovh-sec-groups",
			Usage:  "OVH Cloud comma separated security groups for the machine",
			Value:  DefaultSecurityGroup,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "ovh"
}

func missingEnvOrOption(setting, envVar, opt string) error {
	return fmt.Errorf(
		"%s must be specified either using the environment variable %s or the CLI option %s",
		setting,
		envVar,
		opt,
	)
}

// SetConfigFromFlags assigns and verifies the command-line arguments presented to the driver.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthUrl = AuthURL
	d.Username = flags.String("ovh-username")
	d.Password = flags.String("ovh-password")
	d.TenantId = flags.String("ovh-tenant-id")
	d.TenantName = flags.String("ovh-tenant-name")

	d.Region = flags.String("ovh-region")
	d.FlavorName = flags.String("ovh-flavor")
	d.ImageName = ImageName
	d.SSHUser = SshUserName

	d.IpVersion = 4
	d.NetworkName = NetworkName
	if flags.String("ovh-sec-groups") != "" {
		d.SecurityGroups = strings.Split(flags.String("ovh-sec-groups"), ",")
	}

	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	if d.Username == "" {
		return missingEnvOrOption("Username", "OS_USERNAME", "--ovh-username")
	}
	if d.Username == "" {
		return missingEnvOrOption("Password", "OS_PASSWORD", "--ovh-password")
	}
	if d.Username == "" {
		return missingEnvOrOption("Project name", "OS_TENANT_NAME", "--ovh-tenant_name")
	}
	if d.Username == "" {
		return missingEnvOrOption("Project ID", "OS_TENANT_ID", "--ovh-tenant-id")
	}

	return nil
}
