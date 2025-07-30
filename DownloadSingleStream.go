package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"udm/ufs"
)

// DownloadSingleStream performs a single-threaded download with pause/resume/cancel functionality.
// This function handles downloads for servers that don't support range requests or for small files.
// It includes concurrent header fetching to detect range support during download.
//
// Features:
//   - Play/Pause/Cancel/Resume functionality
//   - Concurrent header analysis during download
//   - Progress tracking with callbacks
//   - Automatic elevation to multi-stream if conditions are met
//   - Robust error handling and recovery
//   - Resume capability using HTTP Range requests
//
// Process Flow:
//  1. Initialize download session and validate prerequisites
//  2. Setup output file and resume detection
//  3. Start concurrent header analysis
//  4. Begin download with progress tracking
//  5. Monitor for elevation to multi-stream
//  6. Handle pause/resume/cancel operations
//  7. Finalize download and cleanup
//
// Parameters:
//   - Uses downloader instance fields for configuration
//
// Returns:
//   - Updates downloader status and calls appropriate callbacks
//
// Example Usage:
//
//	downloader := &Downloader{
//	    Url: "https://example.com/file.zip",
//	    Prefs: UserPreferences{DownloadDir: "./downloads"},
//	}
//	downloader.DownloadSingleStream()
func (d *Downloader) DownloadSingleStream() {
	// Initialize download session
	if err := d.initializeSingleStreamDownload(); err != nil {
		d.handleDownloadError(err)
		return
	}

	// Setup download context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start download process
	d.executeSingleStreamDownload(ctx, cancel)
}

// initializeSingleStreamDownload prepares the download session by validating
// prerequisites, setting up file paths, and initializing progress tracking.
//
// Returns:
//   - error: Error if initialization fails
func (d *Downloader) initializeSingleStreamDownload() error {
	// Set initial status
	d.Status = DOWNLOAD_IN_PROGRESS
	d.TimeStats.StartTime = time.Now()

	// Initialize progress tracker if not exists
	if d.Progress == nil {
		d.Progress = &ProgressTracker{
			LastReported: time.Now(),
		}
	}

	// Initialize pause controller if not exists
	if d.PauseControl == nil {
		d.PauseControl = &PauseController{
			isPaused: false,
		}
		d.PauseControl.cond = sync.NewCond(&d.PauseControl.mu)
	}

	// Setup file paths
	if err := d.setupDownloadPaths(); err != nil {
		return fmt.Errorf("failed to setup download paths: %v", err)
	}

	// Call start callback
	if d.Callbacks != nil && d.Callbacks.OnStart != nil {
		d.Callbacks.OnStart(d)
	}

	return nil
}

// setupDownloadPaths configures the output directory and filename based on
// user preferences, server headers, and system defaults.
//
// Returns:
//   - error: Error if path setup fails
func (d *Downloader) setupDownloadPaths() error {
	// Determine download directory
	downloadDir := d.getDownloadDirectory()
	if downloadDir == "" {
		// Use OS default downloads directory
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current working directory
			downloadDir, _ = os.Getwd()
		} else {
			downloadDir = filepath.Join(userHomeDir, "Downloads")
		}
	}

	// Ensure download directory exists
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	// Determine filename
	filename := d.getUserPreferredFilename()
	if filename == "" {
		filename = d.ServerHeaders.Filename
		if filename == "" {
			filename = "downloaded_file"
			// Add extension from MIME type if available
			if d.ServerHeaders.Filetype != "" {
				ext := mimeExtensionFromContentType(d.ServerHeaders.Filetype)
				filename += ext
			}
		}
	}

	// Create full path and ensure uniqueness
	fullPath := filepath.Join(downloadDir, filename)
	uniquePath := ufs.GenerateUniqueFilename(fullPath)

	// Update file info
	d.fileInfo.Dir = downloadDir
	d.fileInfo.Name = filepath.Base(uniquePath)
	d.fileInfo.FullPath = uniquePath
	d.OutputPath = uniquePath

	return nil
}

