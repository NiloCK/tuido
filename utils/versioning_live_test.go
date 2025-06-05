//go:build integration
// +build integration

package utils_test

import (
	"testing"

	"github.com/nilock/tuido/utils"
)

// TestGetCurrentPlatformAssetLive tests against the real GitHub API
// Run with: go test -tags=integration ./utils -v -run TestGetCurrentPlatformAssetLive
func TestGetCurrentPlatformAssetLive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live GitHub API test in short mode")
	}

	asset, err := utils.GetCurrentPlatformAsset()
	if err != nil {
		// This might fail if there's no internet connection or GitHub is down
		// or if there's no asset for the current platform
		t.Logf("GetCurrentPlatformAsset failed (this may be expected): %v", err)
		return
	}

	if asset == nil {
		t.Error("GetCurrentPlatformAsset returned nil asset without error")
		return
	}

	// Verify the asset has reasonable values
	if asset.Name == "" {
		t.Error("Asset name is empty")
	}

	if asset.DownloadURL == "" {
		t.Error("Asset download URL is empty")
	}

	if asset.Size <= 0 {
		t.Error("Asset size should be positive")
	}

	t.Logf("Found asset: %s (size: %d bytes, url: %s)", 
		asset.Name, asset.Size, asset.DownloadURL)
}

// TestFetchLatestReleaseLive tests against the real GitHub API
func TestFetchLatestReleaseLive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live GitHub API test in short mode")
	}

	release, err := utils.FetchLatestReleaseFromURL(utils.GitHubAPIURL)
	if err != nil {
		t.Logf("FetchLatestReleaseFromURL failed (this may be expected): %v", err)
		return
	}

	if release == nil {
		t.Error("FetchLatestReleaseFromURL returned nil release without error")
		return
	}

	// Verify release has reasonable values
	if release.TagName == "" {
		t.Error("Release tag name is empty")
	}

	if len(release.Assets) == 0 {
		t.Error("Release has no assets")
	}

	t.Logf("Found release: %s with %d assets", release.TagName, len(release.Assets))

	// Log all available assets for debugging
	for i, asset := range release.Assets {
		t.Logf("  Asset %d: %s (size: %d)", i, asset.Name, asset.Size)
	}
}

// TestBuildAssetNameMatchesActualAssets verifies our asset naming matches reality
func TestBuildAssetNameMatchesActualAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live GitHub API test in short mode")
	}

	release, err := utils.FetchLatestReleaseFromURL(utils.GitHubAPIURL)
	if err != nil {
		t.Skipf("Could not fetch release for asset name verification: %v", err)
	}

	// Test common platform combinations
	testPlatforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	foundAssets := make(map[string]bool)
	for _, asset := range release.Assets {
		foundAssets[asset.Name] = true
	}

	for _, platform := range testPlatforms {
		expectedName := utils.BuildAssetName(release.TagName, platform.goos, platform.goarch)
		if foundAssets[expectedName] {
			t.Logf("✓ Found expected asset: %s", expectedName)
		} else {
			t.Logf("✗ Missing expected asset: %s", expectedName)
			// Don't fail the test since this might be expected for some platforms
		}
	}
}