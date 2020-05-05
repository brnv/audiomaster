package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type UploadVars struct {
	Acl                 string `json:"acl"`
	Key                 string `json:"key"`
	SuccessActionStatus string `json:"success_action_status"`
	XAmzAlgorithm       string `json:"x-amz-algorithm"`
	XAmzCredential      string `json:"x-amz-credential"`
	XAmzDate            string `json:"x-amz-date"`
	Policy              string `json:"policy"`
	XAmzSignature       string `json:"x-amz-signature"`
}

type RegisterResponse struct {
	StatusURL   string     `json:"statusurl"`
	RemasterURL string     `json:"remasterurl"`
	UploadVars  UploadVars `json:"postvars"`
}

const (
	defaultPresetOptionsFormat = `{"remaster":false,"engine":9,"strength":"normal","name":"%s","ext":"%s","reference":{"enabled":false,"ratio":50,"width":14,"name":"","ext":"","bass":0.5},"equalization":{"low":"normal","mid":"normal","high":"normal"},"stereowidth":"normal","volume":"normal","genre":"","options":{"channels":"stereo","ultralowbassreduction":true},"effects":{"reverb":"none","echo":"none","chorus":"none"},"eq":{"intensity":100,"i1":100,"i2":100,"i3":100},"compressor":{"intensity":100,"limiter":0}}`
)

func registerNewUpload(path string) (RegisterResponse, error) {
	form := url.Values{}

	filename := strings.Split(filepath.Base(path), ".")[0]
	extension := strings.TrimPrefix(filepath.Ext(path), ".")
	humanFilename := strings.Title(filename)

	logger.Debugf("filename: %s", filename)
	logger.Debugf("extension: %s", extension)
	logger.Debugf("humanFilename: %s", humanFilename)

	if filename == "" || extension == "" {
		return RegisterResponse{}, errors.New("filename or extension is empty")
	}

	form.Add("actid", "")
	form.Add("acttoken", "")
	form.Add("action", "new-master")
	form.Add("key", "")
	form.Add("fname", filename)
	form.Add("human_fname", humanFilename)
	form.Add("fext", extension)
	form.Add("vtime", fmt.Sprintf("%d", time.Now().Unix()))
	form.Add("msto", fmt.Sprintf(
		defaultPresetOptionsFormat, filename, extension,
	))

	request, err := http.NewRequest(
		"POST",
		"https://emastered.com/ajax.php",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't create request: %s", err.Error(),
		)
	}

	request.Header.Set("Host", "emastered.com")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:75.0) Gecko/20100101 Firefox/75.0")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	request.Header.Set("Origin", "https://emastered.com")
	request.Header.Set("DNT", "1")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", "https://emastered.com/")
	request.Header.Set("Content-Length", strconv.Itoa(len(form.Encode())))

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't make request: %s", err.Error(),
		)
	}

	responseRaw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't read response body: %s", err.Error(),
		)
	}

	reader, err := gzip.NewReader(bytes.NewReader(responseRaw))
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't create gzip reader: %s", err.Error(),
		)
	}

	responseBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't read gzipped response: %s", err.Error(),
		)
	}

	result := RegisterResponse{}

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf(
			"can't unmarshal response body: %s", err.Error(),
		)
	}

	return result, nil
}
