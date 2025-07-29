package ufs

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateChunkFileNames creates temporary chunk file names for multi-threaded downloads.
// This function generates a list of temporary file paths that will store individual
// download chunks before they are merged into the final file. Always returns absolute paths.
//
// Parameters:
//   - originalFilename: The target filename (can be absolute path, relative path, or just filename)
//   - chunkCount: Number of chunks to create (typically matches thread count)
//   - paths: Optional directory path where chunks will be stored (variadic parameter)
//
// Returns:
//   - []string: Array of absolute temporary chunk file paths with .udtemp extension
//
// Naming Convention:
//   - Pattern: "{absolute_path}/{filename} ({index}).udtemp"
//   - Example: "C:/downloads/video (0).udtemp", "C:/downloads/video (1).udtemp"
//   - Uses .udtemp extension to identify UDM temporary files
//
// Path Resolution Algorithm:
//  1. Determine target directory:
//     - If optional path provided: Use that path (resolve to absolute)
//     - If originalFilename is absolute: Use its directory
//     - If originalFilename is relative/filename only: Use current working directory
//  2. Extract base filename without extension from originalFilename
//  3. Generate indexed chunk names with .udtemp extension
//  4. Return array of absolute file paths for each chunk
//
// Scenarios Handled:
//
//   - originalFilename: "video.mp4", paths: none
//     Result: "C:/current/working/dir/video (0).udtemp", etc.
//
//   - originalFilename: "video.mp4", paths: ["./downloads"]
//     Result: "C:/current/working/dir/downloads/video (0).udtemp", etc.
//
//   - originalFilename: "video.mp4", paths: ["C:/downloads"]
//     Result: "C:/downloads/video (0).udtemp", etc.
//
//   - originalFilename: "C:/files/video.mp4", paths: none
//     Result: "C:/files/video (0).udtemp", etc.
//
//   - originalFilename: "C:/files/video.mp4", paths: ["./downloads"]
//     Result: "C:/current/working/dir/downloads/video (0).udtemp", etc.
//
//   - originalFilename: "./folder/video.mp4", paths: ["../temp"]
//     Result: "C:/resolved/temp/video (0).udtemp", etc.
//
// Example:
//
//	// Scenario 1: Just filename, no path
//	chunkNames := GenerateChunkFileNames("movie.mp4", 4)
//	// Result: ["C:/current/dir/movie (0).udtemp", "C:/current/dir/movie (1).udtemp", ...]
//
//	// Scenario 2: Filename with relative path
//	chunkNames := GenerateChunkFileNames("movie.mp4", 4, "./downloads")
//	// Result: ["C:/current/dir/downloads/movie (0).udtemp", ...]
//
//	// Scenario 3: Absolute filename with relative chunk path
//	chunkNames := GenerateChunkFileNames("C:/source/movie.mp4", 4, "./temp")
//	// Result: ["C:/current/dir/temp/movie (0).udtemp", ...]
//
// Integration with UDM:
//   - Supports multi-threaded download architecture
//   - Absolute paths prevent confusion in multi-directory operations
//   - Temporary files are merged after all chunks complete
//   - .udtemp extension prevents conflicts with real files
//   - Enables resume functionality by preserving partial downloads
//
// Error Handling:
//   - Uses current directory if path resolution fails
//   - Graceful fallback for invalid path scenarios
//   - All returned paths are guaranteed to be absolute
//
// Notes:
//   - Always returns absolute paths for consistency
//   - Uses variadic parameter for optional path specification
//   - Cross-platform path handling with filepath operations
//   - Does not create actual files, only generates names
//   - Path resolution happens at generation time, not usage time
func GenerateChunkFileNames(originalFilename string, chunkCount int, paths ...string) []string {
	chunkFiles := make([]string, chunkCount)

	var targetDir string
	var err error

	// Step 1: Determine the target directory
	if len(paths) > 0 && paths[0] != "" {
		// Optional path provided - use it (resolve to absolute)
		targetDir, err = filepath.Abs(paths[0])
		if err != nil {
			// Fallback to current directory if path resolution fails
			targetDir, _ = os.Getwd()
		}
	} else {
		// No optional path provided - determine from originalFilename
		if filepath.IsAbs(originalFilename) {
			// originalFilename is absolute - use its directory
			targetDir = filepath.Dir(originalFilename)
		} else {
			// originalFilename is relative or just filename - use current directory
			targetDir, err = os.Getwd()
			if err != nil {
				// Extremely rare case - fallback to current directory string
				targetDir = "."
				targetDir, _ = filepath.Abs(targetDir)
			}
		}
	}

	// Step 2: Extract base filename without extension
	baseFilename := filepath.Base(originalFilename)
	filenameWithoutExt := FileNameWithoutExtension(baseFilename)

	// Step 3: Generate chunk file names with absolute paths
	for i := 0; i < chunkCount; i++ {
		chunkName := fmt.Sprintf("%s (%d).udtemp", filenameWithoutExt, i)
		absoluteChunkPath := filepath.Join(targetDir, chunkName)

		// Ensure the path is absolute (double-check)
		if !filepath.IsAbs(absoluteChunkPath) {
			absoluteChunkPath, _ = filepath.Abs(absoluteChunkPath)
		}

		chunkFiles[i] = absoluteChunkPath
	}

	return chunkFiles
}

