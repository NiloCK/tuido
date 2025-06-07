package utils_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nilock/tuido/utils"
)

func TestDefaultUpgradeConfig(t *testing.T) {
	config := utils.DefaultUpgradeConfig()
	
	if config == nil {
		t.Fatal("DefaultUpgradeConfig returned nil")
	}

	if config.Context == nil {
		t.Error("Expected non-nil context")
	}

	if config.DownloadConfig == nil {
		t.Error("Expected non-nil download config")
	}

	if !config.CreateBackup {
		t.Error("Expected CreateBackup to be true by default")
	}

	if config.BackupSuffix != ".backup" {
		t.Errorf("Expected backup suffix '.backup', got %s", config.BackupSuffix)
	}

	if config.AttemptRestart {
		t.Error("Expected AttemptRestart to be false by default")
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create source file
	srcPath := filepath.Join(tempDir, "source.txt")
	srcContent := "Hello, World! This is test content for copy file test."
	err := os.WriteFile(srcPath, []byte(srcContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tempDir, "destination.txt")
	err = utils.CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != srcContent {
		t.Errorf("Destination content doesn't match source")
	}

	// Verify permissions are preserved
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("File permissions not preserved: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyFileErrors(t *testing.T) {
	tempDir := t.TempDir()

	// Test copying non-existent file
	err := utils.CopyFile("/non/existent/file", filepath.Join(tempDir, "dst"))
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}

	// Test copying to invalid destination
	srcPath := filepath.Join(tempDir, "source.txt")
	err = os.WriteFile(srcPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	err = utils.CopyFile(srcPath, "/invalid/destination/path")
	if err == nil {
		t.Error("Expected error when copying to invalid destination")
	}
}

func TestBusyExecutableError(t *testing.T) {
	err := &utils.BusyExecutableError{
		CurrentPath: "/path/to/current",
		NewPath:     "/path/to/new",
		Message:     "test busy error",
	}

	if err.Error() != "test busy error" {
		t.Errorf("Expected error message 'test busy error', got %s", err.Error())
	}

	if !utils.IsBusyExecutableError(err) {
		t.Error("IsBusyExecutableError should return true for BusyExecutableError")
	}

	// Test with regular error
	regularErr := fmt.Errorf("regular error")
	if utils.IsBusyExecutableError(regularErr) {
		t.Error("IsBusyExecutableError should return false for regular error")
	}

	// Test GetBusyExecutableInfo
	currentPath, newPath, ok := utils.GetBusyExecutableInfo(err)
	if !ok {
		t.Error("GetBusyExecutableInfo should return ok=true for BusyExecutableError")
	}
	if currentPath != "/path/to/current" {
		t.Errorf("Expected current path '/path/to/current', got %s", currentPath)
	}
	if newPath != "/path/to/new" {
		t.Errorf("Expected new path '/path/to/new', got %s", newPath)
	}

	// Test with regular error
	_, _, ok = utils.GetBusyExecutableInfo(regularErr)
	if ok {
		t.Error("GetBusyExecutableInfo should return ok=false for regular error")
	}
}

func TestExtractExecutableFromArchive(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a proper tar.gz archive with executable
	archivePath := filepath.Join(tempDir, "mock_archive.tar.gz")
	mockContent := "mock executable content"
	
	// Create the tar.gz archive
	err := createMockTarGz(archivePath, mockContent)
	if err != nil {
		t.Fatalf("Failed to create mock archive: %v", err)
	}

	// Test extraction
	extractDir := filepath.Join(tempDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create extract directory: %v", err)
	}

	extractedPath, err := utils.ExtractExecutableFromArchive(archivePath, extractDir)
	if err != nil {
		t.Fatalf("ExtractExecutableFromArchive failed: %v", err)
	}

	// Verify extracted file
	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extractedContent) != mockContent {
		t.Errorf("Extracted content doesn't match original")
	}

	// Verify file name
	expectedName := "tuido"
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}
	if filepath.Base(extractedPath) != expectedName {
		t.Errorf("Expected extracted file name %s, got %s", expectedName, filepath.Base(extractedPath))
	}
}

// createMockTarGz creates a proper tar.gz archive containing the tuido executable
func createMockTarGz(archivePath, content string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Determine executable name for this platform
	execName := "tuido"
	if runtime.GOOS == "windows" {
		execName = "tuido.exe"
	}

	// Create tar header for the executable
	header := &tar.Header{
		Name: execName,
		Mode: 0755,
		Size: int64(len(content)),
	}

	// Write header
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	// Write file content
	_, err = tarWriter.Write([]byte(content))
	return err
}

func TestCleanupBackups(t *testing.T) {
	tempDir := t.TempDir()
	execPath := filepath.Join(tempDir, "tuido")
	
	// Create main executable
	err := os.WriteFile(execPath, []byte("main"), 0755)
	if err != nil {
		t.Fatalf("Failed to create main executable: %v", err)
	}

	// Create several backup files
	backupFiles := []string{
		"tuido.backup.1",
		"tuido.backup.2", 
		"tuido.backup.3",
		"tuido.backup.4",
		"tuido.backup.5",
	}

	for _, backup := range backupFiles {
		backupPath := filepath.Join(tempDir, backup)
		err := os.WriteFile(backupPath, []byte("backup"), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup file %s: %v", backup, err)
		}
	}

	// Test cleanup keeping 2 backups
	err = utils.CleanupBackups(execPath, 2)
	if err != nil {
		t.Fatalf("CleanupBackups failed: %v", err)
	}

	// Count remaining backup files
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	backupCount := 0
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".backup") {
			backupCount++
		}
	}

	// Should have kept 2 backups (or fewer if cleanup removed some)
	if backupCount > 2 {
		t.Errorf("Expected at most 2 backup files after cleanup, got %d", backupCount)
	}
}

