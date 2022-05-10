package tui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix()) // a fresh set of tag colors on each run. Spice of life.

	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %s", err)
	}
	tuidoDir := filepath.Join(dir, ".tuido")
	runConfig.writeto = tuidoDir

	loadFromDefaultConfigLocation()

	// make sure the write target exists
	_, err = os.Open(runConfig.writeto)
	if err != nil {
		err = os.Mkdir(runConfig.writeto, 0777)
		if err != nil {
			fmt.Printf("error creating appDirectory %s': %v\n",
				runConfig.writeto, err)
		}
	}
}

func loadFromDefaultConfigLocation() {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error seeking configdir")
		return
	}

	cfgPath := filepath.Join(cfgDir, "tuido.conf")
	cfg := parseConfigIfExists(cfgPath)

	if cfg != nil {
		if len(cfg.extensions) != 0 {
			runConfig.extensions = cfg.extensions
		}
		if cfg.writeto != "" {
			runConfig.writeto = cfg.writeto
		}
	}
}
