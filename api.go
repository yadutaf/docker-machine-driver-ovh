package main

import (
	"encoding/json"
	"fmt"
	"github.com/yadutaf/go-ovh"
)

const (
	CustomerInterface = "https://www.ovh.com/manager/cloud/index.html"
)

type API struct {
	client *ovh.Client
}

type Project struct {
	Name         string `json:"description"`
	Id           string `json:"project_id"`
	Unleash      bool   `json:"unleash"`
	CreationDate string `json:"creationDate"`
	OrderId      int    `json:"orderId"`
	Status       string `json:"status"`
}
type Projects []string

type Flavor struct {
	Region      string `json:"region"`
	Name        string `json:"name"`
	Id          string `json:"id"`
	OS          string `json:"osType"`
	Vcpus       int    `json:"vcpus"`
	MemoryGB    int    `json:"ram"`
	DiskSpaceGB int    `json:"disk"`
	Type        string `json:"type"`
}
type Flavors []Flavor

type Image struct {
	Region       string `json:"region"`
	Name         string `json:"name"`
	Id           string `json:"id"`
	OS           string `json:"type"`
	CreationDate string `json:"creationDate"`
	Status       string `json:"status"`
	MinDisk      int    `json:"minDisk"`
	Visibility   string `json:"visibility"`
}
type Images []Image

type Regions []string

type SshkeyReq struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	Region    string `json:"region,omitempty"`
}

type Sshkey struct {
	Name        string  `json:"name"`
	Id          string  `json:"id"`
	PublicKey   string  `json:"publicKey"`
	Fingerprint string  `json:"fingerPrint"`
	Regions     Regions `json:"region"`
}

type Ip struct {
	Ip   string `json:"ip"`
	Type string `json:"type"`
}
type Ips []Ip

type InstanceReq struct {
	Name           string `json:"name"`
	FlavorId       string `json:"flavorId"`
	ImageId        string `json:"imageId"`
	Region         string `json:"region"`
	SshkeyId       string `json:"sshKeyId"`
	MonthlyBilling bool   `json:"monthlyBilling"`
}

type Instance struct {
	Name           string `json:"name"`
	Id             string `json:"id"`
	Status         string `json:"status"`
	Created        string `json:"created"`
	Region         string `json:"region"`
	Image          Image  `json:"image"`
	Flavor         Flavor `json:"flavor"`
	Sshkey         Sshkey `json:"sshKey"`
	IpAddresses    Ips    `json:"ipAddresses"`
	MonthlyBilling bool   `json:"monthlyBilling"`
}

type RebootReq struct {
	Type string `json:"type"`
}

func NewAPI(endpoint, applicationKey, applicationSecret, consumerKey string) (api *API, err error) {
	client, err := ovh.NewClient(endpoint, applicationKey, applicationSecret, consumerKey)
	if err != nil {
		return nil, err
	}

	return &API{client}, nil
}

// GetProjects returns a list of string project Id
func (a *API) GetProjects() (projects Projects, err error) {
	res, err := a.client.Get("/cloud/project")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject return the details of a project given a project id
func (a *API) GetProject(projectId string) (project *Project, err error) {
	res, err := a.client.Get("/cloud/project/" + projectId)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectByName returns the details of a project given its name. This is slower than GetProject
func (a *API) GetProjectByName(projectName string) (project *Project, err error) {
	// get project list
	projects, err := a.GetProjects()
	if err != nil {
		return nil, err
	}

	// If projectName is a valid projectId return it.
	for _, projectId := range projects {
		if projectId == projectName {
			return a.GetProject(projectId)
		}
	}

	// Attempt to find a project matching projectName. This is potentially slow
	for _, projectId := range projects {
		project, err := a.GetProject(projectId)
		if err != nil {
			return nil, err
		}

		if project.Name == projectName {
			return project, nil
		}
	}

	// Ooops
	return nil, fmt.Errorf("Project '%s' does not exist on OVH cloud. To create or rename a project, please visit %s", projectName, CustomerInterface)
}

// GetRegions returns the list of valid regions for a given project
func (a *API) GetRegions(projectId string) (regions Regions, err error) {
	url := fmt.Sprintf("/cloud/project/%s/region", projectId)
	res, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &regions)
	if err != nil {
		return nil, err
	}

	return regions, nil
}

// GetFlavors returns the list of available flavors for a given project in a giver zone
func (a *API) GetFlavors(projectId, region string) (flavors Flavors, err error) {
	url := fmt.Sprintf("/cloud/project/%s/flavor?region=%s", projectId, region)
	res, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &flavors)
	if err != nil {
		return nil, err
	}

	return flavors, nil
}

