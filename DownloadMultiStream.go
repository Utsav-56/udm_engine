package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"udm/ufs"
)

// DownloadMultiStream performs a multi-threaded download with pause/resume/cancel functionality.
// This function handles downloads for servers that support range requests and large files.
// It downloads chunks concurrently and merges them after completion.
//
// Features:
//   - Multi-threaded concurrent chunk downloading
//   - Play/Pause/Cancel/Resume functionality
//   - Progress tracking with callbacks for overall and per-chunk progress
//   - Automatic chunk file management and merging
//   - Robust error handling and recovery
//   - Resume capability for incomplete downloads
//
// Process Flow:
//  1. Initialize multi-stream session and validate prerequisites
//  2. Calculate optimal chunk divisions based on file size
//  3. Create temporary chunk files
//  4. Start concurrent chunk download workers
//  5. Monitor progress and handle pause/resume/cancel operations
//  6. Merge completed chunks into final file
//  7. Cleanup temporary files and finalize download
//
// Parameters:
//   - Uses downloader instance fields for configuration
//   - Automatically determines optimal thread count if not specified
//
// Returns:
//   - Updates downloader status and calls appropriate callbacks
//
// Example Usage:
//
//	downloader := &Downloader{
//	    Url: "https://example.com/largefile.zip",
//	    Prefs: UserPreferences{
//	        DownloadDir: "./downloads",
//	        threadCount: 8,
//	    },
//	}
//	downloader.DownloadMultiStream()
func (d *Downloader) DownloadMultiStream() {
	// Initialize multi-stream session
	if err := d.initializeMultiStreamDownload(); err != nil {
		d.handleDownloadError(err)
		return
	}

	// Setup download context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start multi-stream download process
	d.executeMultiStreamDownload(ctx, cancel)
}

// initializeMultiStreamDownload prepares the multi-stream download session by validating
// prerequisites, setting up file paths, and initializing progress tracking.
//
// Returns:
//   - error: Error if initialization fails
func (d *Downloader) initializeMultiStreamDownload() error {
	// Set initial status
	d.Status = DOWNLOAD_IN_PROGRESS
	d.TimeStats.StartTime = time.Now()

	// Initialize progress tracker if not exists
	if d.Progress == nil {
		d.Progress = &ProgressTracker{
			LastReported: time.Now(),
			StartTime:    time.Now(),
		}
	}

	// Initialize pause controller if not exists
	if d.PauseControl == nil {
		d.PauseControl = NewPauseController()
	}

	// Setup file paths
	if err := d.setupDownloadPaths(); err != nil {
		return fmt.Errorf("failed to setup download paths: %v", err)
	}

	// Validate server supports ranges
	if !d.ServerHeaders.AcceptsRanges {
		return fmt.Errorf("server does not support range requests - cannot use multi-stream download")
	}

	// Validate file size is known
	if d.ServerHeaders.Filesize <= 0 {
		return fmt.Errorf("file size unknown - cannot divide into chunks")
	}

	// Call start callback
	if d.Callbacks != nil && d.Callbacks.OnStart != nil {
		d.Callbacks.OnStart(d)
	}

	return nil
}

