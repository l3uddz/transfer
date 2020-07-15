package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/cheggaaa/pb/v3"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Globals struct {
	Version VersionFlag `name:"version" help:"Print version information and quit"`
}

var (
	// Build Vars
	Version   string
	Timestamp string
	GitCommit string

	// CLI
	cli struct {
		Globals

		// flags
		URL          string `default:"https://transfer.sh" help:"Transfer.sh Service URL"`
		User         string `help:"Transfer.sh Basic Auth Username"`
		Pass         string `help:"Transfer.sh Basic Auth Password"`
		MaxDownloads int    `help:"Max Downloads"`
		MaxDays      int    `help:"Max Days"`

		// args
		Filepath string `arg:"" required:"1" name:"filepath" help:"File to upload"`
	}
)

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

func main() {
	// parse cli
	ctx := kong.Parse(&cli,
		kong.Name("transfer"),
		kong.Description("Upload files to transfer.sh"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.Vars{
			"version": fmt.Sprintf("%s (%s@%s)", Version, GitCommit, Timestamp),
		},

		kong.Configuration(kong.JSON, "~/.transfer.json", "~/.config/transfer/transfer.json"),
	)

	if err := ctx.Validate(); err != nil {
		fmt.Println("Failed parsing cli:", err)
		os.Exit(1)
	}

	// validate filepath argument
	fi, err := os.Stat(cli.Filepath)
	if err != nil {
		fmt.Println("Failed getting info of file to transfer:", err)
		os.Exit(1)
	}

	f, err := os.Open(cli.Filepath)
	if err != nil {
		fmt.Println("Failed opening file to transfer:", err)
		os.Exit(1)
	}

	defer f.Close()

	// prepare upload
	bar := pb.New64(fi.Size()).Set(pb.Bytes, true).SetRefreshRate(time.Millisecond * 10)
	bar.Start()
	defer bar.Finish()

	// prepare request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", cli.URL, fi.Name()), bar.NewProxyReader(f))
	if err != nil {
		fmt.Println("Failed creating file transfer request:", err)
		os.Exit(1)
	}

	ct, err := getContentFileType(f)
	if err != nil {
		fmt.Println("Failed determining content type for file transfer request:", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", ct)
	if cli.MaxDays > 0 {
		req.Header.Set("Max-Days", strconv.Itoa(cli.MaxDays))
	}
	if cli.MaxDownloads > 0 {
		req.Header.Set("Max-Downloads", strconv.Itoa(cli.MaxDownloads))
	}

	if cli.User != "" && cli.Pass != "" {
		req.SetBasicAuth(cli.User, cli.Pass)
	}

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed sending file transfer request:", err)
		os.Exit(1)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("Failed validating file transfer response, unexpected status:", res.Status)
		os.Exit(1)
	}

	// read response
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed reading file transfer response body:", err)
		os.Exit(1)
	}

	// get delete link
	fmt.Println("")
	fmt.Println("Download URL:", string(b))

	deleteURL := res.Header.Get("X-Url-Delete")
	if deleteURL != "" {
		fmt.Println("---")
		fmt.Println("Delete URL:", deleteURL)
	}
}

func getContentFileType(out *os.File) (string, error) {
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	_, _ = out.Seek(0, 0)
	return contentType, nil
}
