package tui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// ConfigFound is true if a local or global config was loaded at startup.
// When false, the caller should prompt the user to run `tuido init`.
var ConfigFound bool

func init() {
	rand.Seed(time.Now().Unix())

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %s", err)
	}
	runConfig.writeto = filepath.Join(home, ".tuido")

	ConfigFound = loadConfig()

	_, err = os.Open(runConfig.writeto)
	if err != nil {
		err = os.Mkdir(runConfig.writeto, 0777)
		if err != nil {
			fmt.Printf("error creating appDirectory %s': %v\n", runConfig.writeto, err)
		}
	}
}

// loadConfig loads configuration with priority: cwd .tuido > global config.
// Returns true if any config was found.
func loadConfig() bool {
	cwd, err := os.Getwd()
	if err == nil {
		if cfg := parseConfigIfExists(filepath.Join(cwd, ".tuido")); cfg != nil && cfg.hasSettings() {
			applyConfig(cfg)
			return true
		}
	}
	return loadFromDefaultConfigLocation()
}

func loadFromDefaultConfigLocation() bool {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return false
	}
	cfg := parseConfigIfExists(filepath.Join(cfgDir, "tuido.conf"))
	if cfg == nil || !cfg.hasSettings() {
		return false
	}
	applyConfig(cfg)
	return true
}

func applyConfig(cfg *config) {
	if len(cfg.extensions) != 0 {
		runConfig.extensions = cfg.extensions
	}
	if cfg.writeto != "" {
		runConfig.writeto = cfg.writeto
	}
	if len(cfg.exclude) != 0 {
		runConfig.exclude = cfg.exclude
	}
}
