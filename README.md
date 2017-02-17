[![OVH Cloud driver for Docker-Machine](https://raw.githubusercontent.com/yadutaf/docker-machine-driver-ovh/master/img/logo.png)](https://github.com/yadutaf/docker-machine-driver-ovh)

```bash
# Create your first OVH Cloud Docker machine
docker-machine create -d ovh hello-docker
```

OVH Cloud driver is based on OVH Public API. To create OVH
API credentials, you may visit https://api.ovh.com. See
[1. Get your OVH credentials](#1-get-your-ovh-credentials) below for a detailed how-to

## Installation

The easiest way to install ovh docker-machine driver is to:

```bash
go get github.com/yadutaf/docker-machine-driver-ovh
go install github.com/yadutaf/docker-machine-driver-ovh
ln -s $GOPATH/bin/docker-machine-driver-ovh /usr/local/bin/docker-machine-driver-ovh
```

## Example Usage

### 1. Get your OVH credentials

OVH driver never store your OVH login and password. Instead it uses a set of
3 keys dedicated to a given application. If you don't have your keys already,
generating them is easy:

1. Go to the token creation page: https://api.ovh.com/createToken/?GET=/*&POST=/*&DELETE=/*&PUT=/*
2. Enter your login, password, a name and a short description then validate. You may want to increase the validity period.
3. You now have an ``Application Key``, ``Application Secret`` and a ``Consumer Key``.

## 2. Create a configuration file

Create a file named ```ovh.conf```.
This file will allow any application built with official OVH API drivers to
use it without requiring new keys:

```ini
; ovh.conf
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=<Application Key>
application_secret=<Application Secret>
consumer_key=<Consumer Key>
```

### 3. Create your machines!

*Basic example*:

```bash
docker-machine create -d ovh node-1
```

*Advanced example: Use CoreOS with VPS-SSD-2 in Data Center Strasbourg 1*

```bash
docker-machine -D create --ovh-region "SBG1" --ovh-flavor "vps-ssd-2" --ovh-image "CoreOS stable 899.15.0" --ovh-ssh-user "core" --driver ovh node-1
```
Note: For the different image-types you have to use special --ovh-ssh-user (for Example "ubuntu" for Ubuntu OS, "core" for CoreOS and "admin" for Debian)

## Configuration

### Options

|Option Name|Description|Default Value|required|
|---|---|---|---|
|``--ovh-application-secret`` or ``$OVH_APPLICATION_SECRET``|Application Secret|none      |yes|
|``--ovh-application-key`` or ``$OVH_APPLICATION_KEY``      |Application key   |none      |yes|
|``--ovh-consumer-key`` or ``$OVH_CONSUMER_KEY``            |Consumer Key      |none      |yes|
|``--ovh-endpoint`` or ``$OVH_ENDPOINT``                    |Endpoint          |none      |no|
|``--ovh-region``                                           |Cloud region      |GRA1      |no|
|``--ovh-private-network``                                  |Cloud private network |public |no|
|``--ovh-flavor``                                           |Cloud Machine type|vps-ssd-1 |no|
|``--ovh-image``                                            |Cloud Machine image|Ubuntu 16.04 |no|
|``--ovh-ssh-user``                                         |Cloud Machine SSH User|ubuntu |no|
|``--ovh-project``                                          |Cloud Project name/description or id|single one|only if multiple projects|
|``--ovh-ssh-key``                                          |Cloud Machine SSH Key|none |no|
|``--ovh-billing-period``                                   |OVH Cloud billing period (hourly or monthly)|hourly |no|

### Vrack integration

The vRack is [OVH's private networks](https://www.ovh.com/us/solutions/vrack/). A vRack may contain up to 4000 Vlans and any compatible OVH products, including Cloud projects.
To use the vRack with this docker-machine driver:

1. [Create a new vRack](https://www.ovh.com/manager/cloud/index.html#/vrack/new) if needed (this is free)
2. [Add your Cloud project to the vRack](https://www.ovh.com/manager/cloud/index.html#/vrack)
3. [Go back to your Cloud project](https://www.ovh.com/manager/cloud/index.html#/iaas/pci/project)
4. And configure some Vlan

Once this setup is done, you are ready to use the vRack with Docker Machine:

```
VLAN_NUMBER=3
docker-machine create -d ovh --ovh-private-network $VLAN_NUMBER machine-in-the-vrack
```

Note that you will still need to configure the interface. A quick way to do it is:

```
docker-machine ssh machine-in-the-vrack

# Create the configuration file
sudo tee /etc/network/interfaces.d/99-vrack.cfg << EOF
auto ens4
iface ens4 inet dhcp
EOF

# Enable the interface
sudo ifup ens4
```

### Authentication

OVH credentials may be supplied through arguments, environment or configuration file, by order of decreasing priority. The configuration may be:

- global ``/etc/ovh.conf``
- user specific ``~/.ovh.conf``
- application specific ``./ovh.conf``

### SSH Key

Docker-machine can generate a key for each new machine. It is a nice feature to start with but it will quickly load your OVH project with many keys (even though these keys are removed uppon machine deletion).

With the `--ovh-ssh-key` option you can define a key name (already present in your ovh project). This key must be accessible (in ~/.ssh or in the ssh agent) by the ssh binary present on the machine running docker-mamchine.

## Hacking

### Get the sources

```bash
go get github.com/yadutaf/docker-machine-driver-ovh
cd $GOPATH/src/github.com/yadutaf/docker-machine-driver-ovh
```

### Build the driver
Make sure to export `GO15VENDOREXPERIMENT=1`. Every necessary dependency is
stored in the vendoring directory.

### Test the driver

To test the driver make sure your current build directory has the highest
priority in your ``$PATH`` so that docker-machine can find it.

```
export PATH=$GOPATH/src/github.com/yadutaf/docker-machine-driver-ovh:$PATH
```

## Related links

- **OVH Cloud console**: https://www.ovh.com/manager/cloud/index.html
- **OVH Cloud offers**: https://www.ovh.com/us/cloud/
- **Docker Machine**: https://docs.docker.com/machine/
- **Contribute**: https://github.com/yadutaf/docker-machine-driver-ovh
- **Report bugs**: https://github.com/yadutaf/docker-machine-driver-ovh/issues

## License

MIT