// executeMultiStreamDownload performs the actual multi-threaded download with progress tracking.
//
// Parameters:
//   - ctx: Context for cancellation
//   - cancel: Cancel function for stopping download
func (d *Downloader) executeMultiStreamDownload(ctx context.Context, cancel context.CancelFunc) {
	// Determine optimal thread count
	threadCount := d.getOptimalThreadCount()

	// Divide file into chunks
	chunkSizes := DivideChunks(d.ServerHeaders.Filesize, threadCount)

	// Initialize chunk data structures
	if err := d.initializeChunks(chunkSizes); err != nil {
		d.handleDownloadError(fmt.Errorf("failed to initialize chunks: %v", err))
		return
	}

	// Create chunk files
	chunkFileNames := ufs.GenerateChunkFileNames(d.fileInfo.Name, threadCount, d.fileInfo.Dir)
	if err := ufs.GenerateChunkFiles(chunkFileNames); err != nil {
		d.handleDownloadError(fmt.Errorf("failed to create chunk files: %v", err))
		return
	}

	// Initialize progress tracking for total size
	d.Progress.UpdateProgress(0, d.ServerHeaders.Filesize)

	// Start concurrent chunk downloads
	if err := d.downloadChunksConcurrently(ctx, chunkFileNames); err != nil {
		// Cleanup chunk files on failure
		ufs.CleanupChunkFiles(chunkFileNames)
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

	// Merge chunks into final file
	if err := d.mergeChunksToFinalFile(chunkFileNames); err != nil {
		d.handleDownloadError(fmt.Errorf("failed to merge chunks: %v", err))
		return
	}

	// Download completed successfully
	d.finalizeDownload()
}

// getOptimalThreadCount determines the optimal number of threads for download.
//
// Returns:
//   - int: Optimal thread count based on file size and user preferences
func (d *Downloader) getOptimalThreadCount() int {
	userThreadCount := d.getThreadCount()
	if userThreadCount > 0 {
		return userThreadCount
	}

	// Auto-determine based on file size
	fileSize := d.ServerHeaders.Filesize
	switch {
	case fileSize < 10*1024*1024: // < 10MB
		return 2
	case fileSize < 100*1024*1024: // < 100MB
		return 4
	case fileSize < 1024*1024*1024: // < 1GB
		return 8
	default: // >= 1GB
		return 12
	}
}

// initializeChunks creates chunk data structures for tracking download progress.
//
// Parameters:
//   - chunkSizes: Array of chunk sizes in bytes
//
// Returns:
//   - error: Error if initialization fails
func (d *Downloader) initializeChunks(chunkSizes []int64) error {
	d.Chunks = make([]ChunkData, len(chunkSizes))

	var currentOffset int64 = 0
	for i, size := range chunkSizes {
		d.Chunks[i] = ChunkData{
			Index:       i,
			Start:       currentOffset,
			End:         currentOffset + size - 1,
			Size:        size,
			IsCompleted: false,
		}
		currentOffset += size
	}

	// Initialize chunk manager
	d.ChunkManager = &ChunkManager{
		Chunks:         d.Chunks,
		ChunkSize:      chunkSizes[0], // Use first chunk size as reference
		TotalSize:      d.ServerHeaders.Filesize,
		CompletedBytes: 0,
	}

	return nil
}

// downloadChunksConcurrently starts concurrent workers to download all chunks.
//
// Parameters:
//   - ctx: Context for cancellation
//   - chunkFileNames: Array of chunk file paths
//
// Returns:
//   - error: Error if download fails
func (d *Downloader) downloadChunksConcurrently(ctx context.Context, chunkFileNames []string) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(d.Chunks))

	// Track completed bytes atomically
	var totalCompletedBytes int64

	// Start workers for each chunk
	for i, chunk := range d.Chunks {
		wg.Add(1)
		go func(chunkIndex int, chunkData ChunkData, chunkFile string) {
			defer wg.Done()

			// Check for existing partial chunk
			resumeOffset, err := d.detectChunkResumeOffset(chunkFile, chunkData.Size)
			if err != nil {
				errorChan <- fmt.Errorf("chunk %d resume detection failed: %v", chunkIndex, err)
				return
			}

			// Skip if chunk is already complete
			if resumeOffset >= chunkData.Size {
				atomic.AddInt64(&totalCompletedBytes, chunkData.Size)
				d.Chunks[chunkIndex].IsCompleted = true
				if d.Callbacks != nil && d.Callbacks.OnChunkFinish != nil {
					d.Callbacks.OnChunkFinish(d, chunkIndex, chunkData.Start, chunkData.End, chunkData.Size)
				}
				return
			}

			// Download chunk
			if err := d.downloadSingleChunk(ctx, chunkIndex, chunkData, chunkFile, resumeOffset, &totalCompletedBytes); err != nil {
				errorChan <- fmt.Errorf("chunk %d download failed: %v", chunkIndex, err)
				return
			}

		}(i, chunk, chunkFileNames[i])
	}

	// Monitor progress and wait for completion
	go d.monitorMultiStreamProgress(ctx, &totalCompletedBytes)

	// Wait for all chunks to complete
	wg.Wait()
	close(errorChan)

	// Check for errors
	if len(errorChan) > 0 {
		return <-errorChan
	}

	return nil
}