// executeSingleStreamDownload performs the actual download with concurrent
// header analysis and progress tracking.
//
// Parameters:
//   - ctx: Context for cancellation
//   - cancel: Cancel function for stopping download
func (d *Downloader) executeSingleStreamDownload(ctx context.Context, cancel context.CancelFunc) {
	// Start concurrent header analysis
	headerChan := make(chan *ServerData, 1)
	go d.concurrentHeaderAnalysis(ctx, headerChan)

	// Check for existing partial download
	resumeOffset, err := d.detectResumeOffset()
	if err != nil {
		d.handleDownloadError(fmt.Errorf("failed to detect resume offset: %v", err))
		return
	}

	// Perform the download
	if err := d.performSingleStreamDownload(ctx, resumeOffset, headerChan); err != nil {
		if ctx.Err() == context.Canceled {
			d.Status = DOWNLOAD_STOPPED
			if d.Callbacks != nil && d.Callbacks.OnStop != nil {
				d.Callbacks.OnStop(d)
			}
		} else {
			d.handleDownloadError(err)
		}
		return
	}

	// Download completed successfully
	d.finalizeDownload()
}

// concurrentHeaderAnalysis performs header analysis alongside the download
// to detect if the server supports range requests during the actual download.
//
// Parameters:
//   - ctx: Context for cancellation
//   - headerChan: Channel to send updated header information
func (d *Downloader) concurrentHeaderAnalysis(ctx context.Context, headerChan chan<- *ServerData) {
	defer close(headerChan)

	// Wait a bit before starting header analysis to let download begin
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return
	}

	// Perform GET request to get headers during download
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", d.Url, nil)
	if err != nil {
		return
	}

	// Add custom headers if available
	for key, value := range d.Headers.Headers {
		req.Header.Set(key, value)
	}

	if d.Headers.Cookies != "" {
		req.Header.Set("Cookie", d.Headers.Cookies)
	}

	// Make a partial request to get headers
	req.Header.Set("Range", "bytes=0-1023") // Request first 1KB

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Discard the body since we only want headers
	io.Copy(io.Discard, resp.Body)

	// Analyze response headers
	updatedHeaders := &ServerData{
		Filename:      extractFilename(resp),
		Filetype:      resp.Header.Get("Content-Type"),
		FinalURL:      resp.Request.URL.String(),
		AcceptsRanges: resp.StatusCode == 206 || resp.Header.Get("Accept-Ranges") == "bytes",
	}

	// Get content length from Content-Range header if available
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		if n, err := fmt.Sscanf(contentRange, "bytes 0-1023/%d", &updatedHeaders.Filesize); n == 1 && err == nil {
			// Successfully parsed file size from Content-Range
		}
	} else if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			updatedHeaders.Filesize = size
		}
	}

	// Send updated headers
	select {
	case headerChan <- updatedHeaders:
	case <-ctx.Done():
	}
}

// detectResumeOffset checks if there's a partial download and returns the resume offset.
//
// Returns:
//   - int64: Byte offset to resume from (0 if starting fresh)
//   - error: Error if offset detection fails
func (d *Downloader) detectResumeOffset() (int64, error) {
	if !ufs.FileExists(d.fileInfo.FullPath) {
		return 0, nil
	}

	fileInfo, err := os.Stat(d.fileInfo.FullPath)
	if err != nil {
		return 0, nil // Start fresh if we can't get file info
	}

	// If server supports ranges, we can resume
	if d.ServerHeaders.AcceptsRanges {
		return fileInfo.Size(), nil
	}

	// If no range support, start fresh
	return 0, nil
}

