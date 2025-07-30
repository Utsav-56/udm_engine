package main

import (
	"fmt"
	"time"
)

// ProgressBarExample demonstrates the new progress bar functionality
func ProgressBarExample() {
	fmt.Println("🎨 UDM Progress Bar Integration Example")
	fmt.Println("======================================")

	// Create downloader with progress bar enabled
	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/10485760", // 10MB test file
		ID:  "progress-bar-example",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "progress-test.bin",
			threadCount: 4, // Multi-stream to show chunk progress
			maxRetries:  3,
		},
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-ProgressBar-Test/1.0",
			},
		},
		UseProgressBar: true, // Enable progress bar instead of text output
	}

	// Create progress manager
	progressManager := NewProgressManager(downloader)

	// Setup enhanced callbacks that work with progress bar
	SetupProgressCallbacks(downloader, progressManager)

	// Test pause/resume functionality with progress bar
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n🔄 Testing pause (progress bar should show yellow/paused)...")
		downloader.Pause()

		time.Sleep(3 * time.Second)
		fmt.Println("🔄 Testing resume (progress bar should return to green)...")
		downloader.Resume()
	}()

	// Start the download with progress bar
	downloader.StartDownload()
}

// CompareProgressDisplays shows both text and progress bar modes
func CompareProgressDisplays() {
	fmt.Println("🆚 Text Output vs Progress Bar Comparison")
	fmt.Println("=========================================")

	// First, demonstrate text output mode
	fmt.Println("\n1️⃣ Text Output Mode:")
	textDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/5242880", // 5MB
		ID:  "text-output-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "text-output-test.bin",
			threadCount: 1, // Single stream for simplicity
		},
		UseProgressBar: false, // Use text output
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("📥 Started downloading %s\n", d.fileInfo.Name)
			},
			OnProgress: func(d *Downloader) {
				_, _, percentage, speed, eta := d.Progress.GetProgressInfo()
				if int(percentage)%10 == 0 { // Every 10%
					fmt.Printf("📊 Progress: %.1f%% | Speed: %.2f KB/s | ETA: %v\n",
						percentage, speed/1024, eta.Round(time.Second))
				}
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Text output download completed!\n")
			},
		},
	}

	textDownloader.StartDownload()
	time.Sleep(8 * time.Second) // Wait for completion

	// Then, demonstrate progress bar mode
	fmt.Println("\n2️⃣ Progress Bar Mode:")
	progressDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/5242880", // 5MB
		ID:  "progress-bar-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "progress-bar-test.bin",
			threadCount: 4, // Multi-stream to show chunks
		},
		UseProgressBar: true, // Use progress bar
	}

	// Setup progress bar
	progressManager := NewProgressManager(progressDownloader)
	SetupProgressCallbacks(progressDownloader, progressManager)

	progressDownloader.StartDownload()
}

// MultiStreamProgressDemo demonstrates multi-stream download with progress bar
func MultiStreamProgressDemo() {
	fmt.Println("🧵 Multi-Stream Download with Progress Bar")
	fmt.Println("==========================================")

	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/20971520", // 20MB file
		ID:  "multi-stream-progress-demo",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "multi-stream-demo.bin",
			threadCount: 6, // 6 chunks to see chunk progress
			maxRetries:  3,
		},
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-MultiStream-Progress/1.0",
			},
		},
		UseProgressBar: true,
	}

	// Create and setup progress manager
	progressManager := NewProgressManager(downloader)
	SetupProgressCallbacks(downloader, progressManager)

	// Demonstrate pause/resume with visual feedback
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("\n⏸️ Pausing download (watch progress bar turn yellow)...")
		downloader.Pause()

		time.Sleep(4 * time.Second)
		fmt.Println("▶️ Resuming download (watch progress bar turn green)...")
		downloader.Resume()

		time.Sleep(3 * time.Second)
		fmt.Println("⏸️ Pausing again...")
		downloader.Pause()

		time.Sleep(2 * time.Second)
		fmt.Println("▶️ Final resume...")
		downloader.Resume()
	}()

	downloader.StartDownload()
}

// SingleStreamProgressDemo demonstrates single-stream download with progress bar
func SingleStreamProgressDemo() {
	fmt.Println("🔗 Single-Stream Download with Progress Bar")
	fmt.Println("===========================================")

	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/10485760", // 10MB file
		ID:  "single-stream-progress-demo",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "single-stream-demo.bin",
			threadCount: 1, // Force single stream
		},
		UseProgressBar: true,
	}

	progressManager := NewProgressManager(downloader)
	SetupProgressCallbacks(downloader, progressManager)

	// Test control functionality
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n⏸️ Testing pause in single-stream mode...")
		downloader.Pause()

		time.Sleep(2 * time.Second)
		fmt.Println("▶️ Testing resume in single-stream mode...")
		downloader.Resume()
	}()

	downloader.StartDownload()
}

// runProgressExamples is the main function to run progress bar examples
func runProgressExamples() {
	fmt.Println("🎨 UDM Enhanced Progress Bar Examples")
	fmt.Println("====================================")

	// Choose which example to run:

	ProgressBarExample() // Basic progress bar demo
	// CompareProgressDisplays()      // Text vs Progress Bar comparison
	// MultiStreamProgressDemo()      // Multi-stream with chunks
	// SingleStreamProgressDemo()     // Single-stream demo
}
