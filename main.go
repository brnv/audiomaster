package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
)

var (
	logger  = lorg.NewLog()
	version = "[manual build]"
)

const usage = `audiomaster

Usage:
    audiomaster --file <path> [options]
    audiomaster -h | --help

Options:
    --file <path>     File to master.
    --output <path>   Output file [default: ./mastered.mp3].
    --debug           Enable debug output.
    --trace           Enable trace output.
    -h --help         Show this help.
`

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	logger.SetIndentLines(true)

	if args["--debug"].(bool) {
		logger.SetLevel(lorg.LevelDebug)
	}

	if args["--trace"].(bool) {
		logger.SetLevel(lorg.LevelTrace)
	}

	path := args["--file"].(string)
	masteredDestination := args["--output"].(string)

	logger.Debugf("filepath: %s", path)
	logger.Debugf("mastered file destination: %s", masteredDestination)

	file, err := os.Open(path)
	if err != nil {
		logger.Fatal(err)
	}
	file.Close()

	registerResponse, err := registerNewUpload(path)
	if err != nil {
		logger.Fatal(err)
	}

	uploadVars := registerResponse.UploadVars

	statusURL := registerResponse.StatusURL

	err = uploadAudioFile(path, uploadVars)
	if err != nil {
		logger.Fatal(err)
	}

	for {
		time.Sleep(time.Second * 1)

		checkResult, err := checkStatus(statusURL)
		if err != nil {
			logger.Fatal(err)
		}

		progress := checkResult.Status.PercentComplete

		logger.Infof("progress: %d%%", progress)

		if progress == 100 || checkResult.Status.Mastered {
			err = download(masteredDestination, checkResult.Actions.MasteredFile)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infof("saved to %s", masteredDestination)

			break
		}
	}
}

func download(filepath string, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	destination, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, response.Body)
	if err != nil {
		return err
	}

	return nil
}
