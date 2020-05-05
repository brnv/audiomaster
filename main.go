package main

import (
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
    audiomaster --file <string> [options]
    audiomaster -h | --help

Options:
    --file <string>   Email.
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

	logger.Debugf("filepath: %s", path)

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
			logger.Infof("original file: %s", checkResult.Actions.OriginalFile)
			logger.Infof("mastered file: %s", checkResult.Actions.MasteredFile)
			break
		}
	}
}
