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
			inclSnoozed := listCmd.Bool("z", false, "include snoozed items")
			listCmd.BoolVar(inclSnoozed, "zzz", false, "include snoozed items")
			inclAll := listCmd.Bool("a", false, "include snoozed and completed/cancelled items")
			listCmd.BoolVar(inclAll, "all", false, "include snoozed and completed/cancelled items")
			listCmd.Parse(os.Args[2:])

			path := "."
			if listCmd.NArg() > 0 {
				path = listCmd.Arg(0)
			}

			runList(path, *maxItems, *inclSnoozed || *inclAll, *inclAll)
			return

		case "create", "add":
			text := strings.Join(os.Args[2:], " ")
			if text == "" {
				fmt.Fprintln(os.Stderr, "Usage: tuido create <text>")
				os.Exit(1)
			}
			runCreate(text)
			return

		case "version", "-version", "--version":
			showVersionInfo()
			return

		case "init":
			tui.RunInitWizard()
			return

		case "help", "-help", "--help", "-h":
			printHelp()
			return
		}
	}

	if !tui.ConfigFound {
		fmt.Println("No tuido configuration found. Run `tuido init` to set up.")
		os.Exit(1)
	}

	var showVersion = flag.Bool("version", false, "show version and platform information")
	flag.Usage = printHelp
	flag.Parse()

	if *showVersion {
		showVersionInfo()
		os.Exit(0)
	}

	tui.Run()
}

func runList(path string, max int, inclSnoozed bool, inclDone bool) {
	files := tui.GetFiles(path, tui.GetConfigExtensions())
	all := []*tuido.Item{}
	for _, f := range files {
		all = append(all, tui.GetItems(f)...)
	}

	tui.SortItems(all)

	filtered := all[:0]
	for _, item := range all {
		s := item.Satus()
		if !inclDone && (s == tuido.Checked || s == tuido.Obsolete) {
			continue
		}
		if !inclSnoozed && !item.Active() {
			continue
		}
		filtered = append(filtered, item)
	}

	count := len(filtered)
	if max > 0 && max < count {
		count = max
	}

	for i := 0; i < count; i++ {
		fmt.Printf("%s\n", filtered[i].String())
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

func printHelp() {
	fmt.Print(`Usage: tuido [command] [options]

Without a command, opens the interactive TUI.

Commands:
  list [options] [path]   List open/in-progress items (default path: .)
    -z, --zzz               Include snoozed items
    -a, --all               Include snoozed and completed/cancelled items
    --max N                 Limit output to N items
  create <text>           Create a new todo item
  add <text>              Alias for create
  init                    Create a local or global config (interactive)
  version                 Show version and platform information
  help                    Show this help

Flags:
  -version                Show version and platform information
`)
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
