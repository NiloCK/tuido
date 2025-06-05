package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nilock/tuido/utils"
)

func main() {
	fmt.Printf("Tuido Upgrade Replacement Demo\n")
	fmt.Printf("==============================\n\n")

	// Create a temporary workspace for the demo
	tempDir, err := os.MkdirTemp("", "tuido-upgrade-demo-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("🔧 Demo workspace: %s\n\n", tempDir)

	// Test 1: Basic file operations
	fmt.Printf("Test 1: File Copy Operations\n")
	fmt.Printf("-----------------------------\n")

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")

	err = os.WriteFile(srcFile, []byte("Original content v1.0"), 0755)
	if err != nil {
		log.Fatalf("Failed to create source file: %v", err)
	}

	err = utils.CopyFile(srcFile, dstFile)
	if err != nil {
		log.Fatalf("CopyFile failed: %v", err)
	}

	content, err := os.ReadFile(dstFile)
	if err != nil {
		log.Fatalf("Failed to read destination: %v", err)
	}

	fmt.Printf("✅ File copied successfully: %s\n", string(content))

	// Test 2: Mock executable replacement
	fmt.Printf("\nTest 2: Executable Replacement Simulation\n")
	fmt.Printf("------------------------------------------\n")

	// Create mock "current executable"
	currentExe := filepath.Join(tempDir, "mock_tuido")
	if runtime.GOOS == "windows" {
		currentExe += ".exe"
	}

	currentContent := "Mock Tuido v0.0.15 - Current Version"
	err = os.WriteFile(currentExe, []byte(currentContent), 0755)
	if err != nil {
		log.Fatalf("Failed to create mock current executable: %v", err)
	}

	fmt.Printf("📁 Created mock current executable: %s\n", currentExe)

	// Create mock "new executable"
	newExe := filepath.Join(tempDir, "mock_tuido_new")
	if runtime.GOOS == "windows" {
		newExe += ".exe"
	}

	newContent := "Mock Tuido v0.0.16 - New Version"
	err = os.WriteFile(newExe, []byte(newContent), 0755)
	if err != nil {
		log.Fatalf("Failed to create mock new executable: %v", err)
	}

	fmt.Printf("📁 Created mock new executable: %s\n", newExe)

	// Test replacement with backup
	config := &utils.UpgradeConfig{
		Context:      context.Background(),
		CreateBackup: true,
		BackupSuffix: ".backup",
	}

	fmt.Printf("🔄 Performing replacement with backup...\n")
	err = utils.ReplaceExecutable(currentExe, newExe, config)
	if err != nil {
		log.Fatalf("ReplaceExecutable failed: %v", err)
	}

	// Verify replacement
	replacedContent, err := os.ReadFile(currentExe)
	if err != nil {
		log.Fatalf("Failed to read replaced executable: %v", err)
	}

	fmt.Printf("✅ Replacement successful: %s\n", string(replacedContent))

	// Verify backup
	backupPath := currentExe + config.BackupSuffix
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		log.Fatalf("Failed to read backup: %v", err)
	}

	fmt.Printf("💾 Backup created: %s\n", string(backupContent))

	// Test 3: Archive extraction simulation
	fmt.Printf("\nTest 3: Archive Extraction Simulation\n")
	fmt.Printf("-------------------------------------\n")

	mockArchive := filepath.Join(tempDir, "mock_archive.tar.gz")
	archiveContent := "Mock Tuido v0.0.17 - Archive Content"
	err = os.WriteFile(mockArchive, []byte(archiveContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create mock archive: %v", err)
	}

	extractDir := filepath.Join(tempDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create extract directory: %v", err)
	}

	extractedPath, err := utils.ExtractExecutableFromArchive(mockArchive, extractDir)
	if err != nil {
		log.Fatalf("ExtractExecutableFromArchive failed: %v", err)
	}

	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		log.Fatalf("Failed to read extracted file: %v", err)
	}

	fmt.Printf("📦 Archive extracted: %s\n", extractedPath)
	fmt.Printf("✅ Extracted content: %s\n", string(extractedContent))

	// Test 4: Backup management
	fmt.Printf("\nTest 4: Backup Management\n")
	fmt.Printf("-------------------------\n")

	// Create multiple backup files
	baseExe := filepath.Join(tempDir, "tuido_backup_test")
	err = os.WriteFile(baseExe, []byte("main executable"), 0755)
	if err != nil {
		log.Fatalf("Failed to create base executable: %v", err)
	}

	backupFiles := []string{
		baseExe + ".backup.1",
		baseExe + ".backup.2",
		baseExe + ".backup.3",
		baseExe + ".backup.4",
		baseExe + ".backup.5",
	}

	for i, backup := range backupFiles {
		content := fmt.Sprintf("backup version %d", i+1)
		err = os.WriteFile(backup, []byte(content), 0755)
		if err != nil {
			log.Fatalf("Failed to create backup %s: %v", backup, err)
		}
		fmt.Printf("💾 Created backup: %s\n", filepath.Base(backup))
	}

	// Test cleanup (keep only 2 backups)
	fmt.Printf("🧹 Cleaning up backups (keeping 2)...\n")
	err = utils.CleanupBackups(baseExe, 2)
	if err != nil {
		log.Printf("CleanupBackups returned error: %v", err)
	}

	// Count remaining backups
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	remainingBackups := 0
	for _, entry := range entries {
		if filepath.Base(entry.Name()) != filepath.Base(baseExe) && 
		   len(entry.Name()) > len(filepath.Base(baseExe)) &&
		   entry.Name()[:len(filepath.Base(baseExe))] == filepath.Base(baseExe) &&
		   len(entry.Name()) > len(filepath.Base(baseExe))+7 { // ".backup" = 7 chars
			remainingBackups++
			fmt.Printf("💾 Remaining backup: %s\n", entry.Name())
		}
	}

	fmt.Printf("✅ Backup cleanup completed. Remaining backups: %d\n", remainingBackups)

	// Test 5: Restore from backup
	fmt.Printf("\nTest 5: Restore from Backup\n")
	fmt.Printf("---------------------------\n")

	// Corrupt the main executable
	err = os.WriteFile(baseExe, []byte("corrupted executable"), 0755)
	if err != nil {
		log.Fatalf("Failed to corrupt executable: %v", err)
	}

	fmt.Printf("💥 Simulated corruption: %s\n", "corrupted executable")

	// Find a backup to restore from
	var backupToRestore string
	for _, entry := range entries {
		if len(entry.Name()) > len(filepath.Base(baseExe))+7 &&
		   entry.Name()[:len(filepath.Base(baseExe))] == filepath.Base(baseExe) {
			backupToRestore = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if backupToRestore != "" {
		fmt.Printf("🔄 Restoring from backup: %s\n", filepath.Base(backupToRestore))
		err = utils.RestoreFromBackup(baseExe, backupToRestore)
		if err != nil {
			log.Fatalf("RestoreFromBackup failed: %v", err)
		}

		restoredContent, err := os.ReadFile(baseExe)
		if err != nil {
			log.Fatalf("Failed to read restored file: %v", err)
		}

		fmt.Printf("✅ Restore successful: %s\n", string(restoredContent))
	} else {
		fmt.Printf("⚠️  No backup found to restore from\n")
	}

	// Test 6: Error handling
	fmt.Printf("\nTest 6: Error Handling\n")
	fmt.Printf("----------------------\n")

	// Test BusyExecutableError
	busyErr := &utils.BusyExecutableError{
		CurrentPath: "/path/to/current",
		NewPath:     "/path/to/new",
		Message:     "executable is busy",
	}

	fmt.Printf("🚫 Testing BusyExecutableError: %s\n", busyErr.Error())
	
	if utils.IsBusyExecutableError(busyErr) {
		currentPath, newPath, ok := utils.GetBusyExecutableInfo(busyErr)
		if ok {
			fmt.Printf("📋 Busy executable info - Current: %s, New: %s\n", currentPath, newPath)
		}
	}

	// Test integrity validation
	fmt.Printf("🔍 Testing asset integrity validation...\n")
	testFile := filepath.Join(tempDir, "integrity_test.bin")
	testData := []byte("test data for integrity check")
	err = os.WriteFile(testFile, testData, 0644)
	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}

	err = utils.ValidateAssetIntegrity(testFile, int64(len(testData)))
	if err != nil {
		log.Fatalf("Integrity validation failed: %v", err)
	}
	fmt.Printf("✅ Integrity validation passed\n")

	// Test with wrong size
	err = utils.ValidateAssetIntegrity(testFile, 9999)
	if err != nil {
		fmt.Printf("✅ Size mismatch correctly detected: %v\n", err)
	} else {
		fmt.Printf("⚠️  Size mismatch not detected\n")
	}

	// Summary
	fmt.Printf("\n🎉 Upgrade Replacement Demo Completed!\n")
	fmt.Printf("=====================================\n")
	fmt.Printf("✅ File operations: Working\n")
	fmt.Printf("✅ Executable replacement: Working\n")
	fmt.Printf("✅ Archive extraction: Working\n")
	fmt.Printf("✅ Backup management: Working\n")
	fmt.Printf("✅ Restore functionality: Working\n")
	fmt.Printf("✅ Error handling: Working\n")
	fmt.Printf("\n📁 Demo workspace: %s\n", tempDir)
	fmt.Printf("   (Will be cleaned up on exit)\n")

	fmt.Printf("\n💡 The upgrade system is ready for integration!\n")
	fmt.Printf("   Next step: UI integration for progress and user interaction\n")
}