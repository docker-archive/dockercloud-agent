package agent

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TargetDef struct {
	Version             string `json:"version"`
	Download_url        string `json: "download_url"`
	Checksum_md5_url    string `json: "checksum_md5_url"`
	Checksum_sha256_url string `json: "checksum_sha256_url"`
}

func SendRequest(method, url string, data_bytes []byte, headers []string) ([]byte, error) {
	var data io.Reader
	if data_bytes == nil {
		data = nil
	} else {
		data = bytes.NewReader(data_bytes)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for _, header := range headers {
			terms := strings.SplitN(header, " ", 2)
			if len(terms) == 2 {
				req.Header.Add(terms[0], terms[1])
			}
		}
	}
	if *FlagDebugMode {
		Logger.Println("=======Request Info ======")
		Logger.Println("=> URL:", url)
		Logger.Println("=> Method:", method)
		Logger.Println("=> Headers:", req.Header)
		Logger.Println("=> Body:", string(data_bytes))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201, 202:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if *FlagDebugMode {
			Logger.Println("=======Response Info ======")
			Logger.Println("=> Headers:", resp.Header)
			Logger.Println("=> Body:", string(body))
		}
		return body, nil
	default:
		if *FlagDebugMode {
			Logger.Println("=======Response Info (ERROR) ======")
			Logger.Println("=> Headers:", resp.Header)
			b, _ := ioutil.ReadAll(resp.Body)
			Logger.Println("=> Body:", string(b))
		}
		err_msg := fmt.Sprintf("%d", resp.StatusCode)
		return nil, errors.New(err_msg)
	}
}

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		SendError(err, "HTTP get error", nil)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func downloadFile(url, path, name string) {
	Logger.Printf("Downloading %s definition from %s", name, url)
	def := downloadTargetDef(url)

	Logger.Printf("Downloading %s from %s", name, def.Download_url)
	data := downloadTarget(def)
	Logger.Printf("Saving %s to %s", name, path)
	writeToFile(data, path)
}

func downloadTargetDef(url string) *TargetDef {
	def, err := getTargetDef(url)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil || def == nil {
			Logger.Printf("Cannot get target definition: %s. Retry in %d second", err, i)
			time.Sleep(time.Duration(i) * time.Second)
			def, err = getTargetDef(url)
		} else {
			break
		}
	}
	return def
}

func getTargetDef(url string) (*TargetDef, error) {
	var def TargetDef
	body, err := HttpGet(url)
	if err != nil {
		SendError(err, "HTTP get error", nil)
		return nil, err
	}
	if err = json.Unmarshal(body, &def); err != nil {
		SendError(err, "json unmarshal error", nil)
		return nil, err
	}
	if def == (TargetDef{}) {
		SendError(errors.New("Wrong target definition"), "Wrong target definition", nil)
		return nil, errors.New("Wrong target definition")
	}
	return &def, nil
}

func downloadTarget(def *TargetDef) []byte {
	b, err := getTarget(def)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil {
			Logger.Printf("Cannot get target: %s. Retry in %d second", err, i)
			time.Sleep(time.Duration(i) * time.Second)
			b, err = getTarget(def)
		} else {
			break
		}
	}
	return b
}

func getTarget(def *TargetDef) ([]byte, error) {
	b, err := HttpGet(def.Download_url)
	if err != nil {
		SendError(err, "HTTP get error", nil)
		return nil, err
	}

	//validate md5 checksum of the target
	md5hasher := md5.New()
	md5hasher.Write(b)
	md5s := hex.EncodeToString(md5hasher.Sum(nil))
	md5b, err := HttpGet(def.Checksum_md5_url)
	if err != nil {
		SendError(err, "HTTP get error", nil)
		Logger.Println("Failed to get md5 for the target")
		return nil, err
	} else {
		if !strings.Contains(string(md5b), md5s) {
			SendError(errors.New("Failed to pass md5 checksum test"), "Failed on md5 checksum test", nil)
			return nil, errors.New("Failed to pass md5 checksum test")
		}
	}

	//validate sha256 checksum of the target
	sha256hasher := sha256.New()
	sha256hasher.Write(b)
	sha256s := hex.EncodeToString(sha256hasher.Sum(nil))
	sha256b, err := HttpGet(def.Checksum_sha256_url)
	if err != nil {
		SendError(err, "HTTP error", nil)
		Logger.Println("Failed to get sha256 for the target")
		return nil, err
	} else {
		if !strings.Contains(string(sha256b), sha256s) {
			SendError(errors.New("Failed to pass sha256 checksum test"), "Failed on sha256 checksum test", nil)
			return nil, errors.New("Failed to pass sha256 checksum test")
		}
	}

	return b, nil
}

func writeToFile(binary []byte, path string) {
	err := ioutil.WriteFile(path, binary, 0755)
	for i := 1; ; i *= 2 {
		if i > MaxWaitingTime {
			i = 1
		}
		if err != nil {
			SendError(err, "Failed to write to file", nil)
			Logger.Printf("Failed to save the target: %s. Retrying in %d second", err, i)
			time.Sleep(time.Duration(i) * time.Second)
			err = ioutil.WriteFile(path, binary, 0755)
		} else {
			break
		}
	}
}
