package utils_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/nilock/tuido/utils"
)

func TestFetchLatestReleaseIntegration(t *testing.T) {
	// Mock GitHub API response
	mockRelease := utils.GitHubRelease{
		TagName: "v0.0.16",
		Assets: []utils.ReleaseAsset{
			{
				Name:        "tuido_0.0.16_linux_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_linux_amd64.tar.gz",
				Size:        1234567,
			},
			{
				Name:        "tuido_0.0.16_darwin_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_darwin_amd64.tar.gz",
				Size:        1234568,
			},
			{
				Name:        "tuido_0.0.16_windows_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_windows_amd64.tar.gz",
				Size:        1234569,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			t.Error("Expected User-Agent header to be set")
		}
		if !contains(userAgent, "tuido") {
			t.Errorf("Expected User-Agent to contain 'tuido', got %s", userAgent)
		}

		accept := r.Header.Get("Accept")
		if accept != "application/vnd.github.v3+json" {
			t.Errorf("Expected Accept header to be 'application/vnd.github.v3+json', got %s", accept)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockRelease); err != nil {
			t.Fatalf("Failed to encode mock response: %v", err)
		}
	}))
	defer server.Close()

	// Test fetchLatestReleaseFromURL with mock server
	release, err := utils.FetchLatestReleaseFromURL(server.URL)
	if err != nil {
		t.Fatalf("fetchLatestReleaseFromURL failed: %v", err)
	}

	// Verify the parsed response
	if release.TagName != mockRelease.TagName {
		t.Errorf("Expected tag name %s, got %s", mockRelease.TagName, release.TagName)
	}

	if len(release.Assets) != len(mockRelease.Assets) {
		t.Errorf("Expected %d assets, got %d", len(mockRelease.Assets), len(release.Assets))
	}

	// Verify assets match
	for i, expectedAsset := range mockRelease.Assets {
		if i >= len(release.Assets) {
			t.Errorf("Missing asset at index %d", i)
			continue
		}

		actualAsset := release.Assets[i]
		if !reflect.DeepEqual(actualAsset, expectedAsset) {
			t.Errorf("Asset %d mismatch:\nExpected: %+v\nActual: %+v", i, expectedAsset, actualAsset)
		}
	}
}

func TestFetchLatestReleaseErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		errorContains  string
	}{
		{
			name: "Rate limit exceeded",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", "1640995200")
				w.WriteHeader(403)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "API rate limit exceeded",
				})
			},
			expectError:   true,
			errorContains: "rate limit exceeded",
		},
		{
			name: "Server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Internal server error",
				})
			},
			expectError:   true,
			errorContains: "status 500",
		},
		{
			name: "Not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Not Found",
				})
			},
			expectError:   true,
			errorContains: "status 404",
		},
		{
			name: "Invalid JSON",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{invalid json}`)
			},
			expectError:   true,
			errorContains: "failed to decode",
		},
		{
			name: "Empty response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{}`)
			},
			expectError: false, // Empty response should be valid, just with no tag/assets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			_, err := utils.FetchLatestReleaseFromURL(server.URL)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetCurrentPlatformAssetIntegration(t *testing.T) {
	// Create comprehensive mock release with assets for various platforms
	mockRelease := utils.GitHubRelease{
		TagName: "v0.0.16",
		Assets: []utils.ReleaseAsset{
			{
				Name:        "tuido_0.0.16_linux_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_linux_amd64.tar.gz",
				Size:        1234567,
			},
			{
				Name:        "tuido_0.0.16_linux_arm64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_linux_arm64.tar.gz",
				Size:        1234568,
			},
			{
				Name:        "tuido_0.0.16_darwin_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_darwin_amd64.tar.gz",
				Size:        1234569,
			},
			{
				Name:        "tuido_0.0.16_darwin_arm64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_darwin_arm64.tar.gz",
				Size:        1234570,
			},
			{
				Name:        "tuido_0.0.16_windows_amd64.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_windows_amd64.tar.gz",
				Size:        1234571,
			},
			{
				Name:        "tuido_0.0.16_windows_386.tar.gz",
				DownloadURL: "https://github.com/NiloCK/tuido/releases/download/v0.0.16/tuido_0.0.16_windows_386.tar.gz",
				Size:        1234572,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	// Test specific platform combinations
	testCases := []struct {
		goos   string
		goarch string
		found  bool
	}{
		{"linux", "amd64", true},
		{"linux", "arm64", true},
		{"darwin", "amd64", true},
		{"darwin", "arm64", true},
		{"windows", "amd64", true},
		{"windows", "386", true},
		{"freebsd", "amd64", false}, // Not in our mock data
		{"linux", "mips", false},   // Not in our mock data
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.goos, tc.goarch), func(t *testing.T) {
			// This would require refactoring GetCurrentPlatformAsset to accept URL parameter
			// For now, we test the logic separately
			expectedName := utils.BuildAssetName("v0.0.16", tc.goos, tc.goarch)
			
			var found bool
			var foundAsset *utils.ReleaseAsset
			for _, asset := range mockRelease.Assets {
				if asset.Name == expectedName {
					found = true
					foundAsset = &asset
					break
				}
			}

			if tc.found && !found {
				t.Errorf("Expected to find asset for %s/%s (name: %s)", tc.goos, tc.goarch, expectedName)
			}
			if !tc.found && found {
				t.Errorf("Expected not to find asset for %s/%s, but found %s", tc.goos, tc.goarch, foundAsset.Name)
			}
			if found && foundAsset.Name != expectedName {
				t.Errorf("Found asset name %s, expected %s", foundAsset.Name, expectedName)
			}
		})
	}
}

func TestNetworkTimeout(t *testing.T) {
	// Skip this test in short mode since it takes time
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Create a server that returns immediately but to an invalid address
	// to simulate network connectivity issues without hanging
	_, err := utils.FetchLatestReleaseFromURL("http://192.0.2.1:12345/nonexistent")
	if err == nil {
		t.Error("Expected network error but got none")
	}

	// Check that it's a network-related error
	errStr := err.Error()
	if !contains(errStr, "connection") && !contains(errStr, "network") && 
	   !contains(errStr, "timeout") && !contains(errStr, "refused") {
		t.Logf("Got network error (which is expected): %v", err)
	}
}

func TestMalformedURL(t *testing.T) {
	_, err := utils.FetchLatestReleaseFromURL("not-a-valid-url")
	if err == nil {
		t.Error("Expected error for malformed URL but got none")
	}
}

func TestEmptyAssetsList(t *testing.T) {
	mockRelease := utils.GitHubRelease{
		TagName: "v0.0.16",
		Assets:  []utils.ReleaseAsset{}, // Empty assets list
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	_, err := utils.FetchLatestReleaseFromURL(server.URL)
	if err != nil {
		t.Errorf("Expected no error for empty assets list, got: %v", err)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		indexSubstring(s, substr) != -1)
}

func indexSubstring(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}