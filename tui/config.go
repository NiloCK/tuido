package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type config struct {
	// extensions is a collection of file extensions that will be parsed for items
	//
	// default value for extensions is ["xit", "md", "txt"].
	extensions []string

	// writeto is the location that items created in-app will be appended to.
	//
	// writeto can be either:
	//  - a file, which will have new items appended as new lines, or
	//  - a directory, which will be written with YYYY-MM-DD.xit files for each day
	writeto string
}

var runConfig config = config{
	extensions: []string{"xit", "md", "txt"},
	writeto:    "~/.tuido",
}

func adoptConfigSettings(location string) {
	config := parseConfigIfExists(location)

	if config != nil {
		runConfig.extensions = append(runConfig.extensions, config.extensions...)
		if config.writeto != "" {
			runConfig.writeto = config.writeto
		}
	}
}

func parseConfigIfExists(configPath string) *config {

	if config, err := os.Open(configPath); err == nil {
		cfg := parseConfig(config)
		fmt.Printf("reading config %s: %+v\n", configPath, cfg)
		return &cfg
	}
	return nil
}

// parseConfig reads a file for tuido configuration flags according
// to the following. It:
//  - reads from the first line of the file
//  - pulls one config flag from each line
//  - ends reading the file when it encounters a line with no config flags
//
// This allows the .tuido file to be used as both configuration and as an
// append target for new items authored in-tui.
func parseConfig(file *os.File) config {
	cfg := config{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		split := strings.Split(line, "=")

		if len(split) == 2 { // all tuido config lines are of the form "flag=value[,value[,value...]]"

			if split[0] == "extensions" {
				cfg.extensions = strings.Split(split[1], ",")
			}
			if split[0] == "writeto" {
				cfg.writeto = split[1]
			}

		} else {
			// not a config line:
			return cfg
		}
	}

	return cfg
}
