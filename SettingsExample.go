package main

import (
	"fmt"
	"time"
)

// SettingsIntegrationExample demonstrates the new settings system
func ShowSettingsExample() {
	fmt.Println("⚙️ UDM Settings Integration Example")
	fmt.Println("==================================")
	fmt.Println()

	// Initialize settings from config file
	if err := InitializeSettings(); err != nil {
		fmt.Printf("❌ Failed to load settings: %v\n", err)
		return
	}

	fmt.Println("✅ Settings loaded successfully!")
	fmt.Printf("📊 Thread count from config: %d\n", UDMSettings.GetThreadCount())
	fmt.Printf("📁 Main output directory: %s\n", UDMSettings.MainOutputDir)
	fmt.Printf("📏 Minimum file size for multi-stream: %d bytes\n", UDMSettings.MinimumFileSize)
	fmt.Println()

	// Validate settings
	warnings := UDMSettings.ValidateSettings()
	if len(warnings) > 0 {
		fmt.Println("⚠️ Settings warnings:")
		for _, warning := range warnings {
			fmt.Printf("  • %s\n", warning)
		}
		fmt.Println()
	}

	// Create missing directories
	if err := UDMSettings.CreateMissingDirectories(); err != nil {
		fmt.Printf("❌ Failed to create directories: %v\n", err)
	} else {
		fmt.Println("📁 Output directories verified/created")
	}

	fmt.Println()
	fmt.Println("🔍 Testing extension-based directory mapping:")

	// Test different file extensions
	testFiles := []string{
		"document.pdf",
		"video.mp4",
		"archive.zip",
		"music.mp3",
		"image.jpg",
		"installer.exe",
		"code.py",
		"font.ttf",
		"unknown.xyz",
	}

	for _, filename := range testFiles {
		outputDir := UDMSettings.GetOutputDirForFile(filename)
		category := UDMSettings.GetCategoryForExtension(filename)
		fmt.Printf("  📄 %s → %s (category: %s)\n", filename, outputDir, category)
	}

	fmt.Println()
	fmt.Println("🧵 Testing file size-based stream selection:")

	// Test different file sizes
	testSizes := []int64{
		512 * 1024,        // 512KB
		2 * 1024 * 1024,   // 2MB
		15 * 1024 * 1024,  // 15MB
		100 * 1024 * 1024, // 100MB
	}

	for _, size := range testSizes {
		shouldUseSingle := UDMSettings.ShouldUseSingleStream(size)
		sizeStr := formatBytes(size)
		if shouldUseSingle {
			fmt.Printf("  📊 %s → Single-stream (below minimum threshold)\n", sizeStr)
		} else {
			fmt.Printf("  📊 %s → Multi-stream (above minimum threshold)\n", sizeStr)
		}
	}

	fmt.Println()
	fmt.Println("🚀 Testing actual download with settings:")

	// Example 1: Small file (should use single-stream based on config)
	fmt.Println("\n1️⃣ Small file download (PDF document):")
	smallFileDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/2097152", // 2MB
		ID:  "small-pdf-test",
		Prefs: UserPreferences{
			fileName: "test-document.pdf", // PDF will go to documents folder
			// No downloadDir specified - will use config mapping
			// No threadCount specified - will use config
		},
		UseProgressBar: true,
	}

	// Setup progress bar
	progressManager := NewProgressManager(smallFileDownloader)
	SetupProgressCallbacks(smallFileDownloader, progressManager)

	fmt.Printf("📂 Expected output dir: %s\n", UDMSettings.GetOutputDirForFile("test-document.pdf"))
	fmt.Printf("🧵 Expected thread count: %d\n", UDMSettings.GetThreadCount())

	// Start download
	smallFileDownloader.StartDownload()
	time.Sleep(10 * time.Second) // Wait a bit for demo

	fmt.Println("\n2️⃣ Large file download (Video file):")
	largeFileDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/20971520", // 20MB
		ID:  "large-video-test",
		Prefs: UserPreferences{
			fileName: "test-video.mp4", // MP4 will go to videos folder
			// No downloadDir specified - will use config mapping
			// No threadCount specified - will use config
		},
		UseProgressBar: true,
	}

	// Setup progress bar
	progressManager2 := NewProgressManager(largeFileDownloader)
	SetupProgressCallbacks(largeFileDownloader, progressManager2)

	fmt.Printf("📂 Expected output dir: %s\n", UDMSettings.GetOutputDirForFile("test-video.mp4"))
	fmt.Printf("🧵 Expected thread count: %d\n", UDMSettings.GetThreadCount())

	// Start download
	largeFileDownloader.StartDownload()
	time.Sleep(15 * time.Second) // Wait for demo

	fmt.Println("\n🎯 User Preference Override Example:")
	fmt.Println("===================================")

	// Example with user overriding config settings
	userOverrideDownloader := &Downloader{
		Url: "https://httpbin.org/bytes/5242880", // 5MB
		ID:  "user-override-test",
		Prefs: UserPreferences{
			fileName:    "custom-video.mp4",
			threadCount: 12,             // Override config thread count
			DownloadDir: "./custom-dir", // Override config output directory
		},
		UseProgressBar: true,
	}

	// Setup progress bar
	progressManager3 := NewProgressManager(userOverrideDownloader)
	SetupProgressCallbacks(userOverrideDownloader, progressManager3)

	fmt.Printf("🧵 Config thread count: %d\n", UDMSettings.GetThreadCount())
	fmt.Printf("🧵 User override thread count: %d\n", userOverrideDownloader.getThreadCount())
	fmt.Printf("📁 Config output dir for mp4: %s\n", UDMSettings.GetOutputDirForFile("test.mp4"))
	fmt.Printf("� User override output dir: %s\n", userOverrideDownloader.Prefs.DownloadDir)

	// Start download
	userOverrideDownloader.StartDownload()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}
