package udm

import (
	"os"
	"path/filepath"
	"strings"
	"udm/ufs"
	"udm/ujson"
)

type CategoryInfo struct {
	Name      string   `json:"name"`
	Exts      []string `json:"exts"`
	OutputDir string   `json:"outputDir"`
}

type Settings struct {
	ThreadCount            int               `json:"ThreadCount"`
	MaxRetries             int               `json:"MaxRetries"`
	MinimumFileSize        int64             `json:"MinimumFileSize"`
	MaxConcurrentDownloads int               `json:"MaxConcurrentDownloads"`
	Categories             []string          `json:"Categories"`
	Extensions             []string          `json:"Extensions"`
	OutputDir              string            `json:"OutputDir"`
	MainOutputDir          string            `json:"MainOutputDir"`
	CategoryInfo           []CategoryInfo    `json:"categoryInfo"`
	CustomHeaders          map[string]string `json:"CustomHeaders"`
	CustomCookies          string            `json:"CustomCookies"`
}

// UDMSettings holds the global settings instance
var UDMSettings *Settings

// LoadSettings loads settings from the JSON configuration file
func LoadSettings(configPath string) (*Settings, error) {
	// Use default path if not provided
	if configPath == "" {
		configPath = "udmConfigs.json"
	}

	// Read and unmarshal the JSON file directly
	var settings Settings
	if err := ujson.UnmarshalJSONFile(configPath, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// InitializeSettings loads and initializes the global settings
func InitializeSettings() error {
	settings, err := LoadSettings("udmConfigs.json")
	if err != nil {
		return err
	}

	UDMSettings = settings
	return nil
}

// GetThreadCount returns the thread count from config with fallback
func (s *Settings) GetThreadCount() int {
	if s.ThreadCount > 0 {
		return s.ThreadCount
	}
	return 8 // Default fallback
}

// ShouldUseSingleStream determines if single stream should be used based on file size
func (s *Settings) ShouldUseSingleStream(fileSize int64) bool {
	if s.MinimumFileSize <= 0 {
		// Default minimum size for multi-stream (10MB)
		return fileSize < 10*1024*1024
	}
	return fileSize < s.MinimumFileSize
}

// GetOutputDirForFile determines the output directory based on file extension
func (s *Settings) GetOutputDirForFile(filename string) string {
	// Extract file extension
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		return s.getDefaultOutputDir()
	}

	// Look for extension in category info
	for _, category := range s.CategoryInfo {
		for _, categoryExt := range category.Exts {
			if strings.ToLower(categoryExt) == ext {
				if category.OutputDir != "" {
					return category.OutputDir
				}
			}
		}
	}

	// Use MainOutputDir if available
	if s.MainOutputDir != "" {
		return s.MainOutputDir
	}

	// Use OutputDir as fallback
	if s.OutputDir != "" {
		return s.OutputDir
	}

	// Use system default downloads directory
	return s.getDefaultOutputDir()
}

// getDefaultOutputDir returns the system default downloads directory
func (s *Settings) getDefaultOutputDir() string {
	// Try to get user's Downloads folder
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "." // Ultimate fallback
		}
		return cwd
	}

	return filepath.Join(userHomeDir, "Downloads")
}

// GetCustomHeaders returns custom headers if configured
func (s *Settings) GetCustomHeaders() map[string]string {
	if s.CustomHeaders != nil && len(s.CustomHeaders) > 0 {
		return s.CustomHeaders
	}
	return nil
}

// GetCustomCookies returns custom cookies if configured
func (s *Settings) GetCustomCookies() string {
	return s.CustomCookies
}

// GetMaxRetries returns the maximum retry count with fallback
func (s *Settings) GetMaxRetries() int {
	if s.MaxRetries > 0 {
		return s.MaxRetries
	}
	return 3 // Default fallback
}

