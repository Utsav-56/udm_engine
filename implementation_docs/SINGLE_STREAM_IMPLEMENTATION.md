# UDM Single Stream Download Implementation

## Overview

The single-stream download implementation provides a robust, feature-rich download engine with pause/resume/cancel functionality, progress tracking, and automatic server capability detection. This implementation serves as the foundation for UDM's download capabilities and includes provisions for future elevation to multi-stream downloads.

## Architecture

### Core Components

1. **DownloadSingleStream.go** - Main single-threaded download engine
2. **DownloaderModels.go** - Data structures and progress tracking
3. **StartDownload.go** - Download orchestration and strategy selection
4. **ServerHeaders.go** - Server capability analysis
5. **ufs/** - File system utilities for chunk management

### Key Features Implemented

#### ✅ **Play/Pause/Cancel/Resume Functionality**

- Thread-safe pause/resume using mutex and condition variables
- Context-based cancellation for clean shutdown
- Resume capability using HTTP Range requests
- State management with proper status tracking

#### ✅ **Progress Tracking**

- Real-time progress updates with percentage calculation
- Speed calculation (bytes per second)
- ETA (Estimated Time of Arrival) computation
- Thread-safe progress reporting

#### ✅ **Callback System**

- Comprehensive callback support for all download events
- OnStart, OnProgress, OnPause, OnResume, OnFinish, OnError, OnStop
- Event-driven architecture for UI integration

#### ✅ **Server Analysis**

- Initial HEAD request for metadata
- Concurrent header analysis during download
- Range request support detection
- Filename extraction with RFC 6266 compliance

#### ✅ **Smart File Handling**

- User preference priority system
- OS default download directory support
- Unique filename generation to prevent overwrites
- Proper error handling and cleanup

#### ✅ **Elevation to Multi-Stream (Prepared)**

- Concurrent header analysis during download
- Conditions check for multi-stream eligibility
- Framework ready for future multi-stream implementation

## Implementation Details

### Download Process Flow

```
1. StartDownload()
   ├── Initialize download session
   ├── Prefetch server metadata
   ├── Check user preferences
   └── Execute download strategy

2. DownloadSingleStream()
   ├── Initialize single stream session
   ├── Setup output file and resume detection
   ├── Start concurrent header analysis
   ├── Begin download with progress tracking
   ├── Monitor for multi-stream elevation
   ├── Handle pause/resume/cancel operations
   └── Finalize download and cleanup
```

### Concurrent Operations

1. **Main Download Thread**: Handles actual file downloading
2. **Header Analysis Thread**: Analyzes server capabilities during download
3. **Progress Tracking**: Thread-safe progress updates with callbacks

### File Management

#### Directory Resolution Priority:

1. User-specified download directory
2. OS default Downloads folder (`~/Downloads`)
3. Current working directory (fallback)

#### Filename Resolution Priority:

1. User-specified filename
2. Server-provided filename (Content-Disposition header)
3. Filename from URL path
4. Generated name with MIME type extension
5. Default "downloaded_file" fallback

### Progress Tracking System

The `ProgressTracker` provides comprehensive download metrics:

```go
type ProgressTracker struct {
    BytesCompleted  int64     // Total bytes downloaded
    TotalBytes      int64     // Total file size (if known)
    Percentage      float64   // Completion percentage (0-100)
    SpeedBps        float64   // Current speed in bytes/second
    ETA             time.Duration // Estimated time remaining
    BytesPerSecond  int64     // Average speed since start
    StartTime       time.Time // Download start time
    LastCheckTime   time.Time // Last progress update
}
```

### Pause/Resume Implementation

The `PauseController` uses Go's sync primitives for thread-safe operation:

```go
type PauseController struct {
    mu       sync.Mutex
    cond     *sync.Cond  // Condition variable for blocking/unblocking
    isPaused bool
}
```

**Key Methods:**

- `Pause()`: Sets pause state and blocks download
- `Resume()`: Clears pause state and wakes up download
- `WaitIfPaused()`: Blocks calling goroutine while paused

### Resume Capability

The implementation supports resume through:

1. **File Size Detection**: Check existing partial file size
2. **Range Request**: Use HTTP Range header to resume from offset
3. **Server Validation**: Verify server supports range requests
4. **Fallback Strategy**: Start fresh if resume is not possible

## Usage Examples

### Basic Download

```go
downloader := &Downloader{
    Url: "https://example.com/file.zip",
    Callbacks: &Callbacks{
        OnFinish: func(d *Downloader) {
            fmt.Printf("Downloaded: %s\n", d.fileInfo.FullPath)
        },
    },
}
downloader.StartDownload()
```

### Advanced Download with Control

```go
downloader := &Downloader{
    Url: "https://example.com/largefile.zip",
    Prefs: UserPreferences{
        DownloadDir: "./downloads",
        fileName:    "custom-name.zip",
    },
    Callbacks: &Callbacks{
        OnProgress: func(d *Downloader) {
            completed, percentage, speed := d.GetProgress()
            fmt.Printf("Progress: %.1f%% | Speed: %.2f KB/s\n",
                percentage, speed/1024)
        },
    },
}

// Start download
go downloader.StartDownload()

// Control operations
time.Sleep(5 * time.Second)
downloader.Pause()              // Pause download
time.Sleep(2 * time.Second)
downloader.Resume()             // Resume download
```

### Resume from Partial Download

```go
// If file exists partially, download will automatically resume
downloader := &Downloader{
    Url: "https://example.com/file.zip",
    Prefs: UserPreferences{
        DownloadDir: "./downloads",
        fileName:    "existing-partial-file.zip",
    },
}
downloader.StartDownload() // Automatically resumes if file exists
```

## Integration Points

### Callback Integration

All major events trigger callbacks for UI integration:

```go
type Callbacks struct {
    OnStart     func(d *Downloader)              // Download started
    OnProgress  func(d *Downloader)              // Progress update
    OnPause     func(d *Downloader)              // Download paused
    OnResume    func(d *Downloader)              // Download resumed
    OnFinish    func(d *Downloader)              // Download completed
    OnError     func(d *Downloader, err error)   // Error occurred
    OnStop      func(d *Downloader)              // Download cancelled
}
```

### Error Handling

Comprehensive error handling covers:

- Network connectivity issues
- File system errors
- Server response errors
- Timeout handling
- Resume failures

### Status Management

Download status is tracked through constants:

```go
const (
    DOWNLOAD_QUEUED      = "queued"
    DOWNLOAD_IN_PROGRESS = "in_progress"
    DOWNLOAD_PAUSED      = "paused"
    DOWNLOAD_COMPLETED   = "completed"
    DOWNLOAD_FAILED      = "failed"
    DOWNLOAD_STOPPED     = "stopped"
)
```

## Future Enhancements

### Multi-Stream Elevation

The framework is prepared for automatic elevation to multi-stream downloads:

1. **Trigger Conditions**:
    - Server supports range requests
    - File size > 10MB
    - Less than 25% of file downloaded
2. **Implementation Plan**:
    - Pause current single-stream download
    - Switch to multi-threaded chunk download
    - Merge completed chunks with partial download

### Performance Optimizations

- **Adaptive Buffer Sizing**: Dynamic buffer size based on network conditions
- **Connection Pooling**: Reuse HTTP connections for better performance
- **Compression Support**: Handle gzip/deflate compressed downloads
- **Bandwidth Throttling**: Configurable speed limits

### Enhanced Features

- **Download Scheduling**: Queue management and scheduled downloads
- **Retry Logic**: Sophisticated retry strategies for different error types
- **Mirror Support**: Automatic failover to mirror servers
- **Integrity Checking**: MD5/SHA256 checksum validation

## API Reference

### Main Methods

#### `StartDownload()`

Initiates the download process with server analysis and strategy selection.

#### `DownloadSingleStream()`

Executes single-threaded download with all features.

#### `Pause()`, `Resume()`, `Cancel()`

Control download execution state.

#### `GetProgress() (bytesCompleted int64, percentage float64, speedBps float64)`

Retrieves current download progress information.

### Configuration

#### `UserPreferences`

```go
type UserPreferences struct {
    DownloadDir string // Target directory
    fileName    string // Custom filename
    threadCount int    // Number of threads (future use)
    maxRetries  int    // Retry attempts
}
```

#### `CustomHeaders`

```go
type CustomHeaders struct {
    Headers map[string]string // Custom HTTP headers
    Cookies string            // Cookie string
}
```

## Testing and Validation

The implementation includes comprehensive examples and test scenarios:

1. **SimpleDownloadExample**: Basic functionality test
2. **SingleStreamExample**: Full-featured demonstration
3. **AdvancedUsageExample**: Resume and error handling
4. **Control Demonstrations**: Pause/resume/cancel operations

## Conclusion

This single-stream download implementation provides a solid foundation for UDM's download capabilities. It includes all requested features and maintains high code quality with proper documentation, error handling, and extensibility for future enhancements.

The modular design allows for easy integration with UI components through the callback system, while the robust error handling ensures reliable operation across various network conditions and server configurations.
