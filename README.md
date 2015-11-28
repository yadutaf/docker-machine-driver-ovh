[![OVH Cloud driver for Docker-Machine](http://yadutaf.github.io/docker-machine-driver-ovh/img/logo.png)](https://github.com/yadutaf/docker-machine-driver-ovh)

```bash
# Create your first /cloud Docker vps
docker-machine create -d ovh hello-docker
```

OVH Cloud Driver is based on openstack driver. To use it, you need a cloud
project. If you don't have one yet, you may create one from [your OVH control
panel](https://www.ovh.com/manager/cloud/index.html) at any time.

## Installation

The easiest way to install ovh docker-machine driver is to:

```bash
go install github.com/yadutaf/docker-machine-driver-ovh
```

## Example Usage

### 1. Get OpenStack credentials

1. Login to your /Cloud control panel: https://www.ovh.com/manager/cloud/index.html
2. Select "Project management and resource usage", from the top right corner
3. Select the OpenStack tab
4. If you need one, select "Add user"
5. Click on wrench icon, on the right
6. Select "Download an Openstack configuration file"

### 2. Source it

```bash
source openrc.sh
# And type your OpenStack account password (You may regenerate it at any time from the console)
```

### 3. Create your machines!

```bash
docker-machine create -d ovh node-1
```

## Options

|Option Name|Description|Default Value|required|
|---|---|---|---|
|``--ovh-username`` or ``$OS_USERNAME``         |OpenStack user name               |none     |yes|
|``--ovh-password`` or ``$OS_PASSWORD``         |OpenStack user password           |none     |yes|
|``--ovh-tenant-name`` or ``$OS_TENANT_NAME``   |OpenStack project name            |none     |yes|
|``--ovh-tenant-id`` or ``$OS_TENANT_ID``       |OpenStack project id              |none     |yes|
|``--ovh-region`` or ``OS_REGION_NAME``         |OpenStack region                  |GRA1     |no|
|``--ovh-flavor`` or ``OS_FLAVOR_NAME``         |Machine type                      |vps-ssd-1|no|
|``--ovh-sec-groups`` or ``$OS_SECURITY_GROUPS``|Comma separated list of sec-groups|default  |no|

Note: All "OpenStack" parameters come from ``openrc.sh``. It is highly
recommended to use this file to simplify docker-machine creation.

Each environment variable may be orverloaded by it option equivalent at run time.

## Hacking

### Get the sources

```bash
go get github.com/yadutaf/docker-machine-driver-ovh
cd $GOPATH/src/github.com/yadutaf/docker-machine-driver-ovh
```

### Test the driver

To test the driver will develping on it, make sure your current
build directory has the highest priority in your ``$PATH`` so that
docker-machine can find it.

```
export PATH=$GOPATH/src/github.com/yadutaf/docker-machine-driver-ovh:$PATH
```

## Related links

- **OVH Cloud console**: https://www.ovh.com/manager/cloud/index.html
- **OVH Cloud offers**: https://www.ovh.com/us/cloud/
- **Contribute**: https://github.com/yadutaf/docker-machine-driver-ovh
- **Report bugs**: https://github.com/yadutaf/docker-machine-driver-ovh

## License

MIT