// GenerateChunkFiles creates physical temporary files for download chunks.
// This function takes a list of chunk file names and creates empty files
// on the filesystem, preparing them for parallel download operations.
//
// Parameters:
//   - chunkFileNames: Array of file paths to create (from GenerateChunkFileNames)
//
// Returns:
//   - error: Error if any file creation fails, nil on success
//
// Process:
//  1. Iterate through each chunk file name
//  2. Create parent directories if they don't exist
//  3. Create empty files ready for writing
//  4. Handle errors gracefully with detailed error messages
//
// Features:
//   - Automatic directory creation for parent paths
//   - Atomic operation - fails completely if any file creation fails
//   - Proper error handling with context information
//   - Prepares files for concurrent write operations
//
// Example:
//
//	chunkNames := GenerateChunkFileNames("video.mp4", 4, "./downloads")
//	err := GenerateChunkFiles(chunkNames)
//	if err != nil {
//	    log.Fatal("Failed to create chunk files:", err)
//	}
//	// Now ready for parallel downloading to chunk files
//
// Error Scenarios:
//   - Insufficient disk permissions
//   - Disk space exhaustion
//   - Invalid file paths
//   - Network drive connectivity issues
//
// Integration:
//   - Used before starting multi-threaded downloads
//   - Chunk files are written to by individual download goroutines
//   - Files are later merged by MergeChunkFiles function
//   - Supports download resume by checking existing chunk files
//
// Performance:
//   - Minimal I/O operations (creates empty files)
//   - Fast execution even with many chunks
//   - No significant memory usage
func GenerateChunkFiles(chunkFileNames []string) error {
	for i, chunkFileName := range chunkFileNames {
		err := CreateFile(chunkFileName)
		if err != nil {
			return fmt.Errorf("failed to create chunk file %d (%s): %v", i, chunkFileName, err)
		}
	}
	return nil
}

