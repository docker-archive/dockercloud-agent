package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ActiveState/tail"
	"github.com/docker/dockercloud-agent/utils"
)

type TunnelPatchForm struct {
	Tunnel  string `json:"tunnel"`
	Version string `json:"agent_version"`
}

type ReachableForm struct {
	Reachable bool `json:"reachable"`
}

func NatTunnel(url, ngrokPath, ngrokLogPath, ngrokConfPath, uuid string) {
	if isNodeReachable(url, uuid) {
		Logger.Printf("Node %s is publicly reachable", Conf.CertCommonName)
		return
	} else {
		Logger.Printf("Node %s is NOT publicly reachable", Conf.CertCommonName)
	}

	if !utils.FileExist(ngrokPath) {
		Logger.Println("Cannot find ngrok binary at", ngrokPath)
		DownloadNgrok(NgrokBinaryURL, ngrokPath)
	}

	updateNgrokHost(url)
	createNgrokConfFile(ngrokConfPath)

	var cmd *exec.Cmd

	if !utils.FileExist(ngrokConfPath) {
		SendError(errors.New("Cannot find ngrok conf"), "Cannot find ngrok conf file", nil)
		Logger.Println("Cannot find NAT tunnel configuration")
		return
	}
	cmd = exec.Command(ngrokPath,
		"-config", ngrokConfPath,
		"-log", "stdout",
		"-proto", "tcp",
		DockerHostPort)

	os.RemoveAll(ngrokLogPath)
	logFile, err := os.OpenFile(ngrokLogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		SendError(err, "Failed to open ngrok log file", nil)
		Logger.Println(err)
	} else {
		defer logFile.Close()
		cmd.Stdout = logFile
	}

	go monitorTunnels(url, ngrokLogPath)
	Logger.Println("Starting NAT tunnel")

	runNgrok(cmd)

	for {
		if ScheduledShutdown {
			Logger.Println("Scheduling for shutting down, do not restart the tunnel")
			break
		} else {
			Logger.Println("Restarting NAT tunnel in 10 seconds")
			time.Sleep(10 * time.Second)
			runNgrok(cmd)
		}
	}
}

func runNgrok(cmd *exec.Cmd) {
	if err := cmd.Start(); err != nil {
		SendError(err, "Failed to run NAT tunnel", nil)
		Logger.Println(err)
		return
	}
	cmd.Wait()
}

func monitorTunnels(url, ngrokLogPath string) {
	update, _ := tail.TailFile(ngrokLogPath, tail.Config{
		Follow: true,
		ReOpen: true})
	for line := range update.Lines {
		if strings.Contains(line.Text, "[INFO] [client] Tunnel established at") {
			terms := strings.Split(line.Text, " ")
			tunnel := terms[len(terms)-1]
			Logger.Printf("Found new tunnel: %s", tunnel)
			if tunnel != "" {
				patchTunnel(url, tunnel)
			}
		}
	}
}

func patchTunnel(url, tunnel string) {
	Logger.Println("Sending tunnel address to Docker Cloud")
	form := TunnelPatchForm{}
	form.Version = VERSION
	form.Tunnel = tunnel
	data, err := json.Marshal(form)
	if err != nil {
		SendError(err, "Json marshal error", nil)
		Logger.Printf("Cannot marshal the TunnelPatch form:%s\f", err)
	}

	headers := []string{"Authorization TutumAgentToken " + Conf.Token,
		"Content-Type", "application/json"}
	_, err = SendRequest("PATCH", utils.JoinURL(url, Conf.UUID), data, headers)
	if err != nil {
		SendError(err, "Failed to patch tunnel address to Docker Cloud", nil)
		Logger.Println("Failed to patch tunnel address to Docker Cloud,", err)
	}
	Logger.Println("New tunnel has been set up")
}

func DownloadNgrok(url, ngrokBinPath string) {
	if !utils.FileExist(ngrokBinPath) {
		Logger.Println("Downloading NAT tunnel binary ...")
		downloadFile(url, ngrokBinPath, "ngrok")
	}
}

func createNgrokConfFile(ngrokConfPath string) {
	ngrokConfStr := fmt.Sprintf("server_addr: %s\ntrust_host_root_certs: false\ninspect_addr: \"disabled\"", NgrokHost)
	if err := ioutil.WriteFile(ngrokConfPath, []byte(ngrokConfStr), 0666); err != nil {
		SendError(err, "Failed to create ngrok config file", nil)
		Logger.Println("Cannot create ngrok config file:", err)
	}
}

func updateNgrokHost(url string) {
	if NgrokHost != "" {
		return
	}

	headers := []string{"Authorization TutumAgentToken " + Conf.Token,
		"Content-Type application/json"}
	body, err := SendRequest("GET", utils.JoinURL(url, Conf.UUID), nil, headers)
	if err != nil {
		SendError(err, "SendRequest error", nil)
		Logger.Printf("Get registration info error, %s", err)
	} else {
		var form RegGetForm
		if err = json.Unmarshal(body, &form); err != nil {
			SendError(err, "Json unmarshal error", nil)
			Logger.Println("Cannot unmarshal the response", err)
		} else {
			if form.NgrokHost != "" {
				NgrokHost = form.NgrokHost
				Logger.Println("Ngrok server:", NgrokHost)
			}
		}
	}
}

func isNodeReachable(url, uuid string) bool {
	var reachableForm ReachableForm

	detailedUrl := utils.JoinURL(url, uuid+"/ping/")
	headers := []string{"Authorization TutumAgentToken " + Conf.Token,
		"Content-Type application/json",
		"User-Agent dockercloud-agent/" + VERSION}

	//waiting for docker port opens
	Logger.Print("Waiting for docker unix socket to be ready")
	for {
		unixsock := DockerDefaultHost
		if strings.HasPrefix(unixsock, "unix://") {
			unixsock = DockerDefaultHost[7:]
		}

		_, err := net.DialTimeout("unix", unixsock, DialTimeOut*time.Second)
		if err == nil {
			break
		} else {
			time.Sleep(2 * time.Second)
		}
	}
	Logger.Print("Docker unix socket opened")

	//check the port from remote server
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		body, err := SendRequest("POST", detailedUrl, nil, headers)
		if err == nil {
			if err := json.Unmarshal(body, &reachableForm); err != nil {
				SendError(err, "Json unmarshal error", nil)
				Logger.Println("Failed to unmarshal the response", err)
			} else {
				return reachableForm.Reachable
			}
		}
		SendError(err, "Node reachable check HTTP error", nil)
		Logger.Printf("Node reachable check failed, %s. Retry in %d seconds", err, i)
		time.Sleep(time.Duration(i) * time.Second)
	}
}
