package agent

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	f, err := ioutil.TempFile("", "loadconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	defer f.Close()
	defer os.RemoveAll(name)
	testFile := []byte(`{
	"LogSizeLimit": 5,
	"LogRotateInterval": 60,
	"LogTailLines": 30,
	"MetricsCollectInterval": 60,
	"DockerHost":"unix:///run/docker.sock",
	"Token":"abcdefg",
	"Host":"http://cloud.docker.com/",
	"CertCommonName":"*"
}`)
	if _, err := f.Write(testFile); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadConf(name); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigFile_EmptyFile(t *testing.T) {
	f, err := ioutil.TempFile("", "loadconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	defer f.Close()
	defer os.RemoveAll(name)

	if _, err := LoadConf(name); err == nil {
		t.Fatal("Excepted error: Empty json config file")
	}
}

func TestLoadConfigFile_MalformattedConfigFile(t *testing.T) {
	f, err := ioutil.TempFile("", "loadconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	defer f.Close()
	defer os.RemoveAll(name)
	testFile := []byte(`{
    "DebugMode": true
    "DaemonMode": false
    "RegistrationToken": "YourDockerCloudToken"
    "LogsizeLimit": "5M"
    "LogRotateInterval": 15
    "LogTailLines": 20
}`)
	if _, err := f.Write(testFile); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadConf(name); err == nil {
		t.Fatal("Excepted error: Malformatted json config file")
	}
}

func TestLoadConfigFile_FileNotExist(t *testing.T) {
	f, err := ioutil.TempFile("", "loadconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	os.RemoveAll(name)

	if _, err := LoadConf(name); err == nil {
		t.Fatal("Excepted error: File not exist")
	}
}
