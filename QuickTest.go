package main

import (
	"fmt"
	"time"
)

// QuickMultiStreamTest runs a quick test of the multi-stream functionality
func QuickMultiStreamTest() {
	fmt.Println("🚀 UDM Multi-Stream Download Test")
	fmt.Println("=================================")

	// Create a test downloader for a small file
	downloader := &Downloader{
		Url: "https://httpbin.org/bytes/5242880", // 5MB test file
		ID:  "multi-stream-quick-test",
		Prefs: UserPreferences{
			DownloadDir: "./downloads",
			fileName:    "multi-stream-test.bin",
			threadCount: 4, // Use 4 threads
			maxRetries:  2,
		},
		Headers: CustomHeaders{
			Headers: map[string]string{
				"User-Agent": "UDM-MultiStream-Test/1.0",
			},
		},
		Callbacks: &Callbacks{
			OnStart: func(d *Downloader) {
				fmt.Printf("✅ Multi-stream download started\n")
				fmt.Printf("📁 File: %s\n", d.fileInfo.FullPath)
				fmt.Printf("🧵 Threads: %d\n", d.getThreadCount())
				fmt.Printf("📊 File size: %.2f MB\n", float64(d.ServerHeaders.Filesize)/1024/1024)
				fmt.Printf("📡 Range support: %v\n", d.ServerHeaders.AcceptsRanges)
			},
			OnProgress: func(d *Downloader) {
				_, _, percentage, speed, eta := d.Progress.GetProgressInfo()
				if int(percentage)%20 == 0 { // Print every 20%
					fmt.Printf("📊 Progress: %.0f%% | Speed: %.2f KB/s | ETA: %v\n",
						percentage, speed/1024, eta.Round(time.Second))
				}
			},
			OnChunkStart: func(d *Downloader, chunkIndex int, start, end int64) {
				fmt.Printf("🔗 Chunk %d started: %d-%d bytes (%.2f KB)\n",
					chunkIndex, start, end, float64(end-start+1)/1024)
			},
			OnChunkFinish: func(d *Downloader, chunkIndex int, start, end int64, bytesWritten int64) {
				fmt.Printf("✅ Chunk %d completed: %d bytes written\n", chunkIndex, bytesWritten)
			},
			OnChunkError: func(d *Downloader, chunkIndex int, start, end int64, err error) {
				fmt.Printf("❌ Chunk %d error: %v\n", chunkIndex, err)
			},
			OnAssembleStart: func(d *Downloader) {
				fmt.Println("🔧 Merging chunks into final file...")
			},
			OnAssembleFinish: func(d *Downloader) {
				fmt.Println("✅ Chunks merged successfully!")
			},
			OnAssembleError: func(d *Downloader, err error) {
				fmt.Printf("❌ Assembly failed: %v\n", err)
			},
			OnFinish: func(d *Downloader) {
				fmt.Printf("🎉 Multi-stream download completed!\n")
				fmt.Printf("⏱️  Duration: %v\n", d.TimeStats.Elapsed)
				fmt.Printf("🚀 Average speed: %.2f KB/s\n", float64(d.Progress.BytesPerSecond)/1024)
			},
			OnError: func(d *Downloader, err error) {
				fmt.Printf("❌ Download failed: %v\n", err)
			},
			OnPause: func(d *Downloader) {
				fmt.Println("⏸️  Download paused")
			},
			OnResume: func(d *Downloader) {
				fmt.Println("▶️  Download resumed")
			},
			OnStop: func(d *Downloader) {
				fmt.Println("🛑 Download cancelled")
			},
		},
	}

	// Test pause/resume after a short delay
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("\n🔄 Testing pause...")
		downloader.Pause()

		time.Sleep(2 * time.Second)
		fmt.Println("🔄 Testing resume...")
		downloader.Resume()
	}()

	// Start the download
	downloader.StartDownload()
}

// To run this test, use: go run QuickTest.go DownloaderModels.go DownloadSingleStream.go DownloadMultiStream.go ServerHeaders.go StartDownload.go DivideChunks.go ufs/
