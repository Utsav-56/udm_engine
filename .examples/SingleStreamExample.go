package main

import (
	"fmt"
	"log"
	"time"
)

// SingleStreamExample demonstrates how to use the single-threaded download functionality
// with all the available features including pause/resume/cancel and callbacks.
func SingleStreamExample() {
	// Create downloader instance
	downloader := &Downloader{
		Url: "https://releases.ubuntu.com/20.04/ubuntu-20.04.6-desktop-amd64.iso",
		ID:  "example-download-001",

		// User preferences
		Prefs: UserPreferences{
			DownloadDir: "./downloads", // Optional: leave empty for OS default
			fileName:    "",            // Optional: leave empty to use server filename
			threadCount: 1,             // For single stream, this is ignored
			maxRetries:  3,
		},

		// Custom headers (optional)
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-Download-Manager/1.0",
			},
			Cookies: "", // Add cookies if needed
		},

		// Setup callbacks
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("Download started: %s\n", d.Url)
				fmt.Printf("Output file: %s\n", d.fileInfo.FullPath)
			},

			OnProgress: func(d *Downloader) {
				completed, percentage, speed := d.GetProgress()
				fmt.Printf("\rProgress: %.1f%% | Downloaded: %d bytes | Speed: %.2f KB/s",
					percentage, completed, speed/1024)
			},

			OnPause: func(d *Downloader) {
				fmt.Println("\nDownload paused")
			},

			OnResume: func(d *Downloader) {
				fmt.Println("Download resumed")
			},

			OnFinish: func(d *Downloader) {
				fmt.Printf("\nDownload completed successfully!")
				fmt.Printf("File saved to: %s\n", d.fileInfo.FullPath)
				fmt.Printf("Total time: %v\n", d.TimeStats.Elapsed)
			},

			OnError: func(d *Downloader, err error) {
				fmt.Printf("\nDownload failed: %v\n", err)
			},

			OnStop: func(d *Downloader) {
				fmt.Println("\nDownload cancelled by user")
			},
		},
	}

	// Demonstrate download control in a separate goroutine
	go demonstrateDownloadControl(downloader)

	// Start the download
	downloader.StartDownload()
}

// demonstrateDownloadControl shows how to control the download
// by pausing, resuming, and cancelling operations.
func demonstrateDownloadControl(d *Downloader) {
	// Wait for download to start
	time.Sleep(3 * time.Second)

	// Pause the download
	fmt.Println("\n[Control] Pausing download...")
	d.Pause()

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Resume the download
	fmt.Println("[Control] Resuming download...")
	d.Resume()

	// Let it run for a while
	time.Sleep(10 * time.Second)

	// Optionally cancel (uncomment to test cancellation)
	// fmt.Println("[Control] Cancelling download...")
	// d.Cancel()
}

// AdvancedUsageExample demonstrates more advanced usage patterns
// including resume functionality and error handling.
func AdvancedUsageExample() {
	downloader := &Downloader{
		Url: "https://releases.ubuntu.com/20.04/ubuntu-20.04.6-desktop-amd64.iso",
		ID:  "advanced-download-002",

		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "my-custom-filename.zip",
			maxRetries:  5, // More retries for unreliable connections
		},

		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("Starting download: %s\n", d.ServerHeaders.Filename)
				if d.ServerHeaders.Filesize > 0 {
					fmt.Printf("File size: %.2f MB\n", float64(d.ServerHeaders.Filesize)/1024/1024)
				}
				fmt.Printf("Supports resume: %v\n", d.ServerHeaders.AcceptsRanges)
			},

			OnProgress: func(d *Downloader) {
				completed, total, percentage, speed, eta := d.Progress.GetProgressInfo()

				fmt.Printf("\rProgress: %.1f%% | %d/%d bytes | %.2f KB/s | ETA: %v",
					percentage, completed, total, speed/1024, eta.Round(time.Second))
			},

			OnError: func(d *Downloader, err error) {
				log.Printf("Download error: %v", err)

				// Check if it's a network error that might be retryable
				if d.ServerHeaders.AcceptsRanges {
					log.Println("Attempting to resume download...")
					// In a real application, you might want to retry automatically
					time.Sleep(5 * time.Second)
					go d.StartDownload() // Restart download (will resume automatically)
				}
			},

			OnFinish: func(d *Downloader) {
				fmt.Printf("\n✅ Download completed successfully!\n")
				fmt.Printf("📁 File: %s\n", d.fileInfo.FullPath)
				fmt.Printf("📊 Size: %d bytes\n", d.Progress.BytesCompleted)
				fmt.Printf("⏱️  Time: %v\n", d.TimeStats.Elapsed)
				fmt.Printf("🚀 Average Speed: %.2f KB/s\n",
					float64(d.Progress.BytesPerSecond)/1024)
			},
		},
	}

	// Start download
	downloader.StartDownload()
}

// SimpleDownloadExample shows the minimal setup required for a basic download.
func SimpleDownloadExample() {
	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/1048576", // 1MB test file

		// Minimal callbacks
		Callbacks: &Callbacks{
			OnFinish: func(d *Downloader) {
				fmt.Printf("Download completed: %s\n", d.fileInfo.FullPath)
			},
			OnError: func(d *Downloader, err error) {
				fmt.Printf("Download failed: %v\n", err)
			},
		},
	}

	downloader.StartDownload()
}

// runExamples function for testing - uncomment one of the examples to run
func runExamples() {
	fmt.Println("UDM Single Stream Download Examples")
	fmt.Println("===================================")

	// Choose which example to run:

	// SimpleDownloadExample()           // Basic download
	SingleStreamExample() // Full featured example
	// AdvancedUsageExample() // Advanced with resume support
}
