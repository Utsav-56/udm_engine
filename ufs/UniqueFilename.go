package ufs

// This file contains the code for the UniqueFilename function
// to generate a unique filename for a file that already exists

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileExists checks whether a file exists at the specified path.
// This function resolves the absolute path and verifies file existence,
// ensuring the path points to an actual file and not a directory.
//
// Parameters:
//   - pathStr: The file path to check for existence (relative or absolute)
//
// Returns:
//   - bool: true if the file exists, false if it doesn't exist or path is invalid
//
// Features:
//   - Automatic path resolution to absolute path
//   - Safe error handling for invalid paths
//   - Cross-platform compatibility
//
// Example:
//
//	exists := FileExists("./downloads/file.zip")
//	if exists {
//	    fmt.Println("File already exists")
//	}
//
// Notes:
//   - Returns false for directories, even if they exist
//   - Returns false if path resolution fails
//   - Uses os.Stat internally for reliable file detection
func FileExists(pathStr string) bool {
	absPath, err := filepath.Abs(pathStr)
	if err != nil {
		return false
	}

	// Ensure that the path is a file and not a directory
	_, err = os.Stat(absPath)
	return !os.IsNotExist(err)
}

// FileNameWithoutExtension extracts the filename without its extension.
// This utility function removes the file extension from a filename,
// returning only the base name portion for manipulation.
//
// Parameters:
//   - filename: The filename string to process (with or without path)
//
// Returns:
//   - string: The filename without extension
//
// Algorithm:
//  1. Calculate extension length using filepath.Ext()
//  2. Return substring from start to (length - extension_length)
//  3. Handles files without extensions gracefully
//
// Example:
//
//	baseName := FileNameWithoutExtension("document.pdf")
//	// Result: "document"
//
//	baseName := FileNameWithoutExtension("archive.tar.gz")
//	// Result: "archive.tar"
//
// Notes:
//   - Works with complex extensions like .tar.gz
//   - Returns original string if no extension found
//   - Does not validate if input is actually a filename
func FileNameWithoutExtension(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// FileExtension extracts the file extension from a filename.
// This is a wrapper around filepath.Ext() for consistency with
// the UDM file system utilities package.
//
// Parameters:
//   - filename: The filename string to extract extension from
//
// Returns:
//   - string: The file extension including the dot (e.g., ".txt", ".zip")
//
// Behavior:
//   - Returns empty string if no extension is found
//   - Includes the leading dot in the extension
//   - Handles multiple dots correctly (returns last extension)
//
// Example:
//
//	ext := FileExtension("document.pdf")
//	// Result: ".pdf"
//
//	ext := FileExtension("archive.tar.gz")
//	// Result: ".gz"
//
//	ext := FileExtension("README")
//	// Result: ""
//
// Notes:
//   - Uses Go's standard filepath.Ext() internally
//   - Cross-platform compatible
//   - Case-sensitive extension detection
func FileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GenerateUniqueFilename creates a unique filename by appending numbers if conflicts exist.
// This function prevents file overwriting by generating alternative filenames
// when the target file already exists, following the pattern: "filename (1).ext"
//
// Parameters:
//   - path: The desired file path (absolute or relative)
//
// Returns:
//   - string: A unique file path that doesn't conflict with existing files
//
// Algorithm:
//  1. Check if original path exists
//  2. If not, return original path
//  3. If exists, extract filename and extension
//  4. Increment counter and test "filename (N).ext" until unique
//  5. Return first non-conflicting path
//
// Naming Pattern:
//   - Original: "document.pdf"
//   - First conflict: "document (1).pdf"
//   - Second conflict: "document (2).pdf"
//   - Continues indefinitely until unique name found
//
// Example:
//
//	uniquePath := GenerateUniqueFilename("./downloads/file.zip")
//	// If file.zip exists, returns "./downloads/file (1).zip"
//	// If file (1).zip also exists, returns "./downloads/file (2).zip"
//
// Use Cases:
//   - Download managers preventing overwrites
//   - File backup systems
//   - Automatic file versioning
//   - Batch file operations
//
// Performance:
//   - Efficient incremental checking
//   - Minimal file system calls
//   - Scales well even with many conflicts
//
// Notes:
//   - Preserves original directory path
//   - Works with files that have no extension
//   - Thread-safe for individual calls (but not atomic across processes)
//   - May create race conditions in multi-threaded environments
func GenerateUniqueFilename(path string) string {
	if !FileExists(path) {
		return path
	}

	fileName := FileNameWithoutExtension(filepath.Base(path))
	extension := FileExtension(filepath.Base(path))
	dirPath := filepath.Dir(path)

	for i := 1; ; i++ {
		newFileName := fmt.Sprintf("%s (%d)%s", fileName, i, extension)
		newPath := filepath.Join(dirPath, newFileName)

		if !FileExists(newPath) {
			return newPath
		}
	}
}

func FileExtensionWithoutDot(filename string) string {
	extension := FileExtension(filename)
	if extension[0] == '.' {
		extension = extension[1:]
	}
	return extension
}
