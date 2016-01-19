package agent

import (
	"log"
	"os"
)

var (
	FlagDebugMode     *bool
	FlagLogToStdout   *bool
	FlagStandalone    *bool
	FlagSkipNatTunnel *bool
	FlagDockerHost    *string
	FlagDockerOpts    *string
	FlagHost          *string
	FlagToken         *string
	FlagUUID          *string
	FlagNgrokHost     *string
	FlagVersion       *bool

	Conf                      Configuration
	Logger                    *log.Logger
	DockerProcess             *os.Process
	ScheduleToTerminateDocker = false
	ScheduledShutdown         = false
	DockerBinaryURL           = ""
	NgrokBinaryURL            = ""
	NgrokHost                 = ""
	DockerClientVersion       = ""
)

const (
	VERSION               = "1.0.0"
	defaultCertCommonName = ""
	defaultDockerHost     = "tcp://0.0.0.0:2375"
	defaultAgentHost      = "https://cloud.docker.com/"
)

const (
	AgentHome = "/etc/dockercloud/agent"
	DockerDir = "/usr/lib/dockercloud"
	LogDir    = "/var/log/dockercloud"

	DockerSymbolicLink     = "/usr/bin/docker"
	DockerLogFileName      = "docker.log"
	AgentLogFileName       = "agent.log"
	KeyFileName            = "key.pem"
	CertFileName           = "cert.pem"
	CAFileName             = "ca.pem"
	ConfigFileName         = "dockercloud-agent.conf"
	DockerBinaryName       = "docker"
	DockerNewBinaryName    = "docker.new"
	DockerNewBinarySigName = "docker.new.sig"
	NgrokBinaryName        = "ngrok"
	NgrokLogName           = "ngrok.log"
	NgrokConfName          = "ngrok.conf"
	AgentPidFile           = "/var/run/dockercloud-agent.pid"

	RegEndpoint       = "api/agent/v1/node/"
	DockerDefaultHost = "unix:///var/run/docker.sock"

	MaxWaitingTime    = 200 //seconds
	HeartBeatInterval = 5   //seconds

	RenicePriority  = -10
	ReniceSleepTime = 5 //seconds

	DockerHostPort = "2375"

	DialTimeOut = 10 //seconds
)
