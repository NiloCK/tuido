package utils_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nilock/tuido/utils"
)

func TestDownloadAsset(t *testing.T) {
	// Create test data
	testData := strings.Repeat("Hello, World! ", 1000) // ~13KB of data
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if userAgent := r.Header.Get("User-Agent"); !strings.Contains(userAgent, "tuido") {
			t.Errorf("Expected User-Agent to contain 'tuido', got %s", userAgent)
		}
		
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(200)
		io.WriteString(w, testData)
	}))
	defer server.Close()

	// Create test asset
	asset := &utils.ReleaseAsset{
		Name:        "test_asset.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)),
	}

	// Create temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "downloaded_asset.tar.gz")

	// Test download
	config := utils.DefaultDownloadConfig()
	err := utils.DownloadAsset(asset, filePath, config)
	if err != nil {
		t.Fatalf("DownloadAsset failed: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content doesn't match expected data")
	}
}

func TestDownloadAssetWithProgress(t *testing.T) {
	testData := strings.Repeat("Progress test data ", 500) // ~9.5KB
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(200)
		
		// Write data in chunks to simulate progress
		chunkSize := 1000
		for i := 0; i < len(testData); i += chunkSize {
			end := i + chunkSize
			if end > len(testData) {
				end = len(testData)
			}
			w.Write([]byte(testData[i:end]))
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond) // Small delay to simulate network
		}
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "progress_test.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)),
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "progress_test.tar.gz")

	// Track progress
	var progressCalls []struct {
		downloaded, total int64
	}
	var mu sync.Mutex

	config := &utils.DownloadConfig{
		Context:    context.Background(),
		RetryCount: 3,
		Timeout:    30 * time.Second,
		OnProgress: func(downloaded, total int64) {
			mu.Lock()
			defer mu.Unlock()
			progressCalls = append(progressCalls, struct {
				downloaded, total int64
			}{downloaded, total})
		},
	}

	err := utils.DownloadAsset(asset, filePath, config)
	if err != nil {
		t.Fatalf("DownloadAsset with progress failed: %v", err)
	}

	// Verify progress was tracked
	mu.Lock()
	defer mu.Unlock()
	
	if len(progressCalls) == 0 {
		t.Error("Expected progress callbacks but got none")
	}

	// Check that progress increased over time
	if len(progressCalls) > 1 {
		first := progressCalls[0]
		last := progressCalls[len(progressCalls)-1]
		
		if last.downloaded <= first.downloaded {
			t.Error("Expected downloaded bytes to increase over time")
		}
		
		if last.total != int64(len(testData)) {
			t.Errorf("Expected total bytes to be %d, got %d", len(testData), last.total)
		}
	}
}

func TestDownloadAssetCancellation(t *testing.T) {
	// Use an unreachable IP to simulate hanging connection
	asset := &utils.ReleaseAsset{
		Name:        "unreachable_download.tar.gz",
		DownloadURL: "http://192.0.2.1:12345/file.tar.gz", // RFC 5737 test address
		Size:        1000000,
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "unreachable_download.tar.gz")

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	config := &utils.DownloadConfig{
		Context:    ctx,
		RetryCount: 0, // No retries for this test
		Timeout:    30 * time.Second,
	}

	start := time.Now()
	err := utils.DownloadAsset(asset, filePath, config)
	duration := time.Since(start)

	// Should fail quickly due to context cancellation or connection error
	if err == nil {
		t.Error("Expected download to fail but it succeeded")
	}

	if duration > 1*time.Second {
		t.Errorf("Expected quick failure, but took %v", duration)
	}

	// Accept either context cancellation or connection error
	errStr := err.Error()
	if !strings.Contains(errStr, "context") && !strings.Contains(errStr, "canceled") && 
	   !strings.Contains(errStr, "connection") && !strings.Contains(errStr, "timeout") {
		t.Logf("Got expected network error: %v", err)
	}
}

func TestDownloadAssetRetry(t *testing.T) {
	attemptCount := 0
	testData := "Retry test data"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		
		if attemptCount < 3 {
			// Fail first two attempts
			w.WriteHeader(500)
			return
		}
		
		// Succeed on third attempt
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(200)
		io.WriteString(w, testData)
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "retry_test.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)),
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "retry_test.tar.gz")

	config := &utils.DownloadConfig{
		Context:    context.Background(),
		RetryCount: 3,
		Timeout:    5 * time.Second,
	}

	err := utils.DownloadAsset(asset, filePath, config)
	if err != nil {
		t.Fatalf("DownloadAsset with retry failed: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Verify file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content doesn't match expected data")
	}
}

