package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/docker/dockercloud-agent/utils"
)

type RegResponseForm struct {
	UserCaCert      string `json:"user_ca_cert"`
	UUID            string `json:"uuid"`
	CertCommonName  string `json:"external_fqdn"`
	DockerTarURL    string `json:"docker_url"`
	NgrokTarURL     string `json:"ngrok_url"`
	PublicIpAddress string `json:"public_ip"`
}

type RegPostForm struct {
	AgentVersion string `json:"agent_version"`
}

type RegPatchForm struct {
	Public_cert   string `json:"public_cert"`
	AgentVersion  string `json:"agent_version"`
	DockerVersion string `json:"docker_version,omitempty"`
}

type RegGetForm struct {
	AgentVersion string `json:"agent_version"`
	DockerUrl    string `json:"docker_url"`
	ExternalFqdn string `json:"external_fqdn"`
	NgrokUrl     string `json:"ngrok_url"`
	PublicCert   string `json:"public_cert"`
	ResourceUri  string `json:"resource_uri"`
	State        string `json:"state"`
	Tunnel       string `json:"tunnel"`
	UserCaCert   string `json:"user_ca_cert"`
	UUID         string `json:"uuid"`
	NgrokHost    string `json:"ngrok_server_addr"`
}

func RegPost(url, caFilePath, configFilePath string) error {
	form := RegPostForm{}
	form.AgentVersion = VERSION
	data, err := json.Marshal(form)
	if err != nil {
		SendError(err, "Fatal: Json marshal error", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatal("Cannot marshal the POST form", err)
	}
	return register(url, "POST", Conf.Token, Conf.UUID, caFilePath, configFilePath, data)
}

func RegPatch(url, caFilePath, certFilePath, configFilePath string) error {
	form := RegPatchForm{}
	form.AgentVersion = VERSION
	form.DockerVersion = GetDockerClientVersion(path.Join(DockerHome, DockerBinaryName))
	cert, err := GetCertificate(certFilePath)
	if err != nil {
		SendError(err, "Fatal: Failed to load public certificate", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatal("Cannot read public certificate:", err)
	}
	form.Public_cert = *cert
	data, err := json.Marshal(form)
	if err != nil {
		SendError(err, "Fatal: Json marshal error", nil)
		os.RemoveAll(AgentPidFile)
		Logger.Fatal("Cannot marshal the PATCH form", err)
	}

	return register(url, "PATCH", Conf.Token, Conf.UUID, caFilePath, configFilePath, data)
}

func VerifyRegistration(url string) {
	headers := []string{"Authorization TutumAgentToken " + Conf.Token,
		"Content-Type application/json",
		"User-Agent dockercloud-agent/" + VERSION}
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
			if form.State == "Deployed" {
				Logger.Printf("Node registration to %s succeeded", Conf.Host)
				return
			}
		}
	}

	time.Sleep(5 * time.Minute)

	body, err = SendRequest("GET", utils.JoinURL(url, Conf.UUID), nil, headers)
	if err != nil {
		SendError(err, "Failed to get registration info after 5 minutes", nil)
		Logger.Printf("Get registration info error, %s", err)
	} else {
		var form RegGetForm
		if err = json.Unmarshal(body, &form); err != nil {
			SendError(err, "Json unmarshal error", nil)
			Logger.Println("Cannot unmarshal the response", err)
		} else {
			if form.State == "Deployed" {
				Logger.Printf("Node registration to %s succeeded", Conf.Host)
			} else {
				Logger.Printf("Node registration to %s timed out", Conf.Host)
				Logger.Println("Node state:", form.State)
			}
		}
	}
}

func register(url, method, token, uuid, caFilePath, configFilePath string, data []byte) error {
	if token == "" {
		fmt.Fprintf(os.Stderr, "The token is empty. Please run 'dockercloud-agent set Token=xxx' first!\n")
		os.RemoveAll(AgentPidFile)
		Logger.Fatal("The token is empty. Please run 'dockercloud-agent set Token=xxx' first!")
	}

	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		body, err := sendRegRequest(url, method, token, uuid, data)
		if err == nil {
			if err = handleRegResponse(body, caFilePath, configFilePath); err == nil {
				return nil
			} else {
				Logger.Printf("Failed to handle the registration response, %s. Retry in %d seconds", err, i)
				time.Sleep(time.Duration(i) * time.Second)
				continue
			}
		}
		if method == "POST" && (err.Error() == "401") {
			SendError(err, "Registration unauthorized: POST", nil)
			Logger.Print("Cannot register node in Docker Cloud: unauthorized. Please try again with a new token.")
			Logger.Print("Removing the invalid token from config file")
			os.RemoveAll(AgentPidFile)
			Conf.Token = ""
			if err := SaveConf(path.Join(AgentHome, ConfigFileName), Conf); err != nil {
				SendError(err, "Failed to save config to the conf file", nil)
				Logger.Print(err)
			}
			time.Sleep(10 * time.Second)
			Logger.Fatal("Docker Cloud agent is terminated")
		}
		if method == "PATCH" && (err.Error() == "404" || err.Error() == "401") {
			return err
		}
		SendError(err, "Registration HTTP error", nil)
		Logger.Printf("Registration failed, %s. Retry in %d seconds", err, i)
		time.Sleep(time.Duration(i) * time.Second)
	}
}

func sendRegRequest(url, method, token, uuid string, data []byte) ([]byte, error) {
	headers := []string{"Authorization TutumAgentToken " + token,
		"Content-Type application/json",
		"User-Agent dockercloud-agent/" + VERSION}
	return SendRequest(method, utils.JoinURL(url, uuid), data, headers)
}

func handleRegResponse(body []byte, caFilePath, configFilePath string) error {
	var responseForm RegResponseForm

	// Save user ca cert file
	if err := json.Unmarshal(body, &responseForm); err != nil {
		SendError(err, "Json unmarshal error", nil)
		Logger.Println("Failed to unmarshal the response", err)
		return err
	}
	if err := ioutil.WriteFile(caFilePath, []byte(responseForm.UserCaCert), 0644); err != nil {
		SendError(err, "Failed to save user ca cert file", nil)
		Logger.Println("Failed to save", caFilePath, err)
		return err
	}
	// Update global Conf
	isModified := false
	if Conf.CertCommonName != responseForm.CertCommonName {
		Logger.Printf("Cert CommonName has been changed from %s to %s", Conf.CertCommonName, responseForm.CertCommonName)
		isModified = true
		Conf.CertCommonName = responseForm.CertCommonName
	}
	if Conf.UUID != responseForm.UUID {
		Logger.Printf("UUID has been changed from %s to %s", Conf.UUID, responseForm.UUID)
		isModified = true
		Conf.UUID = responseForm.UUID
	}

	DockerTarURL = responseForm.DockerTarURL

	if responseForm.NgrokTarURL != "" {
		NgrokTarURL = responseForm.NgrokTarURL
	}
	// Save to configuration file
	if isModified {
		Logger.Println("Updating configuration file...")
		return SaveConf(configFilePath, Conf)
	}
	return nil
}
