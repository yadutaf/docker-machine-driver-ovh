package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	StatusTimeout = 200
)

// Driver is a machine driver for OVH.
type Driver struct {
	*drivers.BaseDriver

	// Command line parameters
	ProjectName string
	FlavorName  string
	RegionName  string

	// Internal ids
	ProjectId   string
	FlavorId    string
	ImageId     string
	InstanceId  string
	KeyPairName string
	KeyPairId   string

	// Overloaded credentials
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string

	// internal
	client *API
}

// GetCreateFlags registers the "machine create" flags recognized by this driver, including
// their help text and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "OVH_APPLICATION_KEY",
			Name:   "ovh-application-key",
			Usage:  "OVH API application key. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OVH_APPLICATION_SECRET",
			Name:   "ovh-application-secret",
			Usage:  "OVH API application secret. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OVH_APPLICATION_KEY",
			Name:   "ovh-consumer-key",
			Usage:  "OVH API consumer key. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-project",
			Usage: "OVH Cloud project name or id",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-region",
			Usage: "OVH Cloud region name",
			Value: DefaultRegionName,
		},
		mcnflag.StringFlag{
			Name:  "ovh-flavor",
			Usage: "OVH Cloud flavor name or id. Default: vps-ssd-1",
			Value: DefaultFlavorName,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "ovh"
}

// getClient returns an OVH API client
func (d *Driver) getClient() (api *API) {
	if d.client == nil {
		client, err := NewAPI("ovh-eu", d.ApplicationKey, d.ApplicationSecret, d.ConsumerKey)
		if err != nil {
			return nil
		}
		d.client = client
	}

	return d.client
}

// SetConfigFromFlags assigns and verifies the command-line arguments presented to the driver.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApplicationKey = flags.String("ovh-application-key")
	d.ApplicationSecret = flags.String("ovh-application-secret")
	d.ConsumerKey = flags.String("ovh-consumer-key")

	// Store configuration parameters as-is
	d.ProjectName = flags.String("ovh-project")
	d.RegionName = flags.String("ovh-region")
	d.FlavorName = flags.String("ovh-flavor")

	// Swarm configuration, must be in each driver
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	d.SSHUser = "admin"

	return nil
}

// PreCreateCheck does the network side validation
func (d *Driver) PreCreateCheck() error {
	client := d.getClient()

	// Validate project id
	log.Debug("Validating project")
	if d.ProjectName != "" {
		project, err := client.GetProjectByName(d.ProjectName)
		if err != nil {
			return err
		}
		d.ProjectId = project.Id
	} else {
		projects, err := client.GetProjects()
		if err != nil {
			return err
		}

		// If there is only one project, take it
		if len(projects) == 1 {
			d.ProjectId = projects[0]
		} else if len(projects) == 0 {
			return fmt.Errorf("No Cloud project could be found. To create a new one, please visit %s", CustomerInterface)
		} else {
			return fmt.Errorf("Multiple Cloud project found, to select one, use '--ovh-project' option")
		}
	}
	log.Debug("Found project id ", d.ProjectId)

	// Validate region
	log.Debug("Validating region")
	regions, err := client.GetRegions(d.ProjectId)
	if err != nil {
		return err
	}
	var ok bool
	for _, region := range regions {
		if region == d.RegionName {
			ok = true
			break
		}
	}
	if ok != true {
		return fmt.Errorf("Invalid region %s. For a list of valid ovh regions, please visis %s", CustomerInterface)
	}

	// Validate flavor
	log.Debug("Validating flavor")
	flavor, err := client.GetFlavorByName(d.ProjectId, d.RegionName, d.FlavorName)
	if err != nil {
		return err
	}
	d.FlavorId = flavor.Id
	log.Debug("Found flavor id ", d.FlavorId)

	// Validate image
	log.Debug("Validating image")
	image, err := client.GetImageByName(d.ProjectId, d.RegionName, ImageName)
	if err != nil {
		return err
	}
	d.ImageId = image.Id
	log.Debug("Found image id ", d.ImageId)

	// Create Key pair name
	d.KeyPairName = fmt.Sprintf("%s-%s", d.MachineName, mcnutils.GenerateRandomID())

	return nil
}

