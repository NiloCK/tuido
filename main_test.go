package main

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/nilock/tuido/utils"
)

func TestShowVersionInfo(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersionInfo()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Test that output contains expected components
	expectedVersion := utils.Version()
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	expectedAsset := utils.BuildAssetName(expectedVersion, runtime.GOOS, runtime.GOARCH)

	if !strings.Contains(output, expectedVersion) {
		t.Errorf("Output should contain version %s, got: %s", expectedVersion, output)
	}

	if !strings.Contains(output, expectedPlatform) {
		t.Errorf("Output should contain platform %s, got: %s", expectedPlatform, output)
	}

	if !strings.Contains(output, expectedAsset) {
		t.Errorf("Output should contain asset name %s, got: %s", expectedAsset, output)
	}

	// Test that output has the expected structure
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines of output, got %d: %s", len(lines), output)
	}

	// Test first line format
	if !strings.HasPrefix(lines[0], "tuido ") {
		t.Errorf("First line should start with 'tuido ', got: %s", lines[0])
	}

	// Test second line format
	if !strings.HasPrefix(lines[1], "Platform: ") {
		t.Errorf("Second line should start with 'Platform: ', got: %s", lines[1])
	}

	// Test third line format
	if !strings.HasPrefix(lines[2], "Asset: ") {
		t.Errorf("Third line should start with 'Asset: ', got: %s", lines[2])
	}

	// Test fourth line format (Available or error message)
	if len(lines) >= 4 {
		if !strings.HasPrefix(lines[3], "Available: ") {
			t.Errorf("Fourth line should start with 'Available: ', got: %s", lines[3])
		}
	}
}

func TestVersionInfoContainsExpectedAssetName(t *testing.T) {
	version := utils.Version()
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	expectedAsset := utils.BuildAssetName(version, goos, goarch)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersionInfo()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, expectedAsset) {
		t.Errorf("Version info should contain asset name %s, but output was: %s", expectedAsset, output)
	}
}

func TestVersionInfoResilience(t *testing.T) {
	// This test ensures that even if network calls fail,
	// the basic version info is still displayed
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersionInfo()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Even if network fails, we should still get basic info
	if !strings.Contains(output, "tuido") {
		t.Error("Output should always contain 'tuido' even on network failure")
	}

	if !strings.Contains(output, "Platform:") {
		t.Error("Output should always contain 'Platform:' even on network failure")
	}

	if !strings.Contains(output, "Asset:") {
		t.Error("Output should always contain 'Asset:' even on network failure")
	}

	// Should either show "Available:" with asset info or error message
	hasAvailable := strings.Contains(output, "Available:")
	if !hasAvailable {
		t.Error("Output should contain 'Available:' line with either asset info or error message")
	}
}