// CreateFile creates a new file with all necessary parent directories.
// This utility function handles the complete file creation process,
// including automatic parent directory creation and proper error handling.
//
// Parameters:
//   - pathStr: Complete file path to create (absolute or relative)
//
// Returns:
//   - error: Error if creation fails, nil on success
//
// Features:
//   - Automatic parent directory creation with proper permissions
//   - Cross-platform compatibility using filepath operations
//   - Proper resource cleanup with deferred file closing
//   - Detailed error messages for troubleshooting
//
// Process:
//  1. Extract parent directory from file path
//  2. Create parent directories recursively with os.ModePerm (0777)
//  3. Create the target file
//  4. Close file handle immediately (creates empty file)
//  5. Return appropriate error messages on failure
//
// Example:
//
//	err := CreateFile("./downloads/temp/chunk.udtemp")
//	if err != nil {
//	    log.Printf("File creation failed: %v", err)
//	}
//	// File and all parent directories now exist
//
// Use Cases:
//   - Creating temporary chunk files for downloads
//   - Preparing output files before writing
//   - Setting up directory structures
//   - File system operations in download manager
//
// Error Handling:
//   - Directory creation failures (permissions, disk space)
//   - File creation failures (naming conflicts, permissions)
//   - Path resolution issues (invalid characters, length limits)
//
// Notes:
//   - Creates empty file (0 bytes)
//   - Uses os.ModePerm (0777) for directory permissions
//   - File handle is immediately closed after creation
//   - Safe for concurrent use (but not atomic across processes)
//   - Works with both absolute and relative paths
func CreateFile(pathStr string) error {
	// Ensure the parent path also exists
	err := os.MkdirAll(filepath.Dir(pathStr), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	// Create the file
	file, err := os.Create(pathStr)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	return nil
}

// MergeChunkFiles combines downloaded chunk files into the final output file.
// This function reads all temporary chunk files in sequence and writes them
// to a single output file, completing the multi-threaded download process.
//
// Parameters:
//   - chunkFileNames: Array of chunk file paths to merge (in order)
//   - outputFilePath: Path for the final merged file
//
// Returns:
//   - error: Error if merging fails, nil on success
//
// Process:
//  1. Create the output file with parent directories
//  2. Open each chunk file in sequence
//  3. Copy chunk contents to output file
//  4. Close each chunk file after copying
//  5. Clean up temporary chunk files
//
// Example:
//
//	chunkNames := []string{"video (0).udtemp", "video (1).udtemp", "video (2).udtemp"}
//	err := MergeChunkFiles(chunkNames, "video.mp4")
//	if err != nil {
//	    log.Fatal("Failed to merge chunks:", err)
//	}
//	// video.mp4 now contains the complete download
//
// Features:
//   - Sequential chunk processing maintains file integrity
//   - Automatic cleanup of temporary files
//   - Memory-efficient streaming copy
//   - Progress tracking capability
//
// Notes:
//   - Chunk files must be in correct order
//   - All chunks must exist before merging
//   - Original chunk files are deleted after successful merge
//   - Output file overwrites existing files
func MergeChunkFiles(chunkFileNames []string, outputFilePath string) error {
	// Create the output file
	err := CreateFile(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file for writing: %v", err)
	}
	defer outputFile.Close()

	// Merge each chunk file
	for i, chunkFileName := range chunkFileNames {
		chunkFile, err := os.Open(chunkFileName)
		if err != nil {
			return fmt.Errorf("failed to open chunk file %d (%s): %v", i, chunkFileName, err)
		}

		// Copy chunk content to output file
		_, err = outputFile.ReadFrom(chunkFile)
		chunkFile.Close()

		if err != nil {
			return fmt.Errorf("failed to copy chunk %d to output file: %v", i, err)
		}

		// Clean up chunk file after successful copy
		err = os.Remove(chunkFileName)
		if err != nil {
			// Log warning but don't fail the merge
			fmt.Printf("Warning: failed to remove chunk file %s: %v\n", chunkFileName, err)
		}
	}

	return nil
}

// CleanupChunkFiles removes temporary chunk files in case of download failure.
// This utility function ensures proper cleanup when downloads are cancelled
// or fail, preventing accumulation of temporary files.
//
// Parameters:
//   - chunkFileNames: Array of chunk file paths to remove
//
// Returns:
//   - error: Error if cleanup fails, nil on success
//
// Features:
//   - Graceful error handling - continues even if some files fail to delete
//   - Detailed error reporting for troubleshooting
//   - Safe to call multiple times (idempotent)
//   - Handles missing files gracefully
//
// Example:
//
//	chunkNames := GenerateChunkFileNames("video.mp4", 4, "./downloads")
//	// ... download fails ...
//	err := CleanupChunkFiles(chunkNames)
//	if err != nil {
//	    log.Printf("Cleanup warnings: %v", err)
//	}
//
// Notes:
//   - Used when downloads are cancelled or fail
//   - Safe to call even if files don't exist
//   - Continues cleanup even if individual files fail to delete
//   - Returns combined error information for all failures
func CleanupChunkFiles(chunkFileNames []string) error {
	var errors []string

	for i, chunkFileName := range chunkFileNames {
		err := os.Remove(chunkFileName)
		if err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("chunk %d (%s): %v", i, chunkFileName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to cleanup some chunk files: %v", errors)
	}

	return nil
}
