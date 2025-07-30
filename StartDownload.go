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
	// For now, always use single stream
	// Multi-stream will be implemented in the future
	if !d.ServerHeaders.AcceptsRanges {
		// Server doesn't support ranges - must use single stream
		d.DownloadSingleStream()
	} else {
		// Server supports ranges - use single stream for now
		// TODO: Implement multi-stream download for large files
		d.DownloadSingleStream()
	}
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
