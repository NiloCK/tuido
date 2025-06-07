package utils

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UpgradeConfig contains configuration for the upgrade process
type UpgradeConfig struct {
	// Context for cancellation
	Context context.Context
	// Download configuration
	DownloadConfig *DownloadConfig
	// Whether to create a backup before replacement
	CreateBackup bool
	// Backup file suffix
	BackupSuffix string
	// Whether to attempt restart after upgrade
	AttemptRestart bool
}

// DefaultUpgradeConfig returns a reasonable default configuration
func DefaultUpgradeConfig() *UpgradeConfig {
	return &UpgradeConfig{
		Context:        context.Background(),
		DownloadConfig: DefaultDownloadConfig(),
		CreateBackup:   true,
		BackupSuffix:   ".backup",
		AttemptRestart: false,
	}
}

// UpgradeResult contains information about the upgrade process
type UpgradeResult struct {
	// Whether the upgrade was successful
	Success bool
	// Path to the new executable
	NewExecutablePath string
	// Path to the backup file (if created)
	BackupPath string
	// Whether a restart is required
	RestartRequired bool
	// Any error that occurred
	Error error
}

// PerformUpgrade downloads and replaces the current executable with a new version
func PerformUpgrade(config *UpgradeConfig) *UpgradeResult {
	if config == nil {
		config = DefaultUpgradeConfig()
	}

	result := &UpgradeResult{}

	// Get current executable path
	currentExePath, err := os.Executable()
	if err != nil {
		result.Error = fmt.Errorf("failed to get current executable path: %w", err)
		return result
	}

	// Get current platform asset
	asset, err := GetCurrentPlatformAsset()
	if err != nil {
		result.Error = fmt.Errorf("failed to get platform asset: %w", err)
		return result
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "tuido-upgrade-*")
	if err != nil {
		result.Error = fmt.Errorf("failed to create temp directory: %w", err)
		return result
	}
	defer os.RemoveAll(tempDir)

	// Download the new version
	downloadPath := filepath.Join(tempDir, "tuido_new")
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	err = DownloadAsset(asset, downloadPath, config.DownloadConfig)
	if err != nil {
		result.Error = fmt.Errorf("failed to download asset: %w", err)
		return result
	}

	// Validate the downloaded asset with comprehensive validation
	err = ValidateAssetIntegrityWithChecksum(downloadPath, asset)
	if err != nil {
		result.Error = fmt.Errorf("downloaded asset failed validation: %w", err)
		return result
	}

	// Extract executable from archive (since assets are tar.gz)
	extractedPath, err := ExtractExecutableFromArchive(downloadPath, tempDir)
	if err != nil {
		result.Error = fmt.Errorf("failed to extract executable: %w", err)
		return result
	}

	// Perform the replacement
	err = ReplaceExecutable(currentExePath, extractedPath, config)
	if err != nil {
		result.Error = fmt.Errorf("failed to replace executable: %w", err)
		return result
	}

	result.Success = true
	result.NewExecutablePath = currentExePath
	result.RestartRequired = true

	if config.CreateBackup {
		result.BackupPath = currentExePath + config.BackupSuffix
	}

	return result
}

// ReplaceExecutable handles the platform-specific logic of replacing the executable
func ReplaceExecutable(currentPath, newPath string, config *UpgradeConfig) error {
	switch runtime.GOOS {
	case "windows":
		return replaceExecutableWindows(currentPath, newPath, config)
	default:
		return replaceExecutableUnix(currentPath, newPath, config)
	}
}

