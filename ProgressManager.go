package udm

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ProgressManager manages the progress bar display for downloads
type ProgressManager struct {
	downloader *Downloader
	program    *tea.Program
	model      *UDMProgressModel
	tracker    *UDMProgressTracker
	ctx        context.Context
	cancel     context.CancelFunc
	isRunning  bool
}

// NewProgressManager creates a new progress manager for the downloader
func NewProgressManager(downloader *Downloader) *ProgressManager {
	ctx, cancel := context.WithCancel(context.Background())

	tracker := &UDMProgressTracker{
		Filename:       downloader.fileInfo.Name,
		BytesCompleted: 0,
		TotalBytes:     downloader.ServerHeaders.Filesize,
		SpeedBps:       0,
		Percentage:     0,
		ETA:            0,
		StartTime:      time.Now(),
		IsPaused:       false,
		IsCompleted:    false,
		OutputDir:      downloader.fileInfo.Dir,
		IsMultiStream:  downloader.IsMultiStreamDownload(),
		ChunkProgress:  []ChunkProgress{},
	}

	model := NewUDMProgress(tracker)

	return &ProgressManager{
		downloader: downloader,
		model:      model,
		tracker:    tracker,
		ctx:        ctx,
		cancel:     cancel,
		isRunning:  false,
	}
}

// StartProgressDisplay starts the progress bar display in a separate goroutine
func (pm *ProgressManager) StartProgressDisplay() error {
	if pm.isRunning {
		return fmt.Errorf("progress display is already running")
	}

	// Initialize progress tracking for multi-stream downloads
	if pm.downloader.IsMultiStreamDownload() {
		pm.initializeChunkProgress()
	}

	// Create the Bubble Tea program
	pm.program = tea.NewProgram(pm.model, tea.WithAltScreen())

	// Start the program in a goroutine
	go func() {
		pm.isRunning = true
		defer func() { pm.isRunning = false }()

		if err := pm.program.Start(); err != nil {
			fmt.Printf("Error starting progress display: %v\n", err)
		}
	}()

	// Start the progress update loop
	go pm.updateLoop()

	return nil
}

// StopProgressDisplay stops the progress bar display
func (pm *ProgressManager) StopProgressDisplay() {
	if pm.cancel != nil {
		pm.cancel()
	}

	if pm.program != nil {
		pm.program.Quit()
	}
}

// updateLoop continuously updates the progress display
func (pm *ProgressManager) updateLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.updateProgress()
		}
	}
}

// updateProgress updates the progress tracker with current download data
func (pm *ProgressManager) updateProgress() {
	if pm.downloader.Progress == nil {
		return
	}

	// Get current progress data
	bytesCompleted, totalBytes, percentage, speedBps, eta := pm.downloader.Progress.GetProgressInfo()

	// Update tracker
	pm.tracker.BytesCompleted = bytesCompleted
	pm.tracker.TotalBytes = totalBytes
	pm.tracker.Percentage = percentage
	pm.tracker.SpeedBps = speedBps
	pm.tracker.ETA = eta
	pm.tracker.IsPaused = (pm.downloader.Status == DOWNLOAD_PAUSED)
	pm.tracker.IsCompleted = (pm.downloader.Status == DOWNLOAD_COMPLETED)

	// Update chunk progress for multi-stream downloads
	if pm.downloader.IsMultiStreamDownload() {
		pm.updateChunkProgress()
	}

	// Send update to the UI (if program is running)
	if pm.program != nil && pm.isRunning {
		pm.program.Send(progressUpdateMsg(*pm.tracker))
	}
}

// initializeChunkProgress sets up chunk progress tracking
func (pm *ProgressManager) initializeChunkProgress() {
	chunkCount := len(pm.downloader.Chunks)
	if chunkCount == 0 {
		return
	}

	pm.tracker.ChunkProgress = make([]ChunkProgress, chunkCount)
	for i := 0; i < chunkCount; i++ {
		pm.tracker.ChunkProgress[i] = ChunkProgress{
			Index:      i,
			Percentage: 0.0,
			IsComplete: false,
		}
	}
}

// updateChunkProgress updates individual chunk progress
func (pm *ProgressManager) updateChunkProgress() {
	downloadChunkProgress := pm.downloader.GetChunkProgressData()

	for i, chunkProgress := range downloadChunkProgress {
		if i < len(pm.tracker.ChunkProgress) {
			pm.tracker.ChunkProgress[i].Index = chunkProgress.Index
			pm.tracker.ChunkProgress[i].Percentage = chunkProgress.Percentage
			pm.tracker.ChunkProgress[i].IsComplete = chunkProgress.IsComplete
		}
	}
}

