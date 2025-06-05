package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/nilock/tuido/utils"
)

func main() {
	fmt.Printf("Tuido Upgrade Download Demo\n")
	fmt.Printf("===========================\n\n")

	// Show current version and platform info
	fmt.Printf("Current version: %s\n", utils.Version())
	fmt.Printf("Platform: %s/%s\n\n", runtime.GOOS, runtime.GOARCH)

	// Test getting the latest version (existing functionality)
	fmt.Printf("Checking for latest version...\n")
	latest := utils.LatestVersion()
	fmt.Printf("Latest version: %s\n\n", latest)

	// Test the new asset fetching functionality
	fmt.Printf("Fetching current platform asset...\n")
	asset, err := utils.GetCurrentPlatformAsset()
	if err != nil {
		log.Printf("Error fetching asset: %v\n", err)
		
		// Show what the expected asset name would be
		expectedName := utils.BuildAssetName(latest, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Expected asset name would be: %s\n", expectedName)
		return
	}

	// Display asset information
	fmt.Printf("✓ Found asset for current platform!\n")
	fmt.Printf("  Name: %s\n", asset.Name)
	fmt.Printf("  Size: %d bytes (%.2f MB)\n", asset.Size, float64(asset.Size)/(1024*1024))
	fmt.Printf("  Download URL: %s\n", asset.DownloadURL)

	// Test download functionality
	fmt.Printf("\n🔄 Testing download functionality...\n")
	tempDir, err := os.MkdirTemp("", "tuido-download-test-*")
	if err != nil {
		log.Printf("Failed to create temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	downloadPath := filepath.Join(tempDir, asset.Name)
	
	// Create download config with progress tracking
	config := &utils.DownloadConfig{
		Context:    context.Background(),
		RetryCount: 3,
		Timeout:    5 * time.Minute,
		OnProgress: func(downloaded, total int64) {
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				mb_downloaded := float64(downloaded) / (1024 * 1024)
				mb_total := float64(total) / (1024 * 1024)
				fmt.Printf("\r  Progress: %.1f%% (%.2f/%.2f MB)", percent, mb_downloaded, mb_total)
			}
		},
	}

	start := time.Now()
	err = utils.DownloadAsset(asset, downloadPath, config)
	duration := time.Since(start)
	
	fmt.Printf("\n") // New line after progress
	
	if err != nil {
		log.Printf("Download failed: %v\n", err)
		return
	}

	// Verify download
	fmt.Printf("✅ Download completed in %v\n", duration)
	
	fileInfo, err := os.Stat(downloadPath)
	if err != nil {
		log.Printf("Failed to stat downloaded file: %v\n", err)
		return
	}
	
	fmt.Printf("📁 Downloaded file: %s\n", downloadPath)
	fmt.Printf("📊 File size: %d bytes\n", fileInfo.Size())
	
	// Test asset integrity validation
	fmt.Printf("\n🔍 Validating asset integrity...\n")
	err = utils.ValidateAssetIntegrity(downloadPath, asset.Size)
	if err != nil {
		log.Printf("Integrity validation failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Asset integrity validated successfully\n")

	// Test asset name building for various platforms
	fmt.Printf("\nAsset names for different platforms:\n")
	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
		{"windows", "386"},
	}

	for _, p := range platforms {
		name := utils.BuildAssetName(latest, p.goos, p.goarch)
		fmt.Printf("  %s/%s: %s\n", p.goos, p.goarch, name)
	}

	fmt.Printf("\n🎉 Demo completed successfully!\n")
	fmt.Printf("   Downloaded asset is available at: %s\n", downloadPath)
}