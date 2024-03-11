package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/cheggaaa/pb/v3"
)

const (
	// Version vars. Will be set during build
	Version = "1.0.0"
	Timestamp = "2021-01-01T00:00:00Z"
	GitCommit = "0000000"
	Repo = "azhinu/transfer"
)

var (
	// CLI
	cli struct {
		// flags
		Version 	bool `name:"version" help:"Print version information and quit"`
		Update 		bool `name:"update" help:"Check for an updated version"`

		URL       string `short:"u" default:"transfer.sh" help:"Transfer.sh Service URL" env:"TRANSFER_URL"`
		User      string `placeholder:"username" help:"Transfer.sh Basic Auth Username" env:"TRANSFER_USER"`
		Pass      string `placeholder:"secret" help:"Transfer.sh Basic Auth Password" env:"TRANSFER_PASS"`
		Downloads int    `short:"d" default:"5" help:"Maximum amount of downloads" env:"TRANSFER_DOWNLOADS"`
		Days      int    `short:"D" default:"5" help:"Maximum amount of days" env:"TRANSFER_DAYS"`
		Filename  string `short:"n" placeholder:"my_file" help:"Name of file when uploaded"`

		// args
		// Needs to test type
		Filepath string `arg:"" optional:"" type:"path" help:"File or directory to upload"`
	}
)

func main() {
	// parse cli
	ctx := kong.Parse(&cli,
		kong.Name("transfer"),
		kong.Description("Upload files to transfer.sh"),
		kong.UsageOnError(),
		kong.Configuration(kongyaml.Loader, "~/.transfer.yml", "~/.config/transfer/transfer.yml"),
	)
	
	if cli.Version {
		fmt.Println("Version:", Version, "GitCommit:", GitCommit, "Timestamp:", Timestamp)
		os.Exit(0)
	}
	
	if cli.Update {
		if err := selfUpdate(); err != nil {
			os.Exit(1)
			} else {
				os.Exit(0)
			}
		}
		
	if cli.Filepath == "" {
		err := ctx.PrintUsage(true)
		if err != nil {
			fmt.Println("Failed printing usage:", err)
		}
		os.Exit(0)
	}

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
		tf, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s_*.zip", fi.Name()))
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

	// Append https if schema not present
	if cli.URL[:4] != "http" {
		cli.URL = "https://" + cli.URL
	}

	// Remove tailing slash
	if cli.URL[len(cli.URL)-1:] == "/" {
		cli.URL = cli.URL[:len(cli.URL)-1]
	}


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
	req.Header.Set("Max-Days", strconv.Itoa(cli.Days))
	req.Header.Set("Max-Downloads", strconv.Itoa(cli.Downloads))

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
		fmt.Println("Failed to send file, unexpected status:", res.Status)
		return 1
	}

	// read response
	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed to get download link:", err)
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
