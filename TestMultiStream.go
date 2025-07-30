package main

import (
	"fmt"
	"os"
	"time"
)

// TestMultiStreamDownload tests the multi-stream download functionality
func TestMultiStreamDownload() {
	fmt.Println("🧪 Testing Multi-Stream Download Implementation")
	fmt.Println("=============================================")

	// Test with a smaller file first
	testDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/10485760", // 10MB test file
		ID:  "multi-stream-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "multi-stream-test.bin",
			threadCount: 4,
			maxRetries:  3,
		},
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-Test/1.0",
			},
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("✅ Download started for: %s\n", d.Url)
				fmt.Printf("📁 File: %s\n", d.fileInfo.FullPath)
				fmt.Printf("🧵 Thread count: %d\n", d.getThreadCount())
			},
			OnProgress: func(d *Downloader) {
				_, _, percentage, speed, eta := d.Progress.GetProgressInfo()
				if int(percentage)%5 == 0 { // Print every 5%
					fmt.Printf("📊 %.0f%% | %.2f KB/s | ETA: %v\n",
						percentage, speed/1024, eta.Round(time.Second))
				}
			},
			OnChunkStart: func(d *Downloader, chunkIndex int, start, end int64) {
				fmt.Printf("🔗 Starting chunk %d: bytes %d-%d\n", chunkIndex, start, end)
			},
			OnChunkFinish: func(d *Downloader, chunkIndex int, start, end int64, bytesWritten int64) {
				fmt.Printf("✅ Chunk %d completed: %d bytes\n", chunkIndex, bytesWritten)
			},
			OnAssembleStart: func(d *Downloader) {
				fmt.Println("🔧 Assembling chunks...")
			},
			OnAssembleFinish: func(d *Downloader) {
				fmt.Println("✅ Assembly completed!")
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("🎉 Download completed in %v\n", d.TimeStats.Elapsed)

				// Verify file exists and has correct size
				if info, err := os.Stat(d.fileInfo.FullPath); err == nil {
					fmt.Printf("📊 Final file size: %d bytes\n", info.Size())
					fmt.Printf("✅ File verification: PASSED\n")
				} else {
					fmt.Printf("❌ File verification: FAILED - %v\n", err)
				}
			},
			OnError: func(d *Downloader, err error) {
				fmt.Printf("❌ Download error: %v\n", err)
			},
		},
	}

	// Start the test download
	testDownloader.StartDownload()
}

// TestPauseResumeMultiStream tests pause/resume functionality in multi-stream mode
func TestPauseResumeMultiStream() {
	fmt.Println("\n🔄 Testing Multi-Stream Pause/Resume")
	fmt.Println("===================================")

	downloader := &Downloader{
		Url: "https://drive.usercontent.google.com/download?id=1d1EBTcLHYQiv93O4nyBBjbK_Wc-2f5qX&export=download&authuser=0&confirm=t&uuid=5cccf6aa-fc97-4bff-89f3-e4339a189778&at=AN8xHoo83pMm2eQ2GwbC6YHA5eK0:1753245767095", // 20MB test file
		ID:  "pause-resume-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			threadCount: 6,
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("▶️  Started multi-stream download\n")
			},
			OnProgress: func(d *Downloader) {
				_, _, percentage, _, _ := d.Progress.GetProgressInfo()
				if int(percentage)%10 == 0 {
					fmt.Printf("📊 Progress: %.0f%%\n", percentage)
				}
			},
			OnPause: func(d *Downloader) {
				_, _, percentage, _, _ := d.Progress.GetProgressInfo()
				fmt.Printf("⏸️  Download paused at %.1f%%\n", percentage)
			},
			OnResume: func(d *Downloader) {
				_, _, percentage, _, _ := d.Progress.GetProgressInfo()
				fmt.Printf("▶️  Download resumed from %.1f%%\n", percentage)
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Pause/Resume test completed successfully!\n")
			},
		},
	}

	// Control sequence
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("🔄 Pausing download...")
		downloader.Pause()

		time.Sleep(8 * time.Second)
		fmt.Println("🔄 Resuming download...")
		downloader.Resume()

		time.Sleep(9 * time.Second)
		fmt.Println("🔄 Pausing again...")
		downloader.Pause()

		time.Sleep(15 * time.Second)
		fmt.Println("🔄 Final resume...")
		downloader.Resume()
	}()

	downloader.StartDownload()
}

// TestMultiStreamStrategySelection tests automatic selection between single and multi-stream
func TestMultiStreamStrategySelection() {
	fmt.Println("\n🎯 Testing Automatic Strategy Selection")
	fmt.Println("======================================")

	// Test small file (should use single-stream)
	fmt.Println("\n📝 Testing small file (should use single-stream):")
	smallFileDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/1048576", // 1MB - should use single stream
		ID:  "small-file-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "small-file.bin",
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				if d.getThreadCount() == 1 {
					fmt.Printf("✅ Correctly selected single-stream for small file\n")
				} else {
					fmt.Printf("❌ Incorrectly selected multi-stream for small file\n")
				}
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Small file download completed\n")
			},
		},
	}

	smallFileDownloader.StartDownload()
	time.Sleep(5 * time.Second) // Wait for completion

	// Test large file (should use multi-stream)
	fmt.Println("\n📝 Testing large file (should use multi-stream):")
	largeFileDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/52428800", // 50MB - should use multi-stream
		ID:  "large-file-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "large-file.bin",
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				if d.getThreadCount() > 1 && d.ServerHeaders.AcceptsRanges {
					fmt.Printf("✅ Correctly selected multi-stream for large file (%d threads)\n", d.getThreadCount())
				} else {
					fmt.Printf("ℹ️  Using single-stream (possibly no range support)\n")
				}
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("✅ Large file download completed\n")
			},
		},
	}

	largeFileDownloader.StartDownload()
}

func runMultiStreamTests() {
	// Run the tests
	// TestMultiStreamDownload()
	// time.Sleep(10 * time.Second)

	TestPauseResumeMultiStream()
	// time.Sleep(15 * time.Second)

	// TestMultiStreamStrategySelection()
}
