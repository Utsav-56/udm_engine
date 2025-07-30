package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// StartDownload initiates the download process by analyzing server capabilities
// and choosing the appropriate download strategy (single-stream vs multi-stream).
//
// Process Flow:
//  1. Prefetch server metadata and capabilities
//  2. Check user preferences and setup file paths
//  3. Determine download strategy based on server support
//  4. Execute appropriate download method
//
// Features:
//   - Automatic server capability detection
//   - Smart download strategy selection
//   - User preference handling
//   - Error handling and recovery
func (d *Downloader) StartDownload() {
	// Initialize download session
	if err := d.initializeDownload(); err != nil {
		d.handleDownloadError(err)
		return
	}

	// Prefetch server information
	if err := d.Prefetch(); err != nil {
		d.handleDownloadError(err)
		return
	}

	// Choose and execute download strategy
	d.executeDownloadStrategy()
}

// initializeDownload sets up the initial download state and validates prerequisites.
//
// Returns:
//   - error: Error if initialization fails
func (d *Downloader) initializeDownload() error {
	// Initialize progress tracker
	if d.Progress == nil {
		d.Progress = &ProgressTracker{}
	}

	// Initialize pause controller
	if d.PauseControl == nil {
		d.PauseControl = NewPauseController()
	}

	// Initialize time stats
	if d.TimeStats == nil {
		d.TimeStats = &TimeInfo{}
	}

	// Set initial status
	d.Status = DOWNLOAD_QUEUED

	return nil
}

// Prefetch retrieves server metadata and analyzes download capabilities.
// This function performs initial server analysis to determine file information
// and server capabilities before starting the actual download.
//
// Returns:
//   - error: Error if prefetch fails
func (d *Downloader) Prefetch() error {
	// Get server data with retry mechanism
	headers, err := GetServerData(d.Url)
	if err != nil {
		return fmt.Errorf("failed to get server data: %v", err)
	}

	if headers == nil {
		return fmt.Errorf("failed to get server data: %v", err)
	}
	// Store server headers
	d.ServerHeaders = *headers

	// Check and apply user preferences
	if err := d.CheckPreferences(); err != nil {
		return fmt.Errorf("failed to check preferences: %v", err)
	}

	return nil
}

// executeDownloadStrategy chooses and executes the appropriate download method
// based on server capabilities and file characteristics.
func (d *Downloader) executeDownloadStrategy() {
	// Check if server supports range requests and file size is sufficient for multi-stream
	if d.ServerHeaders.AcceptsRanges && d.shouldUseMultiStream() {
		// Use multi-stream download for large files with range support
		d.DownloadMultiStream()
	} else {
		// Use single-stream download for small files or servers without range support
		d.DownloadSingleStream()
	}
}

// shouldUseMultiStream determines if multi-stream download should be used
// based on file size and server capabilities.
//
// Returns:
//   - bool: True if multi-stream download should be used
func (d *Downloader) shouldUseMultiStream() bool {
	// Don't use multi-stream if ranges aren't supported
	if !d.ServerHeaders.AcceptsRanges {
		return false
	}

	// Don't use multi-stream if file size is unknown
	if d.ServerHeaders.Filesize <= 0 {
		return false
	}

	// Use multi-stream for files larger than 10MB
	const minSizeForMultiStream = 10 * 1024 * 1024 // 10MB
	if d.ServerHeaders.Filesize < minSizeForMultiStream {
		return false
	}

	// Check if user explicitly requested single stream (threadCount = 1)
	if d.getThreadCount() == 1 {
		return false
	}

	return true
}

// CheckPreferences validates and applies user preferences for the download.
// This function handles filename resolution, directory setup, and other
// user-configurable options.
//
// Returns:
//   - error: Error if preference setup fails
func (d *Downloader) CheckPreferences() error {
	headers := d.ServerHeaders

	// Determine filename based on preferences and server data
	if d.Prefs.fileName != "" {
		// User specified filename takes priority
		d.fileInfo.Name = d.Prefs.fileName
	} else if headers.Filename != "" {
		// Use server-provided filename
		d.fileInfo.Name = headers.Filename
	} else {
		// Use fallback name
		d.fileInfo.Name = "downloaded_file"
		// Add extension from MIME type if available
		if headers.Filetype != "" {
			ext := mimeExtensionFromContentType(headers.Filetype)
			if ext != "" {
				d.fileInfo.Name += ext
			}
		}
	}

	// Determine download directory
	if d.Prefs.DownloadDir != "" {
		// Use user-specified directory
		d.fileInfo.Dir = d.Prefs.DownloadDir
	} else {
		// Use OS default downloads directory
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %v", err)
			}
			d.fileInfo.Dir = cwd
		} else {
			d.fileInfo.Dir = filepath.Join(userHomeDir, "Downloads")
		}
	}

	// Ensure directory path is absolute
	absDir, err := filepath.Abs(d.fileInfo.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %v", err)
	}
	d.fileInfo.Dir = absDir

	// Create full path
	d.fileInfo.FullPath = filepath.Join(d.fileInfo.Dir, d.fileInfo.Name)

	return nil
}
