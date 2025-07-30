package udm

// Returns a map for progress with all info
// all fields are mandatory and should fill with a valid value
func (d *Downloader) GetProgressMap() map[string]interface{} {

	return map[string]interface{}{
		"id":         d.GetID(),
		"status":     d.GetStatus(),
		"percentage": d.GetProgressPercent(),
		"downloaded": d.GetDownloadedBytes(),
		"filesize":   d.GetFileSize(),
		"speed":      d.GetCurrentSpeed(),
		"eta":        d.GetETA().Seconds(),

		"readable": map[string]interface{}{
			"id":         d.GetID(),
			"status":     d.GetStatus(),
			"percent":    ReadablePercentage(d.GetProgressPercent()),
			"downloaded": ReadableFileSize(d.GetDownloadedBytes()),
			"filesize":   ReadableFileSize(d.GetFileSize()),
			"speed":      InMBPS(d.GetCurrentSpeed()),
			"eta":        ReadableTime(int64(d.GetETA().Seconds())),
		},
	}
}

// Returns a map for finished download with all info
func (d *Downloader) GetFinishedMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         d.GetID(),
		"status":     d.GetStatus(),
		"filename":   d.GetFilename(),
		"output_dir": d.GetOutputDir(),
		"filepath":   d.GetFilePath(),
		"filesize":   d.GetFileSize(),
		"time_taken": int64(d.GetTimeTaken().Seconds()),
		"avg_speed":  d.GetAverageSpeed(),

		"readable": map[string]interface{}{
			"id":         d.GetID(),
			"status":     d.GetStatus(),
			"filename":   d.GetFilename(),
			"output_dir": d.GetOutputDir(),
			"filepath":   d.GetFilePath(),
			"filesize":   ReadableFileSize(d.GetFileSize()),
			"time_taken": ReadableTime(int64(d.GetTimeTaken().Seconds())),
			"avg_speed":  InMBPS(d.GetAverageSpeed()),
		},
	}
}

// Returns a map for downloader configs
func (d *Downloader) GetConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        d.GetID(),
		"outputDir": d.GetOutputDir(),
		"filename":  d.GetFilename(),
		"filesize":  d.GetFileSize(),
		"url":       d.GetURL(),
		"readable": map[string]interface{}{
			"id":        d.GetID(),
			"outputDir": d.GetOutputDir(),
			"filename":  d.GetFilename(),
			"filesize":  ReadableFileSize(d.GetFileSize()),
			"url":       d.GetURL(),
		},
	}
}