func TestRestoreFromBackup(t *testing.T) {
	tempDir := t.TempDir()
	
	execPath := filepath.Join(tempDir, "tuido")
	backupPath := filepath.Join(tempDir, "tuido.backup")
	
	// Create original executable
	originalContent := "original executable content"
	err := os.WriteFile(execPath, []byte(originalContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}

	// Create backup
	backupContent := "backup executable content"
	err = os.WriteFile(backupPath, []byte(backupContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify the executable (simulate corruption)
	err = os.WriteFile(execPath, []byte("corrupted content"), 0755)
	if err != nil {
		t.Fatalf("Failed to modify executable: %v", err)
	}

	// Restore from backup
	err = utils.RestoreFromBackup(execPath, backupPath)
	if err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify restoration
	restoredContent, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("Failed to read restored executable: %v", err)
	}

	if string(restoredContent) != backupContent {
		t.Errorf("Restored content doesn't match backup content")
	}
}

func TestRestoreFromBackupErrors(t *testing.T) {
	tempDir := t.TempDir()
	
	execPath := filepath.Join(tempDir, "tuido")
	nonExistentBackup := filepath.Join(tempDir, "non-existent.backup")
	
	// Test restoring from non-existent backup
	err := utils.RestoreFromBackup(execPath, nonExistentBackup)
	if err == nil {
		t.Error("Expected error when restoring from non-existent backup")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestReplaceExecutableSimulation(t *testing.T) {
	// This test simulates the replacement logic without actually replacing
	// the test executable, since that would be dangerous
	
	tempDir := t.TempDir()
	
	// Create mock current executable
	currentPath := filepath.Join(tempDir, "current_tuido")
	if runtime.GOOS == "windows" {
		currentPath += ".exe"
	}
	
	currentContent := "current executable"
	err := os.WriteFile(currentPath, []byte(currentContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create current executable: %v", err)
	}

	// Create mock new executable
	newPath := filepath.Join(tempDir, "new_tuido")
	if runtime.GOOS == "windows" {
		newPath += ".exe"
	}
	
	newContent := "new executable"
	err = os.WriteFile(newPath, []byte(newContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create new executable: %v", err)
	}

	// Test replacement with backup
	config := &utils.UpgradeConfig{
		Context:      context.Background(),
		CreateBackup: true,
		BackupSuffix: ".backup",
	}

	err = utils.ReplaceExecutable(currentPath, newPath, config)
	if err != nil {
		t.Fatalf("ReplaceExecutable failed: %v", err)
	}

	// Verify replacement
	replacedContent, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("Failed to read replaced executable: %v", err)
	}

	if string(replacedContent) != newContent {
		t.Errorf("Replaced content doesn't match new content")
	}

	// Verify backup was created
	backupPath := currentPath + config.BackupSuffix
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != currentContent {
		t.Errorf("Backup content doesn't match original content")
	}
}

func TestReplaceExecutableWithoutBackup(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create mock files
	currentPath := filepath.Join(tempDir, "current_tuido")
	newPath := filepath.Join(tempDir, "new_tuido")
	
	err := os.WriteFile(currentPath, []byte("current"), 0755)
	if err != nil {
		t.Fatalf("Failed to create current executable: %v", err)
	}

	err = os.WriteFile(newPath, []byte("new"), 0755)
	if err != nil {
		t.Fatalf("Failed to create new executable: %v", err)
	}

	// Test replacement without backup
	config := &utils.UpgradeConfig{
		Context:      context.Background(),
		CreateBackup: false,
	}

	err = utils.ReplaceExecutable(currentPath, newPath, config)
	if err != nil {
		t.Fatalf("ReplaceExecutable failed: %v", err)
	}

	// Verify no backup was created
	backupPath := currentPath + ".backup"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file was created when CreateBackup was false")
	}
}

func TestUpgradeResultStructure(t *testing.T) {
	result := &utils.UpgradeResult{
		Success:           true,
		NewExecutablePath: "/path/to/new/exe",
		BackupPath:        "/path/to/backup",
		RestartRequired:   true,
		Error:             nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.NewExecutablePath != "/path/to/new/exe" {
		t.Errorf("Expected NewExecutablePath '/path/to/new/exe', got %s", result.NewExecutablePath)
	}

	if result.BackupPath != "/path/to/backup" {
		t.Errorf("Expected BackupPath '/path/to/backup', got %s", result.BackupPath)
	}

	if !result.RestartRequired {
		t.Error("Expected RestartRequired to be true")
	}

	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
}

func TestPlatformSpecificBehavior(t *testing.T) {
	// Test that platform-specific functions are accessible
	// This mainly tests compilation and basic structure
	
	tempDir := t.TempDir()
	currentPath := filepath.Join(tempDir, "test_exe")
	newPath := filepath.Join(tempDir, "new_exe")
	
	// Add platform-specific extension
	if runtime.GOOS == "windows" {
		currentPath += ".exe"
		newPath += ".exe"
	}

	// Create test files
	err := os.WriteFile(currentPath, []byte("current"), 0755)
	if err != nil {
		t.Fatalf("Failed to create current file: %v", err)
	}

	err = os.WriteFile(newPath, []byte("new"), 0755)
	if err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	config := utils.DefaultUpgradeConfig()
	
	// This should work on both platforms
	err = utils.ReplaceExecutable(currentPath, newPath, config)
	if err != nil {
		// Some errors are expected (like busy executable), but we mainly
		// want to ensure the platform-specific code paths are reachable
		t.Logf("ReplaceExecutable returned error (may be expected): %v", err)
	}
}

func TestPermissionPreservation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission preservation test skipped on Windows")
	}

	tempDir := t.TempDir()
	
	// Create source file with specific permissions
	srcPath := filepath.Join(tempDir, "source")
	err := os.WriteFile(srcPath, []byte("content"), 0751) // rwxr-x--x
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tempDir, "destination")
	err = utils.CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Check permissions are preserved
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Failed to stat source: %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permissions not preserved: src=%v, dst=%v", srcInfo.Mode(), dstInfo.Mode())
	}
}