// performSingleStreamDownload executes the actual file download with progress tracking.
//
// Parameters:
//   - ctx: Context for cancellation
//   - resumeOffset: Byte offset to resume from
//   - headerChan: Channel receiving updated header information
//
// Returns:
//   - error: Error if download fails
func (d *Downloader) performSingleStreamDownload(ctx context.Context, resumeOffset int64, headerChan <-chan *ServerData) error {

	// Create HTTP client with granular timeouts, but no total timeout
	client := &http.Client{
		Transport: &http.Transport{
			// Timeout for establishing a connection
			DialContext: (&net.Dialer{
				Timeout: 15 * time.Second,
			}).DialContext,
			// Timeout for waiting for the server's response headers
			ResponseHeaderTimeout: 15 * time.Second,
			// Timeout for waiting for a TLS handshake
			TLSHandshakeTimeout: 10 * time.Second,
		},
		// DO NOT SET THE TOP-LEVEL TIMEOUT FIELD FOR DOWNLOADS
		// Timeout: 30 * time.Second,
	}
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", d.Url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add custom headers
	for key, value := range d.Headers.Headers {
		req.Header.Set(key, value)
	}

	if d.Headers.Cookies != "" {
		req.Header.Set("Cookie", d.Headers.Cookies)
	}

	// Add range header for resume if supported and needed
	if resumeOffset > 0 && d.ServerHeaders.AcceptsRanges {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get content length
	contentLength := resp.ContentLength
	totalSize := contentLength
	if resumeOffset > 0 {
		totalSize += resumeOffset
	}

	// Update progress tracker with total size
	d.Progress.mu.Lock()
	d.Progress.BytesCompleted = resumeOffset
	d.Progress.mu.Unlock()

	// Open/create output file
	file, err := d.openOutputFile(resumeOffset)
	if err != nil {
		return fmt.Errorf("failed to open output file: %v", err)
	}
	defer file.Close()

	// Download with progress tracking
	return d.downloadWithProgress(ctx, resp.Body, file, totalSize, headerChan)
}

// openOutputFile opens the output file for writing, handling resume scenarios.
//
// Parameters:
//   - resumeOffset: Byte offset to resume from
//
// Returns:
//   - *os.File: File handle for writing
//   - error: Error if file opening fails
func (d *Downloader) openOutputFile(resumeOffset int64) (*os.File, error) {
	if resumeOffset > 0 {
		// Open for appending
		return os.OpenFile(d.fileInfo.FullPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		// Create new file
		return os.Create(d.fileInfo.FullPath)
	}
}

// downloadWithProgress performs the download with progress tracking and pause/resume support.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reader: Source reader (response body)
//   - writer: Destination writer (file)
//   - totalSize: Total expected size
//   - headerChan: Channel for updated headers
//
// Returns:
//   - error: Error if download fails
func (d *Downloader) downloadWithProgress(ctx context.Context, reader io.Reader, writer io.Writer, totalSize int64, headerChan <-chan *ServerData) error {
	buffer := make([]byte, 32*1024) // 32KB buffer
	elevationChecked := false

	for {
		// Check for pause
		d.checkPauseState()

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case updatedHeaders := <-headerChan:
			// Handle updated headers from concurrent analysis
			if updatedHeaders != nil {
				d.handleUpdatedHeaders(updatedHeaders, &elevationChecked, totalSize)
			}
		default:
		}

		// Read data
		n, err := reader.Read(buffer)
		if n > 0 {
			// Write data
			written, writeErr := writer.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write data: %v", writeErr)
			}

			// Update progress
			d.updateProgress(int64(written), totalSize)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read data: %v", err)
		}
	}

	return nil
}

// handleUpdatedHeaders processes updated server headers received during download.
//
// Parameters:
//   - headers: Updated server headers
//   - elevationChecked: Pointer to elevation check flag
//   - totalSize: Current total download size
func (d *Downloader) handleUpdatedHeaders(headers *ServerData, elevationChecked *bool, totalSize int64) {
	// Update server headers if we got better information
	if headers.Filesize > 0 && d.ServerHeaders.Filesize == 0 {
		d.ServerHeaders.Filesize = headers.Filesize
	}

	if headers.AcceptsRanges && !d.ServerHeaders.AcceptsRanges {
		d.ServerHeaders.AcceptsRanges = true
	}

	// Check for elevation to multi-stream if conditions are met
	if !*elevationChecked && d.shouldElevateToMultiStream(headers, totalSize) {
		*elevationChecked = true
		// TODO: Implement elevation to multi-stream download
		// This would pause current download and switch to multi-threaded mode
	}
}

// shouldElevateToMultiStream determines if download should be elevated to multi-stream.
//
// Parameters:
//   - headers: Updated server headers
//   - totalSize: Total download size
//
// Returns:
//   - bool: True if should elevate to multi-stream
func (d *Downloader) shouldElevateToMultiStream(headers *ServerData, totalSize int64) bool {
	// Don't elevate if range requests aren't supported
	if !headers.AcceptsRanges {
		return false
	}

	// Don't elevate if file is too small (less than 10MB)
	if headers.Filesize > 0 && headers.Filesize < 10*1024*1024 {
		return false
	}

	// Don't elevate if we've already downloaded too much (more than 25%)
	d.Progress.mu.Lock()
	completed := d.Progress.BytesCompleted
	d.Progress.mu.Unlock()

	if totalSize > 0 && float64(completed)/float64(totalSize) > 0.25 {
		return false
	}

	return true
}

