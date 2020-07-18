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
	Update  UpdateFlag  `name:"update" help:"Check for an updated version"`
}

var (
	// CLI
	cli struct {
		Globals

		// flags
		URL       string `default:"https://transfer.sh" help:"Transfer.sh Service URL"`
		User      string `help:"Transfer.sh Basic Auth Username"`
		Pass      string `help:"Transfer.sh Basic Auth Password"`
		Downloads int    `help:"Maximum amount of downloads"`
		Days      int    `help:"Maximum amount of days"`
		Filename  string `help:"Name of file when uploaded"`

		// args
		Filepath string `arg:"" required:"1" name:"filepath" help:"File to upload"`
	}
)

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

	// transfer
	os.Exit(transferFile())
}

func transferFile() (exitCode int) {
	// validate filepath argument
	fi, err := os.Stat(cli.Filepath)
	if err != nil {
		fmt.Println("Failed getting info of file to transfer:", err)
		return 1
	}

	// archive folder
	if fi.IsDir() {
		// get temporary path for the archive file
		tf, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s_*.zip", fi.Name()))
		if err != nil {
			fmt.Println("Failed getting temporary archive path:", err)
		}

		// remove the temporary archive before exiting
		defer func(path string) {
			if err := os.Remove(path); err != nil {
				fmt.Println("Failed removing temporary archive:", err)
				exitCode = 1
			}
		}(tf.Name())

		// archive directory
		if err := archiveFolder(cli.Filepath, tf.Name()); err != nil {
			fmt.Println("Failed archiving to temporary archive:", err)
			return 1
		}

		// validate new filepath to transfer
		fi, err = os.Stat(tf.Name())
		if err != nil {
			fmt.Println("Failed getting info of temporary archive file to transfer:", err)
			return 1
		}

		// update file transfer detail(s)
		cli.Filepath = tf.Name()
	}

	if cli.Filename == "" {
		cli.Filename = fi.Name()
	}

	// open file for transfer
	f, err := os.Open(cli.Filepath)
	if err != nil {
		fmt.Println("Failed opening file to transfer:", err)
		return 1
	}

	defer f.Close()

	// prepare upload
	bar := pb.New64(fi.Size()).Set(pb.Bytes, true).SetRefreshRate(time.Millisecond * 10)
	bar.Start()
	defer bar.Finish()

	// prepare request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", cli.URL, cli.Filename), bar.NewProxyReader(f))
	if err != nil {
		fmt.Println("Failed creating file transfer request:", err)
		return 1
	}

	ct, err := getContentFileType(f)
	if err != nil {
		fmt.Println("Failed determining content type for file transfer request:", err)
		return 1
	}

	req.Header.Set("Content-Type", ct)
	if cli.Days > 0 {
		req.Header.Set("Max-Days", strconv.Itoa(cli.Days))
	}
	if cli.Downloads > 0 {
		req.Header.Set("Max-Downloads", strconv.Itoa(cli.Downloads))
	}

	if cli.User != "" && cli.Pass != "" {
		req.SetBasicAuth(cli.User, cli.Pass)
	}

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed sending file transfer request:", err)
		return 1
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("Failed validating file transfer response, unexpected status:", res.Status)
		return 1
	}

	// read response
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed reading file transfer response body:", err)
		return 1
	}

	// get delete link
	fmt.Println("")
	fmt.Println("Download URL:", string(b))

	deleteURL := res.Header.Get("X-Url-Delete")
	if deleteURL != "" {
		fmt.Println("---")
		fmt.Println("Delete URL:", deleteURL)
	}

	return 0
}

func getContentFileType(out *os.File) (string, error) {
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	if _, err := out.Seek(0, 0); err != nil {
		return "", err
	}

	return contentType, nil
}
