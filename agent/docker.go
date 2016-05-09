package agent

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/dockercloud-agent/utils"
	"github.com/flynn-archive/go-shlex"
)

func DownloadDocker(url, dockerHome string) {
	if !utils.FileExist(path.Join(dockerHome, DockerBinaryName)) {
		Logger.Println("Downloading docker binary...")
		downloadFile(url, dockerHome, "docker")
	}
}

func GetDockerClientVersion(dockerBinPath string) (version string) {
	var versionStr string
	out, err := exec.Command(dockerBinPath, "-v").Output()
	if err != nil {
		SendError(err, "Failed to get the docker version", nil)
	}
	versionStr = string(out)

	if versionStr != "" {
		re := regexp.MustCompile("\\d+\\.\\d+\\.\\d+[a-zA-Z0-9_\\-]*")
		match := re.FindStringSubmatch(versionStr)
		if match != nil && len(match) > 0 {
			version = match[0]
		}
	}
	return
}

func getDockerStartOpt(dockerBinPath, keyFilePath, certFilePath, caFilePath string) []string {
	daemonOpt := "daemon"
	userlandProxyOpt := " --userland-proxy=false"

	debugOpt := ""
	if *FlagDebugMode {
		debugOpt = " -D"
	}

	bindOpt := fmt.Sprintf(" -H %s -H %s", DockerDefaultHost, Conf.DockerHost)

	var certOpt string
	if *FlagStandalone && !utils.FileExist(caFilePath) {
		certOpt = fmt.Sprintf(" --tlscert %s --tlskey %s --tls", certFilePath, keyFilePath)
		fmt.Fprintln(os.Stderr, "WARNING: standalone mode activated but no CA certificate found - client authentication disabled")
	} else {
		certOpt = fmt.Sprintf(" --tlscert %s --tlskey %s --tlscacert %s --tlsverify", certFilePath, keyFilePath, caFilePath)
	}

	extraOpt := ""
	if Conf.DockerOpts != "" {
		extraOpt = " " + Conf.DockerOpts
	}

	optStr := fmt.Sprintf("%s%s%s%s%s%s", daemonOpt, debugOpt, bindOpt, userlandProxyOpt, certOpt, extraOpt)

	optSlice, err := shlex.Split(optStr)
	if err != nil {
		optSlice = strings.Split(optStr, " ")
	}
	return optSlice
}

func StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath string) {
	optSlice := getDockerStartOpt(dockerBinPath, keyFilePath, certFilePath, caFilePath)
	command := exec.Command(dockerBinPath, optSlice...)
	go runDocker(command)
}

func StopDocker() {
	if DockerProcess != nil {
		DockerProcess.Signal(syscall.SIGTERM)
		for {
			if DockerProcess != nil {
				time.Sleep(500 * time.Millisecond)
			} else {
				break
			}
		}
	}
}

func UpdateDocker(dockerHome, dockerTarPath, dockerTarSigPath, keyFilePath, certFilePath, caFilePath string) {
	dockerBinPath := path.Join(dockerHome, DockerBinaryName)
	defer func() {
		if err := recover(); err != nil {
			Logger.Println("Cannot uncomporess the tar file. The update is rejected")
			removeUpdateFiles(dockerTarPath, dockerTarSigPath)
			ScheduleToTerminateDocker = false
			StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)
		}
	}()
	if utils.FileExist(dockerTarPath) {
		Logger.Printf("New docker update (%s) found", dockerTarPath)
		Logger.Println("Updating docker...")
		if verifyDockerSig(dockerTarPath, dockerTarSigPath) {
			Logger.Println("Stopping docker daemon")
			ScheduleToTerminateDocker = true
			StopDocker()

			Logger.Println("Applying new docker update")
			tarfile, err := ioutil.ReadFile(dockerTarPath)
			if err != nil {
				SendError(err, "Failed to read the docker update file", nil)
				Logger.Println("Failed read the docker update file:", err)
			}
			uncompress(tarfile, dockerHome)
			removeUpdateFiles(dockerTarPath, dockerTarSigPath)
			ScheduleToTerminateDocker = false
			StartDocker(dockerBinPath, keyFilePath, certFilePath, caFilePath)
			Logger.Println("Docker binary updated successfully")
		} else {
			Logger.Println("Cannot verify signature. The update is rejected")
			removeUpdateFiles(dockerTarPath, dockerTarSigPath)
			Logger.Println("Failed to update docker binary")
		}
	}
}

