package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
)

var (
	version = "[manual build]"
	usage   = "darc " + version + `

Usage:
  darc [options] [<filename>]
  darc -h | --help
  darc --version

Options:
  -u --url <url>    Upload file to specified server. [default: https://dead.archi/]
  -d --auto-delete  Automatically delete after downloading.
  -l --download     Send headers to download file.
  -h --help         Show this screen.
  --version         Show version.
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	url := strings.TrimSuffix(args["--url"].(string), "/") + "/"
	if !strings.Contains(url, "://") {
		log.Fatalf("URL should contain a scheme")
	}

	filename, ok := args["<filename>"].(string)
	if !ok {
		filename = "/dev/stdin"
	}

	buffer := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buffer)

	if args["--auto-delete"].(bool) {
		err = writer.WriteField("auto_delete", "1")
		if err != nil {
			log.Fatalln(err)
		}
	}

	if args["--download"].(bool) {
		err = writer.WriteField("header_download", "1")
		if err != nil {
			log.Fatalln(err)
		}
	}

	announcedName := filename
	if filename == "/dev/stdin" {
		announcedName = "stdin.txt"
	}

	part, err := writer.CreateFormFile("file", announcedName)
	if err != nil {
		log.Fatalln(err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatalln(err)
	}

	err = writer.Close()
	if err != nil {
		log.Fatalln(err)
	}

	request, err := http.NewRequest("POST", args["--url"].(string), buffer)
	if err != nil {
		log.Fatalln(err)
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("User-Agent", "darc/"+version)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("unexpected server status: %s", response.Status)
	}

	token, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(url + strings.TrimSpace(string(token)))
}
