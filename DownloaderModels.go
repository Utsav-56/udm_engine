package main

import (
	"os"
	"sync"
	"time"
)

type UserPreferences struct {
	DownloadDir string
	fileName    string
	threadCount int
	maxRetries  int
}

type CustomHeaders struct {
	Cookies string
	Headers map[string]string
}

// ChunkData contains information about a chunk of data being downloaded
// It is different from ChunkInfo as it represents dtata for a specific chunk
type ChunkData struct {
	Index int   // Index of the chunk acts as a unique identifier
	Start int64 // Start byte of the chunk
	End   int64 // End byte of the chunk
	Size  int64 // Size of the chunk in bytes it is total size of the chunk expected to be downloaded

	IsCompleted bool // Whether the chunk has been successfully downloaded

}

// TimeInfo contains time-related information for the download
// like start time, end time, and elapsed time
// It is used to track the duration of the download process
type TimeInfo struct {
	StartTime time.Time     // Time when the download started
	EndTime   time.Time     // Time when the download ended
	Elapsed   time.Duration // Total time taken for the download
}

// PauseController is used to manage the pause and resume functionality
// It uses a mutex and condition variable to handle pausing and resuming
type PauseController struct {
	mu       sync.Mutex
	cond     *sync.Cond
	isPaused bool
}

// NewPauseController creates a new PauseController instance.
//
// Returns:
//   - *PauseController: Initialized pause controller
func NewPauseController() *PauseController {
	pc := &PauseController{
		isPaused: false,
	}
	pc.cond = sync.NewCond(&pc.mu)
	return pc
}

// Pause sets the controller to paused state.
func (pc *PauseController) Pause() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.isPaused = true
}

// Resume sets the controller to resumed state and wakes up waiting goroutines.
func (pc *PauseController) Resume() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.isPaused = false
	pc.cond.Broadcast()
}

// IsPaused returns the current pause state.
//
// Returns:
//   - bool: True if paused, false if running
func (pc *PauseController) IsPaused() bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.isPaused
}

// WaitIfPaused blocks the calling goroutine while the controller is paused.
func (pc *PauseController) WaitIfPaused() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for pc.isPaused {
		pc.cond.Wait()
	}
}

// Fileinfo contains the final info of file it is actual file path where it is downloaded
type FileInfo struct {
	Dir      string
	Name     string
	FullPath string
}

// Callbacks contains all callback functions for download events
type Callbacks struct {
	OnProgress func(d *Downloader)
	OnFinish   func(d *Downloader)
	OnError    func(d *Downloader, err error)

	OnStart  func(d *Downloader)
	OnStop   func(d *Downloader)
	OnPause  func(d *Downloader)
	OnResume func(d *Downloader)

	OnAssembleStart  func(d *Downloader)
	OnAssembleFinish func(d *Downloader)
	OnAssembleError  func(d *Downloader, err error)

	OnChunkStart  func(d *Downloader, chunkIndex int, start, end int64)
	OnChunkFinish func(d *Downloader, chunkIndex int, start, end int64, bytesWritten int64)
	OnChunkError  func(d *Downloader, chunkIndex int, start, end int64, err error)

	OnDispose func(d *Downloader)
}

type Downloader struct {
	Url           string
	ID            string
	fileInfo      FileInfo
	Prefs         UserPreferences
	Headers       CustomHeaders
	ServerHeaders ServerData
	Chunks        []ChunkData
	ChunkManager  *ChunkManager

	PauseControl *PauseController
	Progress     *ProgressTracker
	Callbacks    *Callbacks
	TimeStats    *TimeInfo
	Status       string
	Error        error
	OutputPath   string

	// Progress bar support
	ChunkProgress  []ChunkProgressData // Progress tracking for individual chunks
	UseProgressBar bool                // Whether to show progress bar instead of text output
}

// Download statuses
// These constants represent the various states a download can be in
// They are used to track the current state of the download process
const (
	DOWNLOAD_QUEUED      = "queued"
	DOWNLOAD_IN_PROGRESS = "in_progress"
	DOWNLOAD_PAUSED      = "paused"
	DOWNLOAD_COMPLETED   = "completed"
	DOWNLOAD_FAILED      = "failed"
	DOWNLOAD_STOPPED     = "stopped"
)

type ChunkTask struct {
	Chunk      ChunkData
	URL        string
	Headers    map[string]string
	OutputFile *os.File
}

type ChunkManager struct {
	Chunks         []ChunkData
	ChunkSize      int64
	TotalSize      int64
	CompletedBytes int64
	mutex          sync.Mutex
}
type Worker struct {
	ID       int
	Task     ChunkTask
	RetryMax int
	Callback func(index int, err error)
}

type ProgressTracker struct {
	mu             sync.Mutex
	BytesCompleted int64         // Total bytes downloaded so far
	TotalBytes     int64         // Total file size (if known)
	LastReported   time.Time     // Last time progress was reported
	LastCheckTime  time.Time     // Last time progress was checked
	SpeedBps       float64       // Current download speed in bytes per second
	Percentage     float64       // Download completion percentage (0-100)
	ETA            time.Duration // Estimated time remaining
	BytesPerSecond int64         // Average bytes per second since start
	StartTime      time.Time     // When download started

	// Progress bar integration
	ProgressModel interface{} // Will hold the UDM progress model
	ShowProgress  bool        // Whether to show progress bar
}