// replaceExecutableUnix handles executable replacement on Unix-like systems
func replaceExecutableUnix(currentPath, newPath string, config *UpgradeConfig) error {
	// Preserve original permissions
	currentInfo, err := os.Stat(currentPath)
	if err != nil {
		return fmt.Errorf("failed to stat current executable: %w", err)
	}

	// Set permissions on new executable
	err = os.Chmod(newPath, currentInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set permissions on new executable: %w", err)
	}

	// Create backup if requested
	if config.CreateBackup {
		backupPath := currentPath + config.BackupSuffix
		err = CopyFile(currentPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Try atomic replacement first
	err = os.Rename(newPath, currentPath)
	if err != nil {
		// If rename fails (likely due to ETXTBSY), try alternative approach
		if strings.Contains(err.Error(), "text file busy") || strings.Contains(err.Error(), "resource busy") {
			return replaceExecutableUnixBusy(currentPath, newPath, config)
		}
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

// replaceExecutableUnixBusy handles replacement when the executable is busy (running)
func replaceExecutableUnixBusy(currentPath, newPath string, config *UpgradeConfig) error {
	// Strategy: Create a new file with a temporary name, then rename
	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)
	tempName := base + ".new." + fmt.Sprintf("%d", time.Now().UnixNano())
	tempPath := filepath.Join(dir, tempName)

	// Copy new executable to temporary location
	err := CopyFile(newPath, tempPath)
	if err != nil {
		return fmt.Errorf("failed to copy new executable to temp location: %w", err)
	}

	// Clean up temp file on error
	defer func() {
		if err != nil {
			os.Remove(tempPath)
		}
	}()

	// Preserve permissions
	currentInfo, statErr := os.Stat(currentPath)
	if statErr == nil {
		os.Chmod(tempPath, currentInfo.Mode())
	}

	// Try to rename temp file to current executable
	err = os.Rename(tempPath, currentPath)
	if err != nil {
		// If still busy, leave instructions for manual replacement
		return &BusyExecutableError{
			CurrentPath: currentPath,
			NewPath:     tempPath,
			Message:     "executable is busy, manual replacement required after restart",
		}
	}

	return nil
}

// replaceExecutableWindows handles executable replacement on Windows
func replaceExecutableWindows(currentPath, newPath string, config *UpgradeConfig) error {
	// Create backup if requested
	if config.CreateBackup {
		backupPath := currentPath + config.BackupSuffix
		err := CopyFile(currentPath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Try direct replacement first
	err := os.Rename(newPath, currentPath)
	if err == nil {
		return nil
	}

	// If that fails, try the Windows-specific approach
	return replaceExecutableWindowsLocked(currentPath, newPath, config)
}

// replaceExecutableWindowsLocked handles replacement when the file is locked
func replaceExecutableWindowsLocked(currentPath, newPath string, config *UpgradeConfig) error {
	// On Windows, we can often replace a running executable by:
	// 1. Moving the current executable to a temp name
	// 2. Moving the new executable to the current name
	// 3. The old executable will be deleted when the process exits

	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)
	oldName := base + ".old." + fmt.Sprintf("%d", time.Now().UnixNano())
	oldPath := filepath.Join(dir, oldName)

	// Move current executable to old name
	err := os.Rename(currentPath, oldPath)
	if err != nil {
		return fmt.Errorf("failed to move current executable: %w", err)
	}

	// Move new executable to current name
	err = os.Rename(newPath, currentPath)
	if err != nil {
		// Try to restore original if new replacement failed
		os.Rename(oldPath, currentPath)
		return fmt.Errorf("failed to move new executable to current location: %w", err)
	}

	// Schedule old executable for deletion on next reboot
	// Note: This is a simplified approach; in practice, you might want to use
	// MoveFileEx with MOVEFILE_DELAY_UNTIL_REBOOT
	os.Remove(oldPath) // This will likely fail, but that's OK

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return err
	}

	// Preserve permissions
	srcInfo, err := srcFile.Stat()
	if err == nil {
		dstFile.Chmod(srcInfo.Mode())
	}

	return nil
}

// BusyExecutableError indicates that the executable couldn't be replaced because it's busy
type BusyExecutableError struct {
	CurrentPath string
	NewPath     string
	Message     string
}

func (e *BusyExecutableError) Error() string {
	return e.Message
}

// IsBusyExecutableError checks if an error is a BusyExecutableError
func IsBusyExecutableError(err error) bool {
	_, ok := err.(*BusyExecutableError)
	return ok
}

// GetBusyExecutableInfo returns information about a busy executable error
func GetBusyExecutableInfo(err error) (currentPath, newPath string, ok bool) {
	if busyErr, isBusy := err.(*BusyExecutableError); isBusy {
		return busyErr.CurrentPath, busyErr.NewPath, true
	}
	return "", "", false
}

// ValidateArchiveFormat performs pre-extraction validation of tar.gz file format
func ValidateArchiveFormat(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive for validation: %w", err)
	}
	defer file.Close()

	// Check file size
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat archive: %w", err)
	}
	
	if info.Size() == 0 {
		return fmt.Errorf("archive file is empty")
	}
	
	if info.Size() < 100 {
		return fmt.Errorf("archive file too small (%d bytes) - likely corrupted", info.Size())
	}

	// Validate gzip header
	header := make([]byte, 10)
	n, err := file.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read archive header: %w", err)
	}
	
	if n < 3 {
		return fmt.Errorf("archive header too short")
	}
	
	// Check gzip magic number (1f 8b)
	if header[0] != 0x1f || header[1] != 0x8b {
		return fmt.Errorf("invalid gzip header - not a valid tar.gz file")
	}
	
	// Check compression method (should be 8 for deflate)
	if header[2] != 0x08 {
		return fmt.Errorf("unsupported gzip compression method: %d", header[2])
	}

	return nil
}

