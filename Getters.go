package udm

// Returns a map as below
/*
	filename: string
	outputDir: string
	extension: string
url: string
*/
func GetConfigMap(d *Downloader) map[string]string {
	filename := d.fileInfo.Name
	outputDir := d.fileInfo.Dir
	extension := d.fileInfo.FullPath
	url := d.Url
	return map[string]string{
		"filename":  filename,
		"outputDir": outputDir,
		"extension": extension,
		"url":       url,
	}
}

func GetProgressMap(d *Downloader) map[string]int64 {
	return map[string]int64{
		"status": d.Status,
		"total":  d.fileInfo.Size,
	}
}
