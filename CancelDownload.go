package udm

func (d *Downloader) StopDownload() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isStopped {
		return
	}

	// Cancel the context to stop all goroutines
	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	// Close any open files
	//if d.file != nil {
	//	_ = d.file.Close()
	//	d.file = nil
	//}

	// Clear callbacks to prevent memory leaks
	d.Callbacks = nil

	// Mark as stopped
	d.isStopped = true
}

func (d *Downloader) Dispose() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.StopDownload()

}

func (d *Downloader) ClearCallbacks() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Callbacks = nil
}