// ApplySettingsToDownloader applies settings to a downloader instance
func (s *Settings) ApplySettingsToDownloader(d *Downloader) {
	// Apply thread count (always from config)
	if d.Prefs.threadCount <= 0 {
		d.Prefs.threadCount = s.GetThreadCount()
	}

	// Apply max retries if not set
	if d.Prefs.maxRetries <= 0 {
		d.Prefs.maxRetries = s.GetMaxRetries()
	}

	// Apply output directory if user hasn't specified one
	if d.Prefs.DownloadDir == "" {
		// Use filename to determine appropriate directory
		if d.fileInfo.Name != "" {
			d.Prefs.DownloadDir = s.GetOutputDirForFile(d.fileInfo.Name)
		} else {
			// Use default output directory
			d.Prefs.DownloadDir = s.getDefaultOutputDir()
		}
	}

	// Apply custom headers if not already set and available in config
	configHeaders := s.GetCustomHeaders()
	if configHeaders != nil && len(configHeaders) > 0 {
		if d.Headers.Headers == nil {
			d.Headers.Headers = make(map[string]string)
		}

		// Add config headers (user headers take priority)
		for key, value := range configHeaders {
			if _, exists := d.Headers.Headers[key]; !exists {
				d.Headers.Headers[key] = value
			}
		}
	}

	// Apply custom cookies if not already set and available in config
	configCookies := s.GetCustomCookies()
	if configCookies != "" && d.Headers.Cookies == "" {
		d.Headers.Cookies = configCookies
	}
}

// GetCategoryForExtension returns the category name for a given file extension
func (s *Settings) GetCategoryForExtension(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		return "unknown"
	}

	for _, category := range s.CategoryInfo {
		for _, categoryExt := range category.Exts {
			if strings.ToLower(categoryExt) == ext {
				return category.Name
			}
		}
	}

	return "unknown"
}

// ValidateSettings performs basic validation of the settings
func (s *Settings) ValidateSettings() []string {
	var warnings []string

	if s.ThreadCount <= 0 {
		warnings = append(warnings, "ThreadCount should be greater than 0, using default (8)")
	}

	if s.MinimumFileSize <= 0 {
		warnings = append(warnings, "MinimumFileSize should be greater than 0, using default (10MB)")
	}

	if s.MainOutputDir != "" {
		if _, err := os.Stat(s.MainOutputDir); os.IsNotExist(err) {
			warnings = append(warnings, "MainOutputDir does not exist: "+s.MainOutputDir)
		}
	}

	// Validate category output directories
	for _, category := range s.CategoryInfo {
		if category.OutputDir != "" {
			if _, err := os.Stat(category.OutputDir); os.IsNotExist(err) {
				warnings = append(warnings, "Category output directory does not exist: "+category.OutputDir)
			}
		}
	}

	return warnings
}

// CreateMissingDirectories creates any missing output directories
func (s *Settings) CreateMissingDirectories() error {
	// Create main output directory
	if s.MainOutputDir != "" {
		if err := os.MkdirAll(s.MainOutputDir, 0755); err != nil {
			return err
		}
	}

	// Create category directories
	for _, category := range s.CategoryInfo {
		if category.OutputDir != "" {
			if err := os.MkdirAll(category.OutputDir, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

// Deprecated: Use LoadSettings instead
func GenerateSettingsFromJsonFile(jsonPath string) Settings {
	settings, err := LoadSettings(jsonPath)
	if err != nil {
		panic(err)
	}
	return *settings
}

func (s *Settings) ShouldCapture(filename string) bool {
	extension := ufs.FileExtension(filename)
	extension = strings.ToLower(extension)

	// remove the dot
	if extension[0] == '.' {
		extension = extension[1:]
	}

	for _, ext := range s.Extensions {
		if ext == extension {
			return true
		}
	}
	return false
}

func ShouldCapture(filename string) bool {
	settings, err := LoadSettings("udmConfigs.json")
	if err != nil {
		return false
	}
	return settings.ShouldCapture(filename)
}

func GetSettings() *Settings {
	settings, err := LoadSettings("udmConfigs.json")
	if err != nil {
		panic(err)
	}
	return settings
}

func GetOutputDirForFile(filename string) string {
	settings, err := LoadSettings("udmConfigs.json")
	if err != nil {
		panic(err)
	}
	return settings.GetOutputDirForFile(filename)
}
