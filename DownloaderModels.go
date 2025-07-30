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
	BytesCompleted int64
	LastReported   time.Time
	SpeedBps       float64
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
