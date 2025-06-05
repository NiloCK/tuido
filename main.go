package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/nilock/tuido/tui"
	"github.com/nilock/tuido/utils"
)

func main() {
	var showVersion = flag.Bool("version", false, "show version and platform information")
	flag.Parse()

	if *showVersion {
		showVersionInfo()
		os.Exit(0)
	}

	tui.Run()
}

func showVersionInfo() {
	version := utils.Version()
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	fmt.Printf("tuido %s\n", version)
	fmt.Printf("Platform: %s/%s\n", goos, goarch)

	// Show expected asset name for this platform
	assetName := utils.BuildAssetName(version, goos, goarch)
	fmt.Printf("Asset: %s\n", assetName)

	// Show if this matches a known release asset (with graceful failure)
	if asset, err := utils.GetCurrentPlatformAsset(); err == nil {
		fmt.Printf("Available: %s (%d bytes)\n", asset.Name, asset.Size)
	} else {
		fmt.Printf("Available: (unable to check - %v)\n", err)
	}
}
