package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func uploadAudioFile(path string, vars UploadVars) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("acl", vars.Acl)
	writer.WriteField("key", vars.Key)
	writer.WriteField("success_action_status", vars.SuccessActionStatus)
	writer.WriteField("x-amz-algorithm", vars.XAmzAlgorithm)
	writer.WriteField("x-amz-credential", vars.XAmzCredential)
	writer.WriteField("x-amz-date", vars.XAmzDate)
	writer.WriteField("policy", vars.Policy)
	writer.WriteField("x-amz-signature", vars.XAmzSignature)

	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return fmt.Errorf(
			"can't create form file: %s", err.Error(),
		)
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return fmt.Errorf(
			"can't close writer: %s", err.Error(),
		)
	}

	request, err := http.NewRequest(
		"POST",
		"https://emastered.s3-accelerate.amazonaws.com/",
		body,
	)
	if err != nil {
		return fmt.Errorf(
			"can't create request: %s", err.Error(),
		)
	}

	request.Header.Set("Host", "emastered.s3-accelerate.amazonaws.com")
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:75.0) Gecko/20100101 Firefox/75.0")
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Origin", "https://emastered.com")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Referer", "https://emastered.com/")

	client := http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return fmt.Errorf(
			"can't make request: %s", err.Error(),
		)
	}

	return nil
}
