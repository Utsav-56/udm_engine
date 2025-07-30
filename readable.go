package udm

import "fmt"

func printMap(m map[string]any) {
	for key, value := range m {

		switch v := value.(type) {
		case map[string]any:
			printMap(v) // Recursive call for nested maps
		default:
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

func printList(list []string) {
	for i, item := range list {
		fmt.Printf("%d: %s\n", i+1, item)
	}
}

func ReadableFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	} else if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2f TB", float64(size)/(1024*1024*1024*1024))
	}

}

func ReadableTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		return fmt.Sprintf("%d minutes", minutes)
	} else if seconds < 86400 {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	} else {
		days := seconds / 86400
		hours := (seconds % 86400) / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%d days %d hours %d minutes", days, hours, minutes)
	}
}

func InMBPS(speed float64) string {
	if speed <= 0 {
		return "0.00 MB/s" // Avoid division by zero
	}
	return fmt.Sprintf("%.2f MB/s", speed/(1024*1024)) // Convert bytes to megabytes
}

func ReadablePercentage(percentage float64) string {
	if percentage < 0 {
		return "0.00%"
	} else if percentage > 100 {
		return "100.00%"
	}
	return fmt.Sprintf("%.2f%%", percentage)
}