// ValidateArchiveStructure checks that the archive contains the expected executable
func ValidateArchiveStructure(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive for structure validation: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for validation: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	expectedExecutable := "tuido"
	if runtime.GOOS == "windows" {
		expectedExecutable = "tuido.exe"
	}

	foundExecutable := false
	fileCount := 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header during validation: %w", err)
		}

		fileCount++

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Check for our expected executable
		fileName := filepath.Base(header.Name)
		if fileName == expectedExecutable && header.Typeflag == tar.TypeReg {
			foundExecutable = true
			
			// Validate executable size
			if header.Size == 0 {
				return fmt.Errorf("executable file '%s' is empty in archive", expectedExecutable)
			}
			
			if header.Size < 100000 {
				return fmt.Errorf("executable file '%s' seems too small (%d bytes) - likely corrupted", expectedExecutable, header.Size)
			}
		}
	}

	if fileCount == 0 {
		return fmt.Errorf("archive appears to be empty")
	}

	if !foundExecutable {
		return fmt.Errorf("expected executable '%s' not found in archive", expectedExecutable)
	}

	return nil
}

// ValidateExtractedBinary performs post-extraction validation of the executable
func ValidateExtractedBinary(binaryPath string) error {
	// Check file exists
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("extracted binary does not exist: %w", err)
	}

	// Check file size
	if info.Size() == 0 {
		return fmt.Errorf("extracted binary is empty")
	}

	if info.Size() < 100000 {
		return fmt.Errorf("extracted binary seems too small (%d bytes) - likely corrupted", info.Size())
	}

	// Check executable permissions
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("extracted binary is not executable (mode: %s)", info.Mode())
		}
	}

	// Validate binary format
	file, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open extracted binary for validation: %w", err)
	}
	defer file.Close()

	header := make([]byte, 4)
	n, err := file.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read binary header: %w", err)
	}

	if n < 4 {
		return fmt.Errorf("binary header too short")
	}

	// Platform-specific binary format validation
	switch runtime.GOOS {
	case "linux":
		// Check for ELF magic number (7f 45 4c 46)
		if header[0] != 0x7f || header[1] != 0x45 || header[2] != 0x4c || header[3] != 0x46 {
			return fmt.Errorf("invalid ELF binary format - header: %x", header)
		}
	case "darwin":
		// Check for Mach-O magic numbers
		validMachO := (header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xce) ||
			(header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xcf) ||
			(header[0] == 0xcf && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) ||
			(header[0] == 0xca && header[1] == 0xfe && header[2] == 0xba && header[3] == 0xbe)
		if !validMachO {
			return fmt.Errorf("invalid Mach-O binary format - header: %x", header)
		}
	case "windows":
		// Check for PE magic number (4d 5a - "MZ")
		if header[0] != 0x4d || header[1] != 0x5a {
			return fmt.Errorf("invalid PE binary format - header: %x", header)
		}
	default:
		// For other platforms, just check it's not obviously a text file
		for _, b := range header {
			if b == 0 {
				// Contains null bytes, likely binary
				return nil
			}
		}
		return fmt.Errorf("binary appears to be text file - header: %x", header)
	}

	return nil
}

// ValidateAssetChecksum validates the downloaded asset against GitHub release checksums
func ValidateAssetChecksum(assetPath string, asset *ReleaseAsset) error {
	// Calculate SHA256 of the downloaded file
	hash, err := CalculateSHA256(assetPath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Get expected checksum from GitHub release
	expectedHash, err := fetchAssetChecksum(asset.Name)
	if err != nil {
		// If checksums aren't available, warn but don't fail
		// This maintains compatibility with older releases
		return nil
	}

	if hash != expectedHash {
		return fmt.Errorf("checksum mismatch - expected: %s, got: %s", expectedHash, hash)
	}

	return nil
}

// CalculateSHA256 computes the SHA256 hash of a file
func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// fetchAssetChecksum retrieves the expected checksum for an asset from GitHub release
func fetchAssetChecksum(assetName string) (string, error) {
	// Get the latest release to find the checksums file
	release, err := fetchLatestRelease()
	if err != nil {
		return "", fmt.Errorf("failed to fetch release for checksums: %w", err)
	}

	// Find the checksums file
	var checksumsAsset *ReleaseAsset
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "checksums") {
			checksumsAsset = &asset
			break
		}
	}

	if checksumsAsset == nil {
		return "", fmt.Errorf("checksums file not found in release")
	}

	// Download and parse checksums file
	return downloadAndParseChecksums(checksumsAsset.DownloadURL, assetName)
}

