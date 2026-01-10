package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/nilock/tuido/tui"
	"github.com/nilock/tuido/tuido"
	"github.com/nilock/tuido/utils"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			listCmd := flag.NewFlagSet("list", flag.ExitOnError)
			maxItems := listCmd.Int("max", 0, "maximum number of items to display")
			listCmd.Parse(os.Args[2:])

			path := "."
			if listCmd.NArg() > 0 {
				path = listCmd.Arg(0)
			}

			runList(path, *maxItems)
			return

		case "create", "add":
			text := strings.Join(os.Args[2:], " ")
			if text == "" {
				fmt.Println("Usage: tuido create [text]")
				os.Exit(1)
			}
			runCreate(text)
			return

		case "version", "-version", "--version":
			showVersionInfo()
			return
		}
	}

	var showVersion = flag.Bool("version", false, "show version and platform information")
	flag.Parse()

	if *showVersion {
		showVersionInfo()
		os.Exit(0)
	}

	tui.Run()
}

func runList(path string, max int) {
	files := tui.GetFiles(path, tui.GetConfigExtensions())
	items := []*tuido.Item{}
	for _, f := range files {
		items = append(items, tui.GetItems(f)...)
	}

	tui.SortItems(items)

	count := len(items)
	if max > 0 && max < count {
		count = max
	}

	for i := 0; i < count; i++ {
		item := items[i]
		fmt.Printf("%s\n", item.String())
	}
}

func runCreate(text string) {
	writeTo := tui.GetConfigWriteTo()
	item := tuido.New(writeTo, -1, "")
	err := item.SetText(text)
	if err != nil {
		fmt.Printf("Error creating item: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created: %s\n", item.String())
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
