package main

import "bytes"
import "fmt"
import "github.com/docopt/docopt-go"
import "io"
import "io/ioutil"
import "log"
import "mime/multipart"
import "net/http"
import "os"
import "strings"

var (
	version = "[manual build]"
	usage   = "darc " + version + `

Usage:
  darc [options] [<filename>]
  darc -h | --help
  darc --version

Options:
  -u --url <url>  Upload file to specified server. [default: https://dead.archi/]
  -h --help       Show this screen.
  --version       Show version.
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	url := strings.TrimSuffix(args["--url"].(string), "/") + "/"

	filename, ok := args["<filename>"].(string)
	if !ok {
		filename = "/dev/stdin"
	}

	buffer := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buffer)

	part, err := writer.CreateFormFile("file", filename)
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
