package main

import (
	"flag"
	"os"

	"github.com/nilock/tuido/tui"
	"github.com/nilock/tuido/utils"
)

func main() {
	vFlag := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *vFlag {
		v := utils.Version()
		println(v)
		os.Exit(0)
	}

	tui.Run()
}