// downloadSingleChunk downloads a single chunk with progress tracking and pause support.
//
// Parameters:
//   - ctx: Context for cancellation
//   - chunkIndex: Index of the chunk
//   - chunkData: Chunk metadata
//   - chunkFile: Path to chunk file
//   - resumeOffset: Byte offset to resume from
//   - totalCompletedBytes: Pointer to atomic counter for total progress
//
// Returns:
//   - error: Error if chunk download fails
func (d *Downloader) downloadSingleChunk(ctx context.Context, chunkIndex int, chunkData ChunkData, chunkFile string, resumeOffset int64, totalCompletedBytes *int64) error {
	// Call chunk start callback
	if d.Callbacks != nil && d.Callbacks.OnChunkStart != nil {
		d.Callbacks.OnChunkStart(d, chunkIndex, chunkData.Start, chunkData.End)
	}

	// Create HTTP client with appropriate timeouts
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

	// Calculate actual range to download
	startByte := chunkData.Start + resumeOffset
	endByte := chunkData.End

	// Create request with range header
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

	// Set range header for this chunk
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Open chunk file for writing
	file, err := d.openChunkFile(chunkFile, resumeOffset)
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %v", err)
	}
	defer file.Close()

	// Download chunk with progress tracking
	bytesWritten, err := d.downloadChunkWithProgress(ctx, chunkIndex, resp.Body, file, chunkData.Size-resumeOffset, totalCompletedBytes)
	if err != nil {
		if d.Callbacks != nil && d.Callbacks.OnChunkError != nil {
			d.Callbacks.OnChunkError(d, chunkIndex, chunkData.Start, chunkData.End, err)
		}
		return err
	}

	// Mark chunk as completed
	d.Chunks[chunkIndex].IsCompleted = true

	// Call chunk finish callback
	if d.Callbacks != nil && d.Callbacks.OnChunkFinish != nil {
		d.Callbacks.OnChunkFinish(d, chunkIndex, chunkData.Start, chunkData.End, bytesWritten)
	}

	return nil
}

// detectChunkResumeOffset checks if there's a partial chunk and returns the resume offset.
//
// Parameters:
//   - chunkFile: Path to the chunk file
//   - expectedSize: Expected size of the complete chunk
//
// Returns:
//   - int64: Byte offset to resume from (0 if starting fresh)
//   - error: Error if offset detection fails
func (d *Downloader) detectChunkResumeOffset(chunkFile string, expectedSize int64) (int64, error) {
	if !ufs.FileExists(chunkFile) {
		return 0, nil
	}

	fileInfo, err := os.Stat(chunkFile)
	if err != nil {
		return 0, nil // Start fresh if we can't get file info
	}

	currentSize := fileInfo.Size()
	if currentSize >= expectedSize {
		return expectedSize, nil // Chunk is complete
	}

	return currentSize, nil // Resume from current position
}