// ChunkProgressData represents progress for individual chunks in multi-stream downloads
type ChunkProgressData struct {
	Index           int
	Percentage      float64
	IsComplete      bool
	BytesDownloaded int64
	TotalBytes      int64
}

// UpdateProgress updates the progress tracker with new data
// and calculates derived metrics like speed, percentage, and ETA.
//
// Parameters:
//   - bytesRead: Number of new bytes downloaded
//   - totalSize: Total file size (if known, 0 if unknown)
func (pt *ProgressTracker) UpdateProgress(bytesRead int64, totalSize int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	now := time.Now()

	// Initialize start time if not set
	if pt.StartTime.IsZero() {
		pt.StartTime = now
	}

	// Update basic metrics
	pt.BytesCompleted += bytesRead
	pt.TotalBytes = totalSize
	pt.LastCheckTime = now

	// Calculate percentage if total size is known
	if totalSize > 0 {
		pt.Percentage = float64(pt.BytesCompleted) / float64(totalSize) * 100
		if pt.Percentage > 100 {
			pt.Percentage = 100
		}
	}

	// Calculate speed (only if we have a previous report time)
	if !pt.LastReported.IsZero() {
		elapsed := now.Sub(pt.LastReported).Seconds()
		if elapsed > 0 {
			pt.SpeedBps = float64(bytesRead) / elapsed
		}
	}

	// Calculate average speed since start
	totalElapsed := now.Sub(pt.StartTime).Seconds()
	if totalElapsed > 0 {
		pt.BytesPerSecond = int64(float64(pt.BytesCompleted) / totalElapsed)
	}

	// Calculate ETA if we have speed and total size
	if pt.SpeedBps > 0 && totalSize > 0 && pt.BytesCompleted < totalSize {
		remainingBytes := totalSize - pt.BytesCompleted
		etaSeconds := float64(remainingBytes) / pt.SpeedBps
		pt.ETA = time.Duration(etaSeconds) * time.Second
	}

	pt.LastReported = now
}

// GetProgressInfo returns current progress information in a thread-safe manner.
//
// Returns:
//   - bytesCompleted: Number of bytes downloaded
//   - totalBytes: Total file size (0 if unknown)
//   - percentage: Completion percentage (0-100)
//   - speedBps: Current speed in bytes per second
//   - eta: Estimated time remaining
func (pt *ProgressTracker) GetProgressInfo() (bytesCompleted, totalBytes int64, percentage, speedBps float64, eta time.Duration) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	return pt.BytesCompleted, pt.TotalBytes, pt.Percentage, pt.SpeedBps, pt.ETA
}

func (d *Downloader) getUserPreferredFilename() string {
	return d.Prefs.fileName
}

func (d *Downloader) getDownloadDirectory() string {
	return d.Prefs.DownloadDir
}

func (d *Downloader) getThreadCount() int {
	return d.Prefs.threadCount
}

func (d *Downloader) getRetryCount() int {
	return d.Prefs.maxRetries
}

// EnableProgressBar enables the visual progress bar display
func (d *Downloader) EnableProgressBar() {
	d.UseProgressBar = true
	d.Progress.ShowProgress = true
}

// DisableProgressBar disables the visual progress bar and uses text output
func (d *Downloader) DisableProgressBar() {
	d.UseProgressBar = false
	d.Progress.ShowProgress = false
}

// InitializeChunkProgress initializes chunk progress tracking for multi-stream downloads
func (d *Downloader) InitializeChunkProgress(chunkCount int) {
	d.ChunkProgress = make([]ChunkProgressData, chunkCount)
	for i := range d.ChunkProgress {
		d.ChunkProgress[i] = ChunkProgressData{
			Index:           i,
			Percentage:      0.0,
			IsComplete:      false,
			BytesDownloaded: 0,
			TotalBytes:      0,
		}
	}
}

// UpdateChunkProgress updates progress for a specific chunk
func (d *Downloader) UpdateChunkProgress(chunkIndex int, bytesDownloaded, totalBytes int64) {
	if chunkIndex >= 0 && chunkIndex < len(d.ChunkProgress) {
		d.ChunkProgress[chunkIndex].BytesDownloaded = bytesDownloaded
		d.ChunkProgress[chunkIndex].TotalBytes = totalBytes

		if totalBytes > 0 {
			d.ChunkProgress[chunkIndex].Percentage = float64(bytesDownloaded) / float64(totalBytes) * 100
		}

		d.ChunkProgress[chunkIndex].IsComplete = (bytesDownloaded >= totalBytes && totalBytes > 0)
	}
}

// GetChunkProgressData returns the current chunk progress for display
func (d *Downloader) GetChunkProgressData() []ChunkProgressData {
	return d.ChunkProgress
}

// IsMultiStreamDownload returns true if this is a multi-stream download
func (d *Downloader) IsMultiStreamDownload() bool {
	return len(d.ChunkProgress) > 1
}
