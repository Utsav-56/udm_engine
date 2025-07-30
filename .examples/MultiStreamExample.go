package main

import (
	"fmt"
	"time"
)

// MultiStreamExample demonstrates how to use the multi-threaded download functionality
// with all the available features including pause/resume/cancel and callbacks.
func MultiStreamExample() {
	// Create downloader instance for a large file
	downloader := &Downloader{
		Url: "https://releases.ubuntu.com/20.04/ubuntu-20.04.6-desktop-amd64.iso",
		ID:  "multi-stream-example",

		// User preferences
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "ubuntu-multi-stream.iso",
			threadCount: 8, // Use 8 threads for multi-stream
			maxRetries:  3,
		},

		// Custom headers (optional)
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-MultiStream-Manager/1.0",
			},
		},

		// Setup comprehensive callbacks
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("🚀 Multi-stream download started: %s\n", d.Url)
				fmt.Printf("📁 Output: %s\n", d.fileInfo.FullPath)
				if d.ServerHeaders.Filesize > 0 {
					fmt.Printf("📊 File size: %.2f MB\n", float64(d.ServerHeaders.Filesize)/1024/1024)
				}
				fmt.Printf("🧵 Threads: %d\n", d.getThreadCount())
				fmt.Printf("📡 Range support: %v\n", d.ServerHeaders.AcceptsRanges)
			},

			OnProgress: func(d *Downloader) {
				completed, total, percentage, speed, eta := d.Progress.GetProgressInfo()
				fmt.Printf("\r🔄 Progress: %.1f%% | %.2f/%.2f MB | %.2f KB/s | ETA: %v",
					percentage,
					float64(completed)/1024/1024,
					float64(total)/1024/1024,
					speed/1024,
					eta.Round(time.Second))
			},

			OnPause: func(d *Downloader) {
				fmt.Println("\n⏸️  Multi-stream download paused")
			},

			OnResume: func(d *Downloader) {
				fmt.Println("▶️  Multi-stream download resumed")
			},

			OnChunkStart: func(d *Downloader, chunkIndex int, start, end int64) {
				fmt.Printf("\n🔗 Chunk %d started: bytes %d-%d (%.2f MB)\n",
					chunkIndex, start, end, float64(end-start+1)/1024/1024)
			},

			OnChunkFinish: func(d *Downloader, chunkIndex int, start, end int64, bytesWritten int64) {
				fmt.Printf("✅ Chunk %d completed: %d bytes written\n", chunkIndex, bytesWritten)
			},

			OnChunkError: func(d *Downloader, chunkIndex int, start, end int64, err error) {
				fmt.Printf("❌ Chunk %d error: %v\n", chunkIndex, err)
			},

			OnAssembleStart: func(d *Downloader) {
				fmt.Println("\n🔧 Merging chunks into final file...")
			},

			OnAssembleFinish: func(d *Downloader) {
				fmt.Println("✅ Chunks merged successfully!")
			},

			OnAssembleError: func(d *Downloader, err error) {
				fmt.Printf("❌ Chunk assembly failed: %v\n", err)
			},

			OnFinish: func(d *Downloader) {
				fmt.Printf("\n🎉 Multi-stream download completed successfully!\n")
				fmt.Printf("📁 File: %s\n", d.fileInfo.FullPath)
				fmt.Printf("📊 Total size: %.2f MB\n", float64(d.Progress.BytesCompleted)/1024/1024)
				fmt.Printf("⏱️  Duration: %v\n", d.TimeStats.Elapsed)
				fmt.Printf("🚀 Average speed: %.2f KB/s\n", float64(d.Progress.BytesPerSecond)/1024)
			},

			OnError: func(d *Downloader, err error) {
				fmt.Printf("\n❌ Multi-stream download failed: %v\n", err)
			},

			OnStop: func(d *Downloader) {
				fmt.Println("\n🛑 Multi-stream download cancelled")
			},
		},
	}

	// Demonstrate download control
	go demonstrateMultiStreamControl(downloader)

	// Start the download
	downloader.StartDownload()
}

