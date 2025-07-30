package main

import (
	"os"
	"path/filepath"
)

func (d *Downloader) StartDownload() {
	d.Prefetch()
}

func (d *Downloader) Prefetch() {
	headers, _ := GetServerData(d.Url)

	d.ServerHeaders = *headers
	d.CheckPreferences()

	if !headers.AcceptsRanges {
		// d.DownloadSingleStream()
	}

}

func (d *Downloader) CheckPreferences() {

	headers := d.ServerHeaders

	if d.Prefs.fileName == "" && headers.Filename == "" {
		d.fileInfo.Name = "downloaded_file"
	} else if d.Prefs.fileName == "" {
		d.fileInfo.Name = headers.Filename
	} else if headers.Filename == "" {
		d.fileInfo.Name = headers.Filename
	}

	cwd, _ := os.Getwd()
	cwd, _ = filepath.Abs(cwd)
	d.fileInfo.Dir = cwd

	if d.Prefs.DownloadDir != "" {
		d.fileInfo.Dir = d.Prefs.DownloadDir
	}

}
