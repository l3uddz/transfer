package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func selfUpdate () error {
		// parse current version
		v, err := semver.Parse(Version)
		if err != nil {
			fmt.Println("Failed parsing current build version:", err)
			return err
		}
	
		// detect latest version
		fmt.Println("Checking for the latest version...")
		latest, found, err := selfupdate.DetectLatest(Repo)
		if err != nil {
			fmt.Println("Failed determining latest available version:", err)
			return err
		}
	
		// check version
		if !found || latest.Version.LTE(v) {
			fmt.Println("Already using the latest version:", Version)
			return err
		}
	
		// ask update
		fmt.Println("Do you want to update to the latest version (y/n):", latest.Version)
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil || (input != "y\n" && input != "n\n") {
			fmt.Println("Failed validating input...")
			return err
		}
		if input == "n\n" {
			fmt.Println("Skipping update...")
			return nil
		}
	
		// get existing executable path
		exe, err := os.Executable()
		if err != nil {
			fmt.Println("Failed locating current executable path:", err)
			return err
		}
	
		if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
			fmt.Println("Failed updating existing binary to latest release:", err)
			return err
		}
	
		fmt.Println("Successfully updated to", latest.Version)
		return nil
}
