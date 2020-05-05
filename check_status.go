package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CheckStatusResult struct {
	Request struct {
		Success bool   `json:"success"`
		Message string `json:"msg"`
	} `json:"request"`
	Sid    string `json:"sid"`
	Status struct {
		Mastered        bool   `json:"mastered"`
		PercentComplete int    `json:"percentComplete"`
		StatusMessage   string `json:"statusMessage"`
		Error           bool   `json:"error"`
		ErrorMessage    string `json:"errorMessage"`
	} `json:"status"`
	Actions struct {
		MasteredFile string `json:"wf"`
		OriginalFile string `json:"of"`
	} `json:"actions"`
}

func checkStatus(url string) (CheckStatusResult, error) {
	request, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't create request: %s", err.Error(),
		)
	}

	request.Header.Set("Host", "api.emastered.com")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:75.0) Gecko/20100101 Firefox/75.0")
	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Origin", "https://emastered.com")
	request.Header.Set("DNT", "1")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", "https://emastered.com/")

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't make request: %s", err.Error(),
		)
	}

	responseRaw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't read response body: %s", err.Error(),
		)
	}

	reader, err := gzip.NewReader(bytes.NewReader(responseRaw))
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't create gzip reader: %s", err.Error(),
		)
	}

	responseBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't read gzipped response: %s", err.Error(),
		)
	}

	result := CheckStatusResult{}

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return CheckStatusResult{}, fmt.Errorf(
			"can't unmarshal response body: %s", err.Error(),
		)
	}

	if result.Status.Error {
		return CheckStatusResult{}, fmt.Errorf(
			"error: %s", result.Status.ErrorMessage,
		)
	}

	return result, nil
}