// MarkCompleted marks the download as completed and shows final message
func (pm *ProgressManager) MarkCompleted() {
	pm.tracker.IsCompleted = true
	pm.tracker.IsPaused = false

	if pm.program != nil && pm.isRunning {
		pm.program.Send(progressUpdateMsg(*pm.tracker))

		// Give some time for the completion message to display
		time.Sleep(3 * time.Second)
		pm.program.Quit()
	}
}

// MarkError marks the download as failed
func (pm *ProgressManager) MarkError(err error) {
	pm.tracker.IsCompleted = true
	pm.tracker.IsPaused = false

	// You could add error information to the tracker here
	if pm.program != nil && pm.isRunning {
		pm.program.Send(progressUpdateMsg(*pm.tracker))
		time.Sleep(2 * time.Second)
		pm.program.Quit()
	}
}

// SetupProgressCallbacks configures the downloader callbacks to work with progress bar
func SetupProgressCallbacks(downloader *Downloader, pm *ProgressManager) {
	// Store original callbacks
	originalCallbacks := downloader.Callbacks
	if originalCallbacks == nil {
		originalCallbacks = &Callbacks{}
	}

	// Enhanced callbacks that work with progress bar
	downloader.Callbacks = &Callbacks{
		OnStart: func(d *Downloader) {
			if !d.UseProgressBar {
				if originalCallbacks.OnStart != nil {
					originalCallbacks.OnStart(d)
				}
				return
			}

			// Start progress display
			if pm != nil {
				pm.StartProgressDisplay()
			}

			// Call original callback
			if originalCallbacks.OnStart != nil {
				originalCallbacks.OnStart(d)
			}
		},

		OnProgress: func(d *Downloader) {
			// Progress updates are handled automatically by the progress manager
			// Just call original callback for any additional logic
			if originalCallbacks.OnProgress != nil && !d.UseProgressBar {
				originalCallbacks.OnProgress(d)
			}
		},

		OnPause: func(d *Downloader) {
			// Progress bar will automatically show pause state
			if originalCallbacks.OnPause != nil && !d.UseProgressBar {
				originalCallbacks.OnPause(d)
			}
		},

		OnResume: func(d *Downloader) {
			// Progress bar will automatically show resume state
			if originalCallbacks.OnResume != nil && !d.UseProgressBar {
				originalCallbacks.OnResume(d)
			}
		},

		OnFinish: func(d *Downloader) {
			if d.UseProgressBar && pm != nil {
				pm.MarkCompleted()
			}

			// Call original callback
			if originalCallbacks.OnFinish != nil {
				originalCallbacks.OnFinish(d)
			}
		},

		OnError: func(d *Downloader, err error) {
			if d.UseProgressBar && pm != nil {
				pm.MarkError(err)
			}

			// Call original callback
			if originalCallbacks.OnError != nil {
				originalCallbacks.OnError(d, err)
			}
		},

		OnStop: func(d *Downloader) {
			if d.UseProgressBar && pm != nil {
				pm.StopProgressDisplay()
			}

			// Call original callback
			if originalCallbacks.OnStop != nil {
				originalCallbacks.OnStop(d)
			}
		},

		// Multi-stream specific callbacks
		OnChunkStart: func(d *Downloader, chunkIndex int, start, end int64) {
			if originalCallbacks.OnChunkStart != nil && !d.UseProgressBar {
				originalCallbacks.OnChunkStart(d, chunkIndex, start, end)
			}
		},

		OnChunkFinish: func(d *Downloader, chunkIndex int, start, end int64, bytesWritten int64) {
			// Update chunk progress
			if d.UseProgressBar {
				d.UpdateChunkProgress(chunkIndex, bytesWritten, end-start+1)
			}

			if originalCallbacks.OnChunkFinish != nil && !d.UseProgressBar {
				originalCallbacks.OnChunkFinish(d, chunkIndex, start, end, bytesWritten)
			}
		},

		OnChunkError: func(d *Downloader, chunkIndex int, start, end int64, err error) {
			if originalCallbacks.OnChunkError != nil && !d.UseProgressBar {
				originalCallbacks.OnChunkError(d, chunkIndex, start, end, err)
			}
		},

		OnAssembleStart: func(d *Downloader) {
			if originalCallbacks.OnAssembleStart != nil && !d.UseProgressBar {
				originalCallbacks.OnAssembleStart(d)
			}
		},

		OnAssembleFinish: func(d *Downloader) {
			if originalCallbacks.OnAssembleFinish != nil && !d.UseProgressBar {
				originalCallbacks.OnAssembleFinish(d)
			}
		},

		OnAssembleError: func(d *Downloader, err error) {
			if originalCallbacks.OnAssembleError != nil && !d.UseProgressBar {
				originalCallbacks.OnAssembleError(d, err)
			}
		},

		OnDispose: func(d *Downloader) {
			if d.UseProgressBar && pm != nil {
				pm.StopProgressDisplay()
			}

			if originalCallbacks.OnDispose != nil {
				originalCallbacks.OnDispose(d)
			}
		},
	}
}