// openChunkFile opens a chunk file for writing, handling resume scenarios.
//
// Parameters:
//   - chunkFile: Path to the chunk file
//   - resumeOffset: Byte offset to resume from
//
// Returns:
//   - *os.File: File handle for writing
//   - error: Error if file opening fails
func (d *Downloader) openChunkFile(chunkFile string, resumeOffset int64) (*os.File, error) {
	if resumeOffset > 0 {
		// Open for appending
		return os.OpenFile(chunkFile, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		// Create new file
		return os.Create(chunkFile)
	}
}

// downloadChunkWithProgress downloads chunk data with pause support and progress tracking.
//
// Parameters:
//   - ctx: Context for cancellation
//   - chunkIndex: Index of the chunk
//   - reader: Source reader (response body)
//   - writer: Destination writer (chunk file)
//   - expectedBytes: Expected number of bytes to download
//   - totalCompletedBytes: Pointer to atomic counter for total progress
//
// Returns:
//   - int64: Number of bytes actually written
//   - error: Error if download fails
func (d *Downloader) downloadChunkWithProgress(ctx context.Context, chunkIndex int, reader io.Reader, writer io.Writer, expectedBytes int64, totalCompletedBytes *int64) (int64, error) {
	buffer := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	for totalWritten < expectedBytes {
		// Check for pause
		d.checkPauseState()

		// Check for cancellation
		select {
		case <-ctx.Done():
			return totalWritten, ctx.Err()
		default:
		}

		// Read data
		n, err := reader.Read(buffer)
		if n > 0 {
			// Write data
			written, writeErr := writer.Write(buffer[:n])
			if writeErr != nil {
				return totalWritten, fmt.Errorf("failed to write chunk data: %v", writeErr)
			}

			totalWritten += int64(written)

			// Update total progress atomically
			atomic.AddInt64(totalCompletedBytes, int64(written))
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return totalWritten, fmt.Errorf("failed to read chunk data: %v", err)
		}
	}

	return totalWritten, nil
}

// monitorMultiStreamProgress monitors overall download progress and triggers callbacks.
//
// Parameters:
//   - ctx: Context for cancellation
//   - totalCompletedBytes: Pointer to atomic counter for completed bytes
func (d *Downloader) monitorMultiStreamProgress(ctx context.Context, totalCompletedBytes *int64) {
	ticker := time.NewTicker(500 * time.Millisecond) // Update every 500ms
	defer ticker.Stop()

	var lastReported int64
	lastReportTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current := atomic.LoadInt64(totalCompletedBytes)
			now := time.Now()

			// Calculate speed
			elapsed := now.Sub(lastReportTime).Seconds()
			if elapsed >= 1.0 { // Update speed every second
				bytesDiff := current - lastReported
				speed := float64(bytesDiff) / elapsed

				// Update progress tracker
				d.Progress.mu.Lock()
				d.Progress.BytesCompleted = current
				d.Progress.SpeedBps = speed
				d.Progress.LastReported = now
				if d.ServerHeaders.Filesize > 0 {
					d.Progress.Percentage = float64(current) / float64(d.ServerHeaders.Filesize) * 100
				}
				d.Progress.mu.Unlock()

				// Call progress callback
				if d.Callbacks != nil && d.Callbacks.OnProgress != nil {
					d.Callbacks.OnProgress(d)
				}

				lastReported = current
				lastReportTime = now
			}
		}
	}
}

// mergeChunksToFinalFile merges all chunk files into the final output file.
//
// Parameters:
//   - chunkFileNames: Array of chunk file paths in order
//
// Returns:
//   - error: Error if merging fails
func (d *Downloader) mergeChunksToFinalFile(chunkFileNames []string) error {
	// Call assemble start callback
	if d.Callbacks != nil && d.Callbacks.OnAssembleStart != nil {
		d.Callbacks.OnAssembleStart(d)
	}

	// Use the UFS merge function
	err := ufs.MergeChunkFiles(chunkFileNames, d.fileInfo.FullPath)
	if err != nil {
		if d.Callbacks != nil && d.Callbacks.OnAssembleError != nil {
			d.Callbacks.OnAssembleError(d, err)
		}
		return err
	}

	// Call assemble finish callback
	if d.Callbacks != nil && d.Callbacks.OnAssembleFinish != nil {
		d.Callbacks.OnAssembleFinish(d)
	}

	return nil
}