// createSSHKey creates an SSH key for the machine and uploads it
func (d *Driver) createSSHKey() error {
	log.WithField("Name", d.KeyPairName).Debug("Creating Key Pair...")

	// Generate key
	err := ssh.GenerateSSHKey(d.GetSSHKeyPath())
	if err != nil {
		return err
	}
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	// Upload key
	client := d.getClient()
	sshKey, err := client.CreateSshkey(d.ProjectId, d.KeyPairName, string(publicKey))
	if err != nil {
		return err
	}
	d.KeyPairId = sshKey.Id

	log.Debug("Created key id ", d.KeyPairId)
	return nil
}

// waitForInstanceStatus waits until instance reaches status. Copied from openstack Driver
func (d *Driver) waitForInstanceStatus(status string) (instance *Instance, err error) {
	return instance, mcnutils.WaitForSpecificOrError(func() (bool, error) {
		instance, err = d.client.GetInstance(d.ProjectId, d.InstanceId)
		if err != nil {
			return true, err
		}
		log.WithField("MachineId", d.MachineId).WithField("Status", instance.Status).Debug("Machine state")

		if instance.Status == "ERROR" {
			return true, fmt.Errorf("Instance creation failed. Instance is in ERROR state")
		}

		if instance.Status == status {
			return true, nil
		}

		return false, nil
	}, (StatusTimeout / 4), 4*time.Second)
}

// GetSSHHostname returns the hostname for SSH
func (d *Driver) GetSSHHostname() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) Create() error {
	client := d.getClient()

	// Create ssh key
	err := d.createSSHKey()
	if err != nil {
		return err
	}

	// Create instance
	log.Debug("Creating OVH instance...")
	instance, err := client.CreateInstance(
		d.ProjectId,
		d.MachineName,
		d.KeyPairId,
		d.FlavorId,
		d.ImageId,
		d.RegionName,
		false,
	)
	if err != nil {
		return err
	}
	d.InstanceId = instance.Id

	// Wait until instance is ACTIVE
	log.WithField("MachineId", d.InstanceId).Debug("Waiting for OVH instance...")
	instance, err = d.waitForInstanceStatus("ACTIVE")
	if err != nil {
		return err
	}

	// Save Ip address
	d.IPAddress = ""
	for _, ip := range instance.IpAddresses {
		if ip.Type == "public" {
			d.IPAddress = ip.Ip
			break
		}
	}

	if d.IPAddress == "" {
		return fmt.Errorf("No IP found for instance %s", instance.Id)
	}

	log.WithFields(log.Fields{
		"IP":        d.IPAddress,
		"MachineId": d.InstanceId,
	}).Debug("IP address found")

	// All done !
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// GetState return instance status
func (d *Driver) GetState() (state.State, error) {
	log.WithField("MachineId", d.InstanceId).Debug("Get status for OVH instance...")

	client := d.getClient()

	instance, err := client.GetInstance(d.ProjectId, d.InstanceId)
	if err != nil {
		return state.None, err
	}

	log.WithFields(log.Fields{
		"MachineId": d.InstanceId,
		"State":     instance.Status,
	}).Debug("State for OVH instance")

	switch instance.Status {
	case "ACTIVE":
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case "SUSPENDED":
		return state.Saved, nil
	case "SHUTOFF":
		return state.Stopped, nil
	case "BUILDING":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	}

	return state.None, nil
}

func (d *Driver) GetURL() (string, error) {
	if d.IPAddress == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(d.IPAddress, "2376")), nil
}

func (d *Driver) Remove() error {
	log.WithField("MachineId", d.InstanceId).Debug("deleting instance...")
	log.Info("Deleting OVH instance...")

	client := d.getClient()

	// Deletes instance
	err := client.DeleteInstance(d.ProjectId, d.InstanceId)
	if err != nil {
		return err
	}

	// Deletes ssh key
	log.WithField("KeyPairId", d.KeyPairId).Debug("deleting key pair...")
	err = client.DeleteSshkey(d.ProjectId, d.KeyPairId)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	log.WithField("MachineId", d.InstanceId).Info("Restarting OVH instance...")

	client := d.getClient()

	err := client.RebootInstance(d.ProjectId, d.InstanceId, false)
	if err != nil {
		return err
	}
	return nil
}

//
// STUBS
//
func (d *Driver) Kill() (err error) {
	return fmt.Errorf("Killing machines is not possible on OVH Cloud")
}
func (d *Driver) Start() (err error) {
	return fmt.Errorf("Starting machines is not possible on OVH Cloud")
}
func (d *Driver) Stop() (err error) {
	return fmt.Errorf("Stopping machines is not possible on OVH Cloud")
}
