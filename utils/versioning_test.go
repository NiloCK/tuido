package utils_test

import (
	"fmt"
	"os"
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

	want := tags[len(tags)-1]

	if got := utils.Version(); got != want {
		t.Errorf("Version() = %v, want %v", got, want)
	}
}
