package utils

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const version = "v0.0.8"
const ReleaseURL = "https://github.com/NiloCK/tuido/releases/latest"

// Version returns the currently running version of the application.
func Version() string {
	return version
}

func getLatestRedirectURL() (string, error) {

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Stop after the first redirect
		},
	}

	resp, err := client.Get(ReleaseURL)
	if err != nil {
		// If the error is due to stopping redirects, extract the Location header
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Err == http.ErrUseLastResponse {
			if location, err := resp.Location(); err == nil {
				return location.String(), nil
			}
		}
		return "", err
	}
	defer resp.Body.Close()

	loc, err := resp.Location()
	if err != nil {
		return loc.String(), nil
	}

	split := strings.Split(loc.String(), "/")
	version := split[len(split)-1]
	if version != "" {
		return version, nil
	}

	return "", errors.New("no redirect")
}

func LatestVersion() string {
	// lookup latest version from github
	latest, err := getLatestRedirectURL()

	if err != nil {
		return fmt.Sprintf("Error getting latest version: %s", err)
	}

	return latest
}