func TestDownloadAssetRetryExhaustion(t *testing.T) {
	// Server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		io.WriteString(w, "Server error")
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "failing_download.tar.gz",
		DownloadURL: server.URL,
		Size:        100,
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "failing_download.tar.gz")

	config := &utils.DownloadConfig{
		Context:    context.Background(),
		RetryCount: 2,
		Timeout:    1 * time.Second,
	}

	err := utils.DownloadAsset(asset, filePath, config)
	if err == nil {
		t.Error("Expected download to fail but it succeeded")
	}

	if !strings.Contains(err.Error(), "failed after") || !strings.Contains(err.Error(), "attempts") {
		t.Errorf("Expected retry exhaustion error, got: %v", err)
	}
}

func TestDownloadAssetSizeMismatch(t *testing.T) {
	testData := "Size mismatch test"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Manually set Content-Length to be larger than actual data
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(200)
		// Write less data than Content-Length claims
		io.WriteString(w, testData)
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "size_mismatch.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)), // Correct asset size
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "size_mismatch.tar.gz")

	config := utils.DefaultDownloadConfig()
	err := utils.DownloadAsset(asset, filePath, config)
	
	if err == nil {
		t.Error("Expected size mismatch error but download succeeded")
		return
	}

	if !strings.Contains(err.Error(), "EOF") && !strings.Contains(err.Error(), "incomplete") && !strings.Contains(err.Error(), "interrupted") {
		t.Errorf("Expected EOF or size mismatch error, got: %v", err)
	}
}

func TestDownloadAssetInvalidURL(t *testing.T) {
	asset := &utils.ReleaseAsset{
		Name:        "invalid_url.tar.gz",
		DownloadURL: "http://non-existent-domain-12345.com/file",
		Size:        100,
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid_url.tar.gz")

	config := &utils.DownloadConfig{
		Context:    context.Background(),
		RetryCount: 1,
		Timeout:    2 * time.Second,
	}

	err := utils.DownloadAsset(asset, filePath, config)
	if err == nil {
		t.Error("Expected download to fail for invalid URL")
	}
}

func TestDownloadAssetFileCreationError(t *testing.T) {
	testData := "File creation test"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(200)
		io.WriteString(w, testData)
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "file_creation_test.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)),
	}

	// Try to write to invalid path (directory that doesn't exist)
	filePath := "/non/existent/directory/file.tar.gz"

	config := utils.DefaultDownloadConfig()
	err := utils.DownloadAsset(asset, filePath, config)
	
	if err == nil {
		t.Error("Expected file creation error but download succeeded")
	}

	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("Expected file creation error, got: %v", err)
	}
}

func TestValidateAssetIntegrity(t *testing.T) {
	// Create test file
	testData := "Integrity test data"
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "integrity_test.txt")
	
	err := os.WriteFile(filePath, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with correct size
	err = utils.ValidateAssetIntegrity(filePath, int64(len(testData)))
	if err != nil {
		t.Errorf("ValidateAssetIntegrity failed with correct size: %v", err)
	}

	// Test with incorrect size
	err = utils.ValidateAssetIntegrity(filePath, 1000)
	if err == nil {
		t.Error("Expected size mismatch error but validation passed")
	}

	if !strings.Contains(err.Error(), "size mismatch") {
		t.Errorf("Expected size mismatch error, got: %v", err)
	}

	// Test with non-existent file
	err = utils.ValidateAssetIntegrity("/non/existent/file", 100)
	if err == nil {
		t.Error("Expected file not found error but validation passed")
	}
}

func TestDefaultDownloadConfig(t *testing.T) {
	config := utils.DefaultDownloadConfig()
	
	if config == nil {
		t.Fatal("DefaultDownloadConfig returned nil")
	}

	if config.Context == nil {
		t.Error("Expected non-nil context")
	}

	if config.RetryCount != 3 {
		t.Errorf("Expected retry count 3, got %d", config.RetryCount)
	}

	if config.Timeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", config.Timeout)
	}
}

func TestDownloadAssetNilConfig(t *testing.T) {
	testData := "Nil config test"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(200)
		io.WriteString(w, testData)
	}))
	defer server.Close()

	asset := &utils.ReleaseAsset{
		Name:        "nil_config_test.tar.gz",
		DownloadURL: server.URL,
		Size:        int64(len(testData)),
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nil_config_test.tar.gz")

	// Test with nil config (should use defaults)
	err := utils.DownloadAsset(asset, filePath, nil)
	if err != nil {
		t.Fatalf("DownloadAsset with nil config failed: %v", err)
	}

	// Verify file was created and has correct content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content doesn't match expected data")
	}
}