// demonstrateMultiStreamControl shows how to control multi-stream downloads
func demonstrateMultiStreamControl(d *Downloader) {
	// Wait for download to start and chunks to initialize
	time.Sleep(5 * time.Second)

	// Pause the download
	fmt.Println("\n🔄 Testing multi-stream pause...")
	d.Pause()

	// Wait a bit
	time.Sleep(3 * time.Second)

	// Resume the download
	fmt.Println("🔄 Testing multi-stream resume...")
	d.Resume()

	// Let it run for a while
	time.Sleep(30 * time.Second)

	// Optionally cancel (uncomment to test cancellation)
	// fmt.Println("🔄 Testing multi-stream cancel...")
	// d.Cancel()
}

// ComparisonExample demonstrates the difference between single-stream and multi-stream
func ComparisonExample() {
	testURL := "https://httpbin.org/bytes/104857600" // 100MB test file

	fmt.Println("🆚 Single-Stream vs Multi-Stream Comparison")
	fmt.Println("===========================================")

	// Test single-stream download
	fmt.Println("\n1️⃣  Testing Single-Stream Download:")
	singleStreamDownloader := &Downloader{
		Url: testURL,
		ID:  "single-stream-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "test-single-stream.bin",
			threadCount: 1, // Force single stream
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("📥 Single-stream started\n")
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Single-stream completed in %v\n", d.TimeStats.Elapsed)
			},
		},
	}

	singleStreamDownloader.StartDownload()

	// Small delay between tests
	time.Sleep(2 * time.Second)

	// Test multi-stream download
	fmt.Println("\n8️⃣  Testing Multi-Stream Download:")
	multiStreamDownloader := &Downloader{
		Url: testURL,
		ID:  "multi-stream-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "test-multi-stream.bin",
			threadCount: 8, // Use 8 threads
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("📥 Multi-stream started with %d threads\n", d.getThreadCount())
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Multi-stream completed in %v\n", d.TimeStats.Elapsed)
			},
		},
	}

	multiStreamDownloader.StartDownload()
}

// ResumeCapabilityExample demonstrates resume functionality for multi-stream downloads
func ResumeCapabilityExample() {
	fmt.Println("🔄 Multi-Stream Resume Capability Test")
	fmt.Println("=====================================")

	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/52428800", // 50MB test file
		ID:  "resume-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "resume-test-file.bin",
			threadCount: 4,
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("🔄 Resume test started\n")
			},
			OnChunkStart: func(d *Downloader, chunkIndex int, start, end int64) {
				fmt.Printf("🔗 Chunk %d: %d-%d bytes\n", chunkIndex, start, end)
			},
			OnProgress: func(d *Downloader) {
				_, _, percentage, _, _ := d.Progress.GetProgressInfo()
				if int(percentage)%10 == 0 { // Print every 10%
					fmt.Printf("📊 Progress: %.0f%%\n", percentage)
				}
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Resume test completed!\n")
			},
		},
	}

	// Simulate interruption and resume
	go func() {
		time.Sleep(3 * time.Second) // Let some progress happen
		fmt.Println("⏹️  Simulating interruption...")
		downloader.Cancel()

		time.Sleep(2 * time.Second)
		fmt.Println("🔄 Restarting download (should resume from partial chunks)...")

		// Restart download - should resume from existing chunks
		downloader.Status = DOWNLOAD_QUEUED
		downloader.StartDownload()
	}()

	downloader.StartDownload()
}

// runMultiStreamExamples is the main function to run examples
func runMultiStreamExamples() {
	fmt.Println("🧵 UDM Multi-Stream Download Examples")
	fmt.Println("====================================")

	// Choose which example to run:

	// MultiStreamExample()           // Full multi-stream example
	// ComparisonExample()            // Single vs Multi comparison
	ResumeCapabilityExample() // Resume functionality test
}
