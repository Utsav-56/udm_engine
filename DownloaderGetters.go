package udm

import (
	"time"
)

// Getter methods for Downloader struct
// These methods provide safe access to downloader fields and computed values

// GetID returns the unique identifier of the downloader
func (d *Downloader) GetID() string {
	return d.ID
}

// GetURL returns the download URL
func (d *Downloader) GetURL() string {
	return d.Url
}

// GetStatus returns the current download status
func (d *Downloader) GetStatus() string {
	return d.Status
}

// GetProgressPercent returns the download completion percentage (0-100)
func (d *Downloader) GetProgressPercent() float64 {
	if d.Progress == nil {
		return 0.0
	}

	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	return d.Progress.Percentage
}

// GetDownloadedBytes returns the number of bytes downloaded so far
func (d *Downloader) GetDownloadedBytes() int64 {
	if d.Progress == nil {
		return 0
	}

	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	return d.Progress.BytesCompleted
}

// GetFileSize returns the total file size in bytes
func (d *Downloader) GetFileSize() int64 {
	if d.ServerHeaders.Filesize > 0 {
		return d.ServerHeaders.Filesize
	}

	if d.Progress != nil {
		d.Progress.mu.Lock()
		defer d.Progress.mu.Unlock()
		return d.Progress.TotalBytes
	}

	return 0
}

// GetCurrentSpeed returns the current download speed in bytes per second
func (d *Downloader) GetCurrentSpeed() float64 {
	if d.Progress == nil {
		return 0.0
	}

	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	return d.Progress.SpeedBps
}

// GetAverageSpeed returns the average download speed in bytes per second since start
func (d *Downloader) GetAverageSpeed() float64 {
	if d.Progress == nil {
		return 0.0
	}

	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	return float64(d.Progress.BytesPerSecond)
}

// GetETA returns the estimated time remaining for the download
func (d *Downloader) GetETA() time.Duration {
	if d.Progress == nil {
		return 0
	}

	d.Progress.mu.Lock()
	defer d.Progress.mu.Unlock()

	return d.Progress.ETA
}

// GetFilename returns the filename of the download
func (d *Downloader) GetFilename() string {
	if d.fileInfo.Name != "" {
		return d.fileInfo.Name
	}

	// Fallback to server headers filename
	if d.ServerHeaders.Filename != "" {
		return d.ServerHeaders.Filename
	}

	// Fallback to user preference
	if d.Prefs.fileName != "" {
		return d.Prefs.fileName
	}

	return "unknown_file"
}

// GetOutputDir returns the output directory for the download
func (d *Downloader) GetOutputDir() string {
	if d.fileInfo.Dir != "" {
		return d.fileInfo.Dir
	}

	// Fallback to user preference
	if d.Prefs.DownloadDir != "" {
		return d.Prefs.DownloadDir
	}

	return "./"
}

// GetFilePath returns the full file path of the download
func (d *Downloader) GetFilePath() string {
	if d.fileInfo.FullPath != "" {
		return d.fileInfo.FullPath
	}

	// Fallback to OutputPath
	if d.OutputPath != "" {
		return d.OutputPath
	}

	return ""
}

// GetTimeTaken returns the total time taken for the download
func (d *Downloader) GetTimeTaken() time.Duration {
	if d.TimeStats == nil {
		return 0
	}

	// If download is completed, return the elapsed time
	if !d.TimeStats.EndTime.IsZero() {
		return d.TimeStats.Elapsed
	}

	// If download is in progress, calculate current elapsed time
	if !d.TimeStats.StartTime.IsZero() {
		return time.Since(d.TimeStats.StartTime)
	}

	return 0
}

// GetStartTime returns when the download started
func (d *Downloader) GetStartTime() time.Time {
	if d.TimeStats != nil {
		return d.TimeStats.StartTime
	}

	if d.Progress != nil {
		d.Progress.mu.Lock()
		defer d.Progress.mu.Unlock()
		return d.Progress.StartTime
	}

	return time.Time{}
}

// GetEndTime returns when the download ended (if completed)
func (d *Downloader) GetEndTime() time.Time {
	if d.TimeStats != nil {
		return d.TimeStats.EndTime
	}

	return time.Time{}
}

// GetError returns the last error that occurred during download
func (d *Downloader) GetError() error {
	return d.Error
}

// IsCompleted returns true if the download is completed
func (d *Downloader) IsCompleted() bool {
	return d.Status == DOWNLOAD_COMPLETED
}

// IsPaused returns true if the download is paused
func (d *Downloader) IsPaused() bool {
	return d.Status == DOWNLOAD_PAUSED
}

// IsInProgress returns true if the download is in progress
func (d *Downloader) IsInProgress() bool {
	return d.Status == DOWNLOAD_IN_PROGRESS
}

// IsFailed returns true if the download has failed
func (d *Downloader) IsFailed() bool {
	return d.Status == DOWNLOAD_FAILED
}

// IsStopped returns true if the download was stopped/cancelled
func (d *Downloader) IsStopped() bool {
	return d.Status == DOWNLOAD_STOPPED
}

// GetThreadCount returns the number of threads used for multi-stream downloads
func (d *Downloader) GetThreadCount() int {
	return d.getThreadCount()
}

// GetRetryCount returns the maximum number of retries configured
func (d *Downloader) GetRetryCount() int {
	return d.getRetryCount()
}

// GetFileType returns the MIME type of the file
func (d *Downloader) GetFileType() string {
	return d.ServerHeaders.Filetype
}

// GetFileExtension returns the file extension
func (d *Downloader) GetFileExtension() string {
	// Extract extension from filename
	filename := d.GetFilename()
	if filename == "" {
		return ""
	}

	// Find the last dot in the filename
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}

	return ""
}

// SupportsRangeRequests returns true if the server supports range requests
func (d *Downloader) SupportsRangeRequests() bool {
	return d.ServerHeaders.AcceptsRanges
}

// GetFinalURL returns the final URL after all redirects
func (d *Downloader) GetFinalURL() string {
	if d.ServerHeaders.FinalURL != "" {
		return d.ServerHeaders.FinalURL
	}
	return d.Url
}
