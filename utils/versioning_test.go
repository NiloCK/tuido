package utils_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/nilock/tuido/utils"
)

func TestVersion(t *testing.T) {
	// get latest tag

	// print working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting working directory: %s", err)
	}

	fmt.Println("PWD:", wd)

	gitRepo, err := git.PlainOpen("..")
	if err != nil {
		t.Fatalf("Error opening git repo: %s", err)
	}

	tagRefs, err := gitRepo.Tags()
	tags := []string{}

	tagRefs.ForEach(func(tag *plumbing.Reference) error {
		tags = append(tags, tag.Name().Short())
		return nil
	})

	fmt.Println("TAGS:", tags)

	// get latest tag
	var latestTag string
	var latestMajor, latestMinor, latestPatch = -1, -1, -1

	for _, tag := range tags {
		// Remove 'v' prefix
		version := strings.TrimPrefix(tag, "v")

		split := strings.Split(version, ".")
		if len(split) != 3 {
			continue
		}

		major, err := strconv.Atoi(split[0])
		if err != nil {
			continue
		}
		minor, err := strconv.Atoi(split[1])
		if err != nil {
			continue
		}
		patch, err := strconv.Atoi(split[2])
		if err != nil {
			continue
		}

		// First tag or higher major version
		if latestMajor == -1 || major > latestMajor {
			latestMajor, latestMinor, latestPatch = major, minor, patch
			latestTag = tag
			continue
		}

		// Same major, higher minor
		if major == latestMajor && minor > latestMinor {
			latestMajor, latestMinor, latestPatch = major, minor, patch
			latestTag = tag
			continue
		}

		// Same major and minor, higher patch
		if major == latestMajor && minor == latestMinor && patch > latestPatch {
			latestMajor, latestMinor, latestPatch = major, minor, patch
			latestTag = tag
		}
	}

	want := latestTag

	if got := utils.Version(); got != want {
		t.Errorf("Version() = %v, want %v", got, want)
	}
}
