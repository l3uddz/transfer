package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"os"
)

type UpdateFlag string

func (f UpdateFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (f UpdateFlag) IsBool() bool                         { return true }
func (f UpdateFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	// parse current version
	v, err := semver.Parse(Version)
	if err != nil {
		fmt.Println("Failed parsing current build version:", err)
		app.Exit(1)
	}

	// detect latest version
	fmt.Println("Checking for the latest version...")
	latest, found, err := selfupdate.DetectLatest("l3uddz/transfer")
	if err != nil {
		fmt.Println("Failed determining latest available version:", err)
		app.Exit(1)
	}

	// check version
	if !found || latest.Version.LTE(v) {
		fmt.Println("Already using the latest version:", Version)
		app.Exit(0)
	}

	// ask update
	fmt.Println("Do you want to update to the latest version (y/n):", latest.Version)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Println("Failed validating input...")
		app.Exit(1)
	} else if input == "n\n" {
		app.Exit(0)
	}

	// get existing executable path
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Failed locating current executable path:", err)
		app.Exit(1)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		fmt.Println("Failed updating existing binary to latest release:", err)
		app.Exit(1)
	}

	fmt.Println("Successfully updated to the latest version:", latest.Version)
	app.Exit(0)
	return nil
}
