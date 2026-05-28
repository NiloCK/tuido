package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	// frictionThreshold is the number of items that can be displayed or added to
	// before a nag deterrent is displayed.
	frictionThreshold int
}

func (cfg config) String() string {
	return fmt.Sprintf("extensions=%s\nwriteto=%s\nfrictionThreshold=%d",
		strings.Join(cfg.extensions, ","), cfg.writeto, cfg.frictionThreshold)
}

// runConfig is the initial, default values for the application configuration.
//
// its `writeto` value gets overwritten in `init()` by a golang lookup of,
// in practice, the same value, but hopefully in a cross-platform safe way.
//
// **all** values are overwritten in `loadFromDefaultConfigLocation()` via
// `init()`, if a configuration file is found in the default location.
var runConfig config = config{
	extensions:        []string{"xit", "md", "txt"},
	writeto:           "~/.tuido",
	frictionThreshold: 5,
}

func parseConfigIfExists(configPath string) *config {

	if config, err := os.Open(configPath); err == nil {
		cfg := parseConfig(config)
		return &cfg
	}
	return nil
}

// parseConfig reads a file for tuido configuration flags according
// to the following. It:
//   - reads from the first line of the file
//   - pulls one config flag from each line
//   - ends reading the file when it encounters a line with no config flags
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
			if split[0] == "frictionThreshold" {
				n, err := strconv.Atoi(split[1])
				if err == nil {
					cfg.frictionThreshold = n
				}
			}

		} else {
			// not a config line:
			return cfg
		}
	}

	return cfg
}

func (cfg config) hasSettings() bool {
	return len(cfg.extensions) > 0 || cfg.writeto != ""
}

// RunInitWizard interactively creates a local or global tuido config file.
func RunInitWizard() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Configure locally (./.tuido) or globally (~/.config/tuido.conf)? [l/g]: ")
	choice, _ := reader.ReadString('\n')
	isLocal := strings.ToLower(strings.TrimSpace(choice)) != "g"

	var configPath, defaultWriteto string
	if isLocal {
		cwd, _ := os.Getwd()
		configPath = filepath.Join(cwd, ".tuido")
		defaultWriteto = filepath.Join(cwd, ".tuido")
	} else {
		cfgDir, _ := os.UserConfigDir()
		configPath = filepath.Join(cfgDir, "tuido.conf")
		home, _ := os.UserHomeDir()
		defaultWriteto = filepath.Join(home, ".tuido")
	}

	fmt.Printf("Where should new items be written? [%s]: ", defaultWriteto)
	writeto, _ := reader.ReadString('\n')
	writeto = strings.TrimSpace(writeto)
	if writeto == "" {
		writeto = defaultWriteto
	}

	defaultExt := strings.Join(runConfig.extensions, ",")
	fmt.Printf("File extensions to scan (comma-separated)? [%s]: ", defaultExt)
	extInput, _ := reader.ReadString('\n')
	extInput = strings.TrimSpace(extInput)
	if extInput == "" {
		extInput = defaultExt
	}

	if err := writeConfigFile(configPath, extInput, writeto); err != nil {
		fmt.Printf("Error writing config to %s: %v\n", configPath, err)
		os.Exit(1)
	}
	fmt.Printf("Config written to %s\n", configPath)
}

// writeConfigFile writes config lines to path, preserving any non-config
// content (todo items) that already exists below the config header.
func writeConfigFile(path, extensions, writeto string) error {
	var itemLines []string
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		pastConfig := false
		for scanner.Scan() {
			line := scanner.Text()
			if !pastConfig {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					continue
				}
				pastConfig = true
			}
			itemLines = append(itemLines, line)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("extensions=%s\n", extensions))
	sb.WriteString(fmt.Sprintf("writeto=%s\n", writeto))
	for _, line := range itemLines {
		sb.WriteString(line + "\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func GetConfigExtensions() []string {
	return runConfig.extensions
}

func GetConfigWriteTo() string {
	return runConfig.writeto
}