// GetFlavorByName returns the details of a flavor given its name. Slower than getting by id
func (a *API) GetFlavorByName(projectId, region, flavorName string) (flavor *Flavor, err error) {
	// Get flavor list
	flavors, err := a.GetFlavors(projectId, region)
	if err != nil {
		return nil, err
	}

	// Find first matching Linux flavor
	for _, flavor := range flavors {
		if flavor.OS != "linux" {
			continue
		}

		if flavor.Id == flavorName || flavor.Name == flavorName {
			return &flavor, nil
		}
	}

	// Ooops
	return nil, fmt.Errorf("Flavor '%s' does not exist on OVH cloud. To find a list of available flavors, please visit %s", flavorName, CustomerInterface)
}

// GetImages returns a list of images for a given project in a given region
func (a *API) GetImages(projectId, region string) (images Images, err error) {
	url := fmt.Sprintf("/cloud/project/%s/image?osType=linux&region=%s", projectId, region)
	res, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &images)
	if err != nil {
		return nil, err
	}

	return images, nil
}

// GetImageByName returns the details of an image given its name, a project and a region. This is slower than id access
func (a *API) GetImageByName(projectId, region, imageName string) (image *Image, err error) {
	// Get image list
	images, err := a.GetImages(projectId, region)
	if err != nil {
		return nil, err
	}

	// Find first matching image
	for _, image := range images {
		if image.OS != "linux" {
			continue
		}

		if image.Id == imageName || image.Name == imageName {
			return &image, nil
		}
	}

	// Ooops
	return nil, fmt.Errorf("Image '%s' does not exist on OVH cloud. To find a list of available images, please visit %s", imageName, CustomerInterface)
}

// CreateSshkey uploads a new public key with name and returns resulting object
func (a *API) CreateSshkey(projectId, name, pubkey string) (sshkey *Sshkey, err error) {
	var sshkeyreq SshkeyReq
	sshkeyreq.Name = name
	sshkeyreq.PublicKey = pubkey

	url := fmt.Sprintf("/cloud/project/%s/sshkey", projectId)
	res, err := a.client.Post(url, sshkeyreq)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &sshkey)
	if err != nil {
		return nil, err
	}

	return sshkey, nil
}

// DeleteSshKey deletes an existing sshkey
func (a *API) DeleteSshkey(projectId, instanceId string) (err error) {
	url := fmt.Sprintf("/cloud/project/%s/sshkey/%s", projectId, instanceId)
	_, err = a.client.Delete(url)
	return err
}

// CreateInstance start a new public cloud instance and returns resulting object
func (a *API) CreateInstance(projectId, name, pubkeyId, flavorId, ImageId, region string, monthlyBilling bool) (instance *Instance, err error) {
	var instanceReq InstanceReq
	instanceReq.Name = name
	instanceReq.SshkeyId = pubkeyId
	instanceReq.FlavorId = flavorId
	instanceReq.ImageId = ImageId
	instanceReq.Region = region
	instanceReq.MonthlyBilling = monthlyBilling

	url := fmt.Sprintf("/cloud/project/%s/instance", projectId)
	res, err := a.client.Post(url, instanceReq)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// RebootInstance reboot an instance
func (a *API) RebootInstance(projectId, instanceId string, hard bool) (err error) {
	var rebootReq RebootReq
	if hard == true {
		rebootReq.Type = "hard"
	} else {
		rebootReq.Type = "soft"
	}

	url := fmt.Sprintf("/cloud/project/%s/instance/%s", projectId, instanceId)
	_, err = a.client.Post(url, rebootReq)

	return err
}

// DeleteInstance stops and destroys a public cloud instance
func (a *API) DeleteInstance(projectId, instanceId string) (err error) {
	url := fmt.Sprintf("/cloud/project/%s/instance/%s", projectId, instanceId)
	_, err = a.client.Delete(url)
	return err
}

func (a *API) GetInstance(projectId, instanceId string) (instance *Instance, err error) {
	url := fmt.Sprintf("/cloud/project/%s/instance/%s", projectId, instanceId)
	res, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res.Body, &instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
