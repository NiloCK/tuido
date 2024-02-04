package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// walkRepo walks through the repository directory, applying .gitignore rules.
func WalkRepo(root string, walkFn filepath.WalkFunc) error {
	var ps []gitignore.Pattern
	domain := []string{}

	// Recursive function to walk through directories
	var walk func(string, []string) error
	walk = func(path string, domain []string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		files, err := f.Readdir(-1)
		if err != nil {
			return err
		}

		for _, file := range files {
			filePath := filepath.Join(path, file.Name())

			// Check if the file is a .gitignore file
			if file.Name() == ".gitignore" {
				filePatterns, err := parseFilePatterns(filePath, domain)

				if err != nil {
					return err
				}

				ps = append(ps, filePatterns...)
			}

			matcher := gitignore.NewMatcher(ps)
			if !matcher.Match(domain, file.IsDir()) {
				err := walkFn(filePath, file, nil)
				if err != nil {
					if err == filepath.SkipDir && file.IsDir() {
						continue
					} else {
						return err
					}
				}

				if file.IsDir() {
					newDomain := append(domain, file.Name())
					err := walk(filePath, newDomain)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	return walk(root, domain)
}

// parseFilePatterns parses the .gitignore file and returns a list of gitignore.Patterns.
func parseFilePatterns(path string, domain []string) ([]gitignore.Pattern, error) {
	if !strings.HasSuffix(path, ".gitignore") {
		return nil, fmt.Errorf("file %s is not a .gitignore file", path)
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	filePatterns := []gitignore.Pattern{}

	// Split the contents of the .gitignore file into rawPatterns
	rawPatterns := strings.Split(string(fileBytes), "\n")
	for _, rawPattern := range rawPatterns {
		// Ignore empty lines and comments
		if rawPattern == "" || strings.HasPrefix(rawPattern, "#") {
			continue
		}

		filePatterns = append(filePatterns, gitignore.ParsePattern(rawPattern, domain))
	}
	return filePatterns, nil
}
