package main // import "github.com/docker/dockercloud-agent"

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"syscall"
	"time"

	. "github.com/docker/dockercloud-agent/agent"
	"github.com/docker/dockercloud-agent/utils"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func main() {
	dockerBinPath := path.Join(DockerHome, DockerBinaryName)
	dockerTarPath := path.Join(AgentLibHome, DockerTarName)
	dockerTarSigPath := path.Join(AgentLibHome, DockerTarSigName)
	configFilePath := path.Join(AgentHome, ConfigFileName)
	keyFilePath := path.Join(AgentHome, KeyFileName)
	certFilePath := path.Join(AgentHome, CertFileName)
	caFilePath := path.Join(AgentHome, CAFileName)
	ngrokLogPath := path.Join(LogDir, NgrokLogName)
	ngrokConfPath := path.Join(AgentHome, NgrokConfName)

	_ = os.MkdirAll(AgentHome, 0755)
	_ = os.MkdirAll(AgentLibHome, 0755)
	_ = os.MkdirAll(LogDir, 0755)

	ParseFlag()

	if *FlagVersion {
		fmt.Println(VERSION)
		return
	}
	SetLogger(path.Join(LogDir, AgentLogFileName))
	Logger.Print("Running dockercloud-agent: version ", VERSION)
	CreatePidFile(AgentPidFile)

	PrepareFiles(configFilePath, dockerBinPath, keyFilePath, certFilePath)
	SetConfigFile(configFilePath)

	regUrl := utils.JoinURL(Conf.Host, RegEndpoint)
	if Conf.UUID == "" {
		os.RemoveAll(keyFilePath)
		os.RemoveAll(certFilePath)
		os.RemoveAll(caFilePath)

		if !*FlagStandalone {
			Logger.Printf("Registering in Docker Cloud via POST: %s", regUrl)
			RegPost(regUrl, caFilePath, configFilePath)
		}
	}

	if *FlagStandalone {
		commonName := Conf.CertCommonName
		if commonName == "" {
			commonName = "*"
		}
		CreateCerts(keyFilePath, certFilePath, commonName)
	} else {
		CreateCerts(keyFilePath, certFilePath, Conf.CertCommonName)
	}

	if !*FlagStandalone {
		Logger.Printf("Registering in Docker Cloud via PATCH: %s",
			regUrl+Conf.UUID)
		err := RegPatch(regUrl, caFilePath, certFilePath, configFilePath)
		if err != nil {
			Logger.Printf("PATCH error %s :either UUID (%s) or Token is invalid", err.Error(), Conf.UUID)
			Conf.UUID = ""
			SaveConf(configFilePath, Conf)

			os.RemoveAll(keyFilePath)
			os.RemoveAll(certFilePath)
			os.RemoveAll(caFilePath)

			Logger.Printf("Registering in Docker Cloud via POST: %s", regUrl)
			RegPost(regUrl, caFilePath, configFilePath)

			CreateCerts(keyFilePath, certFilePath, Conf.CertCommonName)
			DownloadDocker(DockerTarURL, DockerHome)

			Logger.Printf("Registering in Docker Cloud via PATCH: %s",
				regUrl+Conf.UUID)
			if err = RegPatch(regUrl, caFilePath, certFilePath, configFilePath); err != nil {
				SendError(err, "Registion HTTP error", nil)
			}
		}
	}

	if err := SaveConf(configFilePath, Conf); err != nil {
		SendError(err, "Failed to save config to the conf file", nil)
		Logger.Fatalln(err)
	}

	DownloadDocker(DockerTarURL, DockerHome)
	Logger.Print("Found docker: version ", GetDockerClientVersion(dockerBinPath))

	HandleSig()
	syscall.Setpriority(syscall.PRIO_PROCESS, os.Getpid(), RenicePriority)

	Logger.Println("Initializing docker daemon")
	StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)

	if !*FlagStandalone {
		if *FlagSkipNatTunnel {
			Logger.Println("Skip NAT tunnel")
		} else {
			Logger.Println("Loading NAT tunnel module")
			go NatTunnel(regUrl, AgentLibHome, ngrokLogPath, ngrokConfPath, Conf.UUID)
		}
	} else {
		DownloadNgrok(NgrokTarURL, AgentLibHome)
	}

	if !*FlagStandalone {
		Logger.Println("Verifying the registration with Docker Cloud")
		go VerifyRegistration(regUrl)
	}

	Logger.Println("Docker server started. Entering maintenance loop")
	for {
		time.Sleep(HeartBeatInterval * time.Second)
		UpdateDocker(DockerHome, dockerTarPath, dockerTarSigPath, keyFilePath, certFilePath, caFilePath)

		// try to restart docker daemon if it dies somehow
		if DockerProcess == nil {
			time.Sleep(HeartBeatInterval * time.Second)
			if DockerProcess == nil && ScheduleToTerminateDocker == false {
				Logger.Println("Respawning docker daemon")
				StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)
			}
		}
	}
}

func PrepareFiles(configFilePath, dockerBinPath, keyFilePath, certFilePath string) {
	Logger.Println("Checking if config file exists")
	if !utils.FileExist(configFilePath) {
		LoadDefaultConf()
		if err := SaveConf(configFilePath, Conf); err != nil {
			SendError(err, "Failed to save config to the conf file", nil)
			Logger.Fatalln(err)
		}
	}

	Logger.Println("Loading Configuration file")
	conf, err := LoadConf(configFilePath)
	if err != nil {
		SendError(err, "Failed to load configuration file", nil)
		Logger.Fatalln("Failed to load configuration file:", err)
	} else {
		Conf = *conf
	}

	if *FlagDockerHost != "" {
		Logger.Printf("Override 'DockerHost' from command line flag: %s\n", *FlagDockerHost)
		Conf.DockerHost = *FlagDockerHost
	}
	if *FlagHost != "" {
		Logger.Printf("Override 'Host' from command line flag: %s\n", *FlagHost)
		Conf.Host = *FlagHost
	}
	if *FlagToken != "" {
		Logger.Printf("Override 'Token' from command line flag: %s\n", *FlagToken)
		Conf.Token = *FlagToken
	}
	if *FlagUUID != "" {
		Logger.Printf("Override 'UUID' from command line flag: %s\n", *FlagUUID)
		Conf.UUID = *FlagUUID
	}
	if *FlagDockerOpts != "" {
		Logger.Printf("Override 'DockerOpts' from command line flag: %s\n", *FlagDockerOpts)
		Conf.DockerOpts = *FlagDockerOpts
	}
}