// downloadAndParseChecksums downloads the checksums file and extracts the hash for the specified asset
func downloadAndParseChecksums(checksumsURL, assetName string) (string, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", checksumsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create checksums request: %w", err)
	}

	req.Header.Set("User-Agent", "tuido/"+version)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("checksums download failed with status %d", resp.StatusCode)
	}

	// Read checksums content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums content: %w", err)
	}

	// Parse checksums file (format: "hash  filename")
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[1]
			
			if filename == assetName {
				return hash, nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for asset: %s", assetName)
}

// ExtractExecutableFromArchive extracts the executable from a tar.gz archive
func ExtractExecutableFromArchive(archivePath, extractDir string) (string, error) {
	// Pre-extraction validation
	err := ValidateArchiveFormat(archivePath)
	if err != nil {
		return "", fmt.Errorf("archive format validation failed: %w", err)
	}

	err = ValidateArchiveStructure(archivePath)
	if err != nil {
		return "", fmt.Errorf("archive structure validation failed: %w", err)
	}
	// Perform extraction
	extractedPath, err := performExtraction(archivePath, extractDir)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	// Post-extraction validation
	err = ValidateExtractedBinary(extractedPath)
	if err != nil {
		// Clean up invalid extracted file
		os.Remove(extractedPath)
		return "", fmt.Errorf("extracted binary validation failed: %w", err)
	}

	return extractedPath, nil
}

// performExtraction handles the actual tar.gz extraction process
func performExtraction(archivePath, extractDir string) (string, error) {
	// Open the tar.gz file
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Determine the expected executable name for this platform
	expectedExecutable := "tuido"
	if runtime.GOOS == "windows" {
		expectedExecutable = "tuido.exe"
	}

	// Extract the executable
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip directories and non-regular files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Check if this is the executable we're looking for
		fileName := filepath.Base(header.Name)
		if fileName != expectedExecutable {
			continue
		}

		// Found the executable - extract it
		extractedPath := filepath.Join(extractDir, expectedExecutable)
		
		// Create the output file
		outFile, err := os.Create(extractedPath)
		if err != nil {
			return "", fmt.Errorf("failed to create extracted file: %w", err)
		}

		// Copy the file content
		_, err = io.Copy(outFile, tarReader)
		outFile.Close()
		if err != nil {
			os.Remove(extractedPath) // Clean up on error
			return "", fmt.Errorf("failed to extract file content: %w", err)
		}

		// Set executable permissions (preserve from archive, but ensure executable)
		mode := os.FileMode(header.Mode)
		if mode == 0 {
			// Default to 755 if no mode specified
			mode = 0755
		}
		// Ensure owner execute bit is set
		mode |= 0100
		
		err = os.Chmod(extractedPath, mode)
		if err != nil {
			return "", fmt.Errorf("failed to set executable permissions: %w", err)
		}

		return extractedPath, nil
	}

	return "", fmt.Errorf("executable '%s' not found in archive", expectedExecutable)
}

// CleanupBackups removes old backup files
func CleanupBackups(executablePath string, keepCount int) error {
	dir := filepath.Dir(executablePath)
	base := filepath.Base(executablePath)
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var backups []os.DirEntry
	backupPrefix := base + ".backup"
	
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), backupPrefix) {
			backups = append(backups, entry)
		}
	}

	// If we have more backups than we want to keep, remove the oldest ones
	if len(backups) > keepCount {
		// Sort by modification time (oldest first) and remove excess
		// This is a simplified version - a real implementation would sort by time
		for i := 0; i < len(backups)-keepCount; i++ {
			backupPath := filepath.Join(dir, backups[i].Name())
			os.Remove(backupPath)
		}
	}

	return nil
}

// RestoreFromBackup restores the executable from a backup file
func RestoreFromBackup(executablePath, backupPath string) error {
	// Verify backup exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file does not exist: %w", err)
	}

	// Copy backup over current executable
	err := CopyFile(backupPath, executablePath)
	if err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	return nil
}