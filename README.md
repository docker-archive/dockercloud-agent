dockercloud-agent
===========


## What's this?

This is the agent Docker Cloud uses to set up nodes. It's a daemon that will register the host with the DockerCloud API using a user token (`Token`), and will manage the installation, configuration and ongoing upgrade of the Docker daemon.

For information on how to install it in your host, please check the [Bring Your Own Node](https://docs.docker.com/docker-cloud/infrastructure/byoh/) documentation.


## Running

If installing from a `.deb` or `.rpm` package, Docker Cloud Agent will be configured in upstart to be launched on boot.

```
# dockercloud-agent -h
Usage of ./dockercloud-agent:
  -debug
    	Enable debug mode
  -docker-host string
    	Override 'DockerHost'
  -docker-opts string
    	Add additional flags to run docker daemon
  -host string
    	Override 'Host' in the configuration file
  -ngrok-host string
    	ngrok host for NAT tunneling
  -skip-nat-tunnel
    	Skip NAT tunnel
  -standalone
    	Standalone mode, skipping reg with Docker Cloud
  -stdout
    	Print log to stdout
  -token string
    	Override 'Token' in the configuration file
  -uuid string
    	Override 'UUID'  in the configuration file
  -v	show version
   set: Set items in the config file and exit, supported items
          CertCommonName="xxx"
          DockerHost="xxx"
          Host="xxx"
          Token="xxx"
          UUID="xxx"
          DockerOpts="xxx"

```


Configuration file is located in `/etc/dockercloud/agent/dockercloud-agent.conf` (JSON file) with the following structure:

```
{
	"CertCommonName":"",
	"DockerHost":"tcp://0.0.0.0:2375",
	"Host":"https://cloud.docker.com/",
	"Token":"",
	"UUID":"",
	"DockerOpts":""
}
```

## Logging

Logs are stored under `/var/log/dockercloud/`:

* `agent.log` contains the logs of the agent itself
* `docker.log` contains the Docker daemon logs


## Building

Run `make` to build binaries and `.deb` and `.rpm` packages which will be stored in the `build/` folder.

# Proxy

If `HTTP_PROXY` and `HTTPS_PRXOY` is defined, cloud-agent will read and use them. (this is supported by golang natively)

## Supported Distributions

Currently supported and tested on:

- Ubuntu 14.04, 15.04, 15.10
- CentOS 7
- Fedora 21, 22
- Debian 8
- Red Hat Enterprise Linux 7


## Reporting security issues

To report a security issue, please send us an email to [security@docker.com](mailto:security@docker.com). Thank you!
