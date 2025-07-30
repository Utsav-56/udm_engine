package main

import (
	"fmt"
)

// Simple test function to verify settings system
func main() {
	fmt.Println("🧪 Testing UDM Settings System")
	fmt.Println("==============================")

	// Test 1: Initialize settings
	fmt.Println("📂 Initializing settings from udmConfigs.json...")
	if err := InitializeSettings(); err != nil {
		fmt.Printf("❌ Failed to load settings: %v\n", err)
		fmt.Println("ℹ️  Make sure udmConfigs.json exists in the current directory")
		return
	}

	fmt.Println("✅ Settings loaded successfully!")

	// Test 2: Display basic settings
	fmt.Printf("📊 Thread count: %d\n", UDMSettings.GetThreadCount())
	fmt.Printf("📁 Main output directory: %s\n", UDMSettings.MainOutputDir)
	fmt.Printf("📏 Minimum file size for multi-stream: %d bytes\n", UDMSettings.MinimumFileSize)
	fmt.Printf("🔄 Max retries: %d\n", UDMSettings.GetMaxRetries())

	// Test 3: Directory mapping
	fmt.Println("\n🗂️  Testing extension-based directory mapping:")
	testFiles := []string{
		"document.pdf",
		"video.mp4",
		"archive.zip",
		"music.mp3",
		"image.jpg",
		"installer.exe",
		"unknown.xyz",
	}

	for _, filename := range testFiles {
		outputDir := UDMSettings.GetOutputDirForFile(filename)
		category := UDMSettings.GetCategoryForExtension(filename)
		fmt.Printf("  📄 %s → %s (category: %s)\n", filename, outputDir, category)
	}

	// Test 4: File size stream selection
	fmt.Println("\n🧵 Testing file size-based stream selection:")
	testSizes := []int64{
		512 * 1024,        // 512KB
		2 * 1024 * 1024,   // 2MB
		15 * 1024 * 1024,  // 15MB
		100 * 1024 * 1024, // 100MB
	}

	for _, size := range testSizes {
		shouldUseSingle := UDMSettings.ShouldUseSingleStream(size)
		sizeStr := formatFileSize(size)
		streamType := "Multi-stream"
		if shouldUseSingle {
			streamType = "Single-stream"
		}
		fmt.Printf("  📊 %s → %s\n", sizeStr, streamType)
	}

	// Test 5: Validate settings
	fmt.Println("\n🔍 Validating settings...")
	warnings := UDMSettings.ValidateSettings()
	if len(warnings) > 0 {
		fmt.Println("⚠️  Settings warnings:")
		for _, warning := range warnings {
			fmt.Printf("  • %s\n", warning)
		}
	} else {
		fmt.Println("✅ No issues found with settings")
	}

	// Test 6: Create directories (if needed)
	fmt.Println("\n📁 Creating output directories...")
	if err := UDMSettings.CreateMissingDirectories(); err != nil {
		fmt.Printf("❌ Failed to create directories: %v\n", err)
	} else {
		fmt.Println("✅ Output directories verified/created")
	}

	// Test 7: Custom headers and cookies
	fmt.Println("\n🔧 Custom headers and cookies:")
	customHeaders := UDMSettings.GetCustomHeaders()
	customCookies := UDMSettings.GetCustomCookies()

	if customHeaders != nil && len(customHeaders) > 0 {
		fmt.Println("📋 Custom headers configured:")
		for key, value := range customHeaders {
			fmt.Printf("  %s: %s\n", key, value)
		}
	} else {
		fmt.Println("📋 No custom headers configured")
	}

	if customCookies != "" {
		fmt.Printf("🍪 Custom cookies: %s\n", customCookies)
	} else {
		fmt.Println("🍪 No custom cookies configured")
	}

	fmt.Println("\n✅ Settings system test completed successfully!")
	fmt.Println("🚀 The UDM engine is ready to use with configuration-based downloads!")
}

// formatFileSize formats bytes in human readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