// checkPauseState handles pause functionality by blocking when download is paused.
func (d *Downloader) checkPauseState() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	if d.PauseControl.isPaused {
		// We're paused, call the pause callback once without holding the mutex
		pauseCallback := d.Callbacks != nil && d.Callbacks.OnPause != nil
		var pauseFunc func(d *Downloader)
		if pauseCallback {
			pauseFunc = d.Callbacks.OnPause
		}

		d.PauseControl.mu.Unlock()
		if pauseCallback {
			pauseFunc(d)
		}
		d.PauseControl.mu.Lock()

		// Wait for resume
		for d.PauseControl.isPaused {
			d.PauseControl.cond.Wait()
		}

		// We're resumed, call the resume callback once without holding the mutex
		resumeCallback := d.Callbacks != nil && d.Callbacks.OnResume != nil
		var resumeFunc func(d *Downloader)
		if resumeCallback {
			resumeFunc = d.Callbacks.OnResume
		}

		d.PauseControl.mu.Unlock()
		if resumeCallback {
			resumeFunc(d)
		}
		d.PauseControl.mu.Lock()
	}
}

// updateProgress updates the download progress and triggers callbacks.
//
// Parameters:
//   - bytesRead: Number of bytes read in this update
//   - totalSize: Total expected download size
func (d *Downloader) updateProgress(bytesRead int64, totalSize int64) {
	var shouldCallCallback bool

	d.Progress.mu.Lock()
	d.Progress.BytesCompleted += bytesRead
	now := time.Now()

	// Calculate speed every second
	if now.Sub(d.Progress.LastReported) >= time.Second {
		elapsed := now.Sub(d.Progress.LastReported).Seconds()
		d.Progress.SpeedBps = float64(bytesRead) / elapsed
		d.Progress.LastReported = now
		shouldCallCallback = true
	}
	d.Progress.mu.Unlock()

	// Call progress callback outside of mutex to prevent deadlock
	if shouldCallCallback && d.Callbacks != nil && d.Callbacks.OnProgress != nil {
		d.Callbacks.OnProgress(d)
	}
}

// finalizeDownload completes the download process and updates status.
func (d *Downloader) finalizeDownload() {
	d.Status = DOWNLOAD_COMPLETED
	d.TimeStats.EndTime = time.Now()
	d.TimeStats.Elapsed = d.TimeStats.EndTime.Sub(d.TimeStats.StartTime)

	// Call completion callback
	if d.Callbacks != nil && d.Callbacks.OnFinish != nil {
		d.Callbacks.OnFinish(d)
	}
}

// handleDownloadError handles download errors and updates status.
//
// Parameters:
//   - err: The error that occurred
func (d *Downloader) handleDownloadError(err error) {
	d.Status = DOWNLOAD_FAILED
	d.Error = err
	d.TimeStats.EndTime = time.Now()
	d.TimeStats.Elapsed = d.TimeStats.EndTime.Sub(d.TimeStats.StartTime)

	// Call error callback
	if d.Callbacks != nil && d.Callbacks.OnError != nil {
		d.Callbacks.OnError(d, err)
	}
}

// Pause pauses the current download operation.
func (d *Downloader) Pause() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	if !d.PauseControl.isPaused {
		d.PauseControl.isPaused = true
		d.Status = DOWNLOAD_PAUSED
	}
}

// Resume resumes a paused download operation.
func (d *Downloader) Resume() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	if d.PauseControl.isPaused {
		d.PauseControl.isPaused = false
		d.Status = DOWNLOAD_IN_PROGRESS
		d.PauseControl.cond.Broadcast()
	}
}

// Cancel cancels the current download operation.
func (d *Downloader) Cancel() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	d.PauseControl.isPaused = false
	d.Status = DOWNLOAD_STOPPED
	d.PauseControl.cond.Broadcast()
}

// GetProgress returns current download progress information.
//
// Returns:
//   - bytesCompleted: Number of bytes downloaded
//   - percentage: Download completion percentage (0-100)
//   - speedBps: Current download speed in bytes per second
func (d *Downloader) GetProgress() (bytesCompleted int64, percentage float64, speedBps float64) {
	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	bytesCompleted = d.Progress.BytesCompleted
	speedBps = d.Progress.SpeedBps

	if d.ServerHeaders.Filesize > 0 {
		percentage = float64(bytesCompleted) / float64(d.ServerHeaders.Filesize) * 100
	}

	return
}
