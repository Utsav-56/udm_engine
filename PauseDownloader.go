package udm

// Pause pauses the current download operation.
func (d *Downloader) Pause() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	if !d.PauseControl.isPaused {
		d.PauseControl.isPaused = true
		d.Status = DOWNLOAD_PAUSED
	}
}

// Resume resumes a paused download operation.
func (d *Downloader) Resume() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	if d.PauseControl.isPaused {
		d.PauseControl.isPaused = false
		d.Status = DOWNLOAD_IN_PROGRESS
		d.PauseControl.cond.Broadcast()
	}
}

// Cancel cancels the current download operation.
func (d *Downloader) Cancel() {
	d.PauseControl.mu.Lock()
	defer d.PauseControl.mu.Unlock()

	d.PauseControl.isPaused = false
	d.Status = DOWNLOAD_STOPPED
	d.PauseControl.cond.Broadcast()
}