func removeUpdateFiles(dockerTarPath, dockerTarSigPath string) {
	Logger.Println("Removing the invalid docker update file", dockerTarPath)
	if err := os.RemoveAll(dockerTarPath); err != nil {
		SendError(err, "Failed to remove the invalid docker update file", nil)
		Logger.Println(err)
	}
	Logger.Println("Removing the invalid signature file", dockerTarSigPath)
	if err := os.RemoveAll(dockerTarSigPath); err != nil {
		SendError(err, "Failed to remove the invalid docker sig file", nil)
		Logger.Println(err)
	}
}

func verifyDockerSig(dockerNewBinPath, dockerNewBinSigPath string) bool {
	cmd := exec.Command("gpg", "--verify", dockerNewBinSigPath, dockerNewBinPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		SendError(err, "GPG verification failed", nil)
		Logger.Printf("GPG verification failed: %s, %s", err, string(out))
		return false
	}
	Logger.Println("GPG verification passed")
	return true
}

func runDocker(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		SendError(err, "Failed to get docker piped stdout", nil)
		Logger.Println(err)
		Logger.Println("Cannotget docker piped stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		SendError(err, "Failed to get docker piped stdout", nil)
		Logger.Println(err)
		Logger.Println("Cannotget docker piped stdout")
	}

	//open file to log docker logs
	dockerLog := path.Join(LogDir, DockerLogFileName)
	f, err := os.OpenFile(dockerLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		SendError(err, "Failed to set docker log file", nil)
		Logger.Println(err)
		Logger.Println("Cannot set docker log to", dockerLog)
	} else {
		go io.Copy(f, stdout)
		go io.Copy(f, stderr)
		defer f.Close()
	}

	Logger.Println("Starting docker daemon:", cmd.Args)
	if err := cmd.Start(); err != nil {
		SendError(err, "Failed to start docker daemon", nil)
		Logger.Println("Cannot start docker daemon:", err)
	}
	DockerProcess = cmd.Process
	Logger.Printf("Docker daemon (PID:%d) has been started", DockerProcess.Pid)

	syscall.Setpriority(syscall.PRIO_PROCESS, DockerProcess.Pid, RenicePriority)

	exit_renice := make(chan int)

	go decreaseDockerChildProcessPriority(exit_renice)

	if err := cmd.Wait(); err != nil {
		Logger.Println("Docker daemon died with error:", err)
		out, tailErr := exec.Command("tail", "-n", "10", dockerLog).Output()
		if tailErr != nil {
			SendError(tailErr, "Failed to tail docker logs when docker terminates unexpectedly", nil)
			Logger.Printf("Failed to tail docker logs when docker terminates unexpectedly: %s", err)
			SendError(err, "Docker daemon terminates unexpectedly", nil)
		} else {
			extra := map[string]interface{}{"docker-log": string(out)}
			SendError(err, "Docker daemon terminates unexpectedly", extra)
			Logger.Printf("\n=======DOCKER LOGS BEGIN========\n%s=======DOCKER LOGS END========\n", string(out))
		}
	} else {
		Logger.Print("Docker daemon exited")
	}
	exit_renice <- 1
	DockerProcess = nil
}

func decreaseDockerChildProcessPriority(exit_renice chan int) {
	for {
		select {
		case <-exit_renice:
			return
		default:
			out, err := exec.Command("ps", "axo", "pid,ppid,ni").Output()
			if err != nil {
				SendError(err, "Failed to run ps command", nil)
				time.Sleep(ReniceSleepTime * time.Second)
				continue
			}
			lines := strings.Split(string(out), "\n")
			ppids := []int{DockerProcess.Pid}
			for _, line := range lines {
				items := strings.Fields(line)
				if len(items) != 3 {
					continue
				}
				pid, err := strconv.Atoi(items[0])
				if err != nil {
					continue
				}
				ppid, err := strconv.Atoi(items[1])
				if err != nil {
					continue
				}
				ni, err := strconv.Atoi(items[2])
				if err != nil {
					continue
				}
				if ni != RenicePriority {
					continue
				}
				if pid == DockerProcess.Pid {
					continue
				}
				for _, _ppid := range ppids {
					if ppid == _ppid {
						syscall.Setpriority(syscall.PRIO_PROCESS, pid, 0)
						ppids = append(ppids, pid)
						break
					}
				}
			}
			time.Sleep(ReniceSleepTime * time.Second)
		}
	}
}
