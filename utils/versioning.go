package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

const version = "v0.0.16"
const ReleaseURL = "https://github.com/NiloCK/tuido/releases/latest"
const GitHubAPIURL = "https://api.github.com/repos/NiloCK/tuido/releases/latest"

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

// ReleaseAsset represents a GitHub release asset
type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// GetCurrentPlatformAsset returns the appropriate asset for the current platform
func GetCurrentPlatformAsset() (*ReleaseAsset, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	expectedName := BuildAssetName(release.TagName, runtime.GOOS, runtime.GOARCH)

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no asset found for platform %s/%s (expected: %s)",
		runtime.GOOS, runtime.GOARCH, expectedName)
}

// fetchLatestRelease retrieves the latest release information from GitHub API
func fetchLatestRelease() (*GitHubRelease, error) {
	return FetchLatestReleaseFromURL(GitHubAPIURL)
}

// FetchLatestReleaseFromURL retrieves release information from a specific URL (testable version)
func FetchLatestReleaseFromURL(apiURL string) (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid rate limiting
	req.Header.Set("User-Agent", "tuido/"+version)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		resetTime := resp.Header.Get("X-RateLimit-Reset")
		return nil, fmt.Errorf("GitHub API rate limit exceeded, resets at %s", resetTime)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release response: %w", err)
	}

	return &release, nil
}

// BuildAssetName constructs the expected asset name for a given platform
func BuildAssetName(version, goos, goarch string) string {
	// Remove 'v' prefix from version if present
	cleanVersion := version
	if strings.HasPrefix(version, "v") {
		cleanVersion = version[1:]
	}

	// Handle architecture mappings
	arch := goarch
	switch goarch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	case "arm":
		arch = "armv6" // GoReleaser uses armv6 for arm
	}

	// Handle OS mappings and file extensions
	os := goos
	ext := ".tar.gz"
	switch goos {
	case "windows":
		ext = ".tar.gz" // Windows also uses tar.gz in GoReleaser
	case "darwin":
		os = "darwin"
	case "linux":
		os = "linux"
	}

	// Format: tuido_0.0.16_linux_amd64.tar.gz
	return fmt.Sprintf("tuido_%s_%s_%s%s", cleanVersion, os, arch, ext)
}

// ProgressCallback is called during download to report progress
type ProgressCallback func(downloaded, total int64)

// DownloadConfig contains configuration for downloading assets
type DownloadConfig struct {
	// Context for cancellation
	Context context.Context
	// Progress callback (optional)
	OnProgress ProgressCallback
	// Retry count for failed downloads
	RetryCount int
	// Timeout for each download attempt
	Timeout time.Duration
}

// DefaultDownloadConfig returns a reasonable default configuration
func DefaultDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		Context:    context.Background(),
		RetryCount: 3,
		Timeout:    5 * time.Minute,
	}
}

// DownloadAsset downloads an asset to the specified file path with progress tracking
func DownloadAsset(asset *ReleaseAsset, filePath string, config *DownloadConfig) error {
	if config == nil {
		config = DefaultDownloadConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.RetryCount; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			waitTime := time.Duration(attempt*attempt) * time.Second
			select {
			case <-time.After(waitTime):
			case <-config.Context.Done():
				return config.Context.Err()
			}
		}

		err := downloadAssetAttempt(asset, filePath, config)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on context cancellation or file system errors
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		// Don't retry on file creation errors - these won't be fixed by retrying
		if strings.Contains(err.Error(), "failed to create output file") {
			break
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", config.RetryCount+1, lastErr)
}

// downloadAssetAttempt performs a single download attempt
func downloadAssetAttempt(asset *ReleaseAsset, filePath string, config *DownloadConfig) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(config.Context, "GET", asset.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "tuido/"+version)

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create output file
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	progressContentLength := contentLength
	if progressContentLength <= 0 {
		progressContentLength = asset.Size // Fall back to asset size for progress only
	}

	// Download with progress tracking
	var downloaded int64
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		select {
		case <-config.Context.Done():
			return config.Context.Err()
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			downloaded += int64(written)

			// Call progress callback if provided
			if config.OnProgress != nil {
				config.OnProgress(downloaded, progressContentLength)
			}
		}

		if err == io.EOF {
			break // Download complete
		}
		if err != nil {
			return fmt.Errorf("download interrupted: %w", err)
		}
	}

	// Verify download size - use Content-Length if available, otherwise asset size
	expectedSize := contentLength
	if expectedSize <= 0 {
		expectedSize = asset.Size
	}

	if expectedSize > 0 && downloaded != expectedSize {
		return fmt.Errorf("download incomplete: got %d bytes, expected %d", downloaded, expectedSize)
	}

	return nil
}

// ValidateAssetIntegrity verifies the downloaded asset (placeholder for future checksum validation)
func ValidateAssetIntegrity(filePath string, expectedSize int64) error {
	// Verify file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	// Verify file size
	if expectedSize > 0 && info.Size() != expectedSize {
		return fmt.Errorf("file size mismatch: got %d bytes, expected %d", info.Size(), expectedSize)
	}

	// TODO: Implement checksum validation when checksums are available in releases
	return nil
}
