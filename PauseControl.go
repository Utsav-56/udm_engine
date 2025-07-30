package udm

import "sync"

// PauseController is used to manage the pause and resume functionality
// It uses a mutex and condition variable to handle pausing and resuming
type PauseController struct {
	mu       sync.Mutex
	cond     *sync.Cond
	isPaused bool
}

// NewPauseController creates a new PauseController instance.
//
// Returns:
//   - *PauseController: Initialized pause controller
func NewPauseController() *PauseController {
	pc := &PauseController{
		isPaused: false,
	}
	pc.cond = sync.NewCond(&pc.mu)
	return pc
}

// Pause sets the controller to paused state.
func (pc *PauseController) Pause() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.isPaused = true
}

// Resume sets the controller to resumed state and wakes up waiting goroutines.
func (pc *PauseController) Resume() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.isPaused = false
	pc.cond.Broadcast()
}

// IsPaused returns the current pause state.
//
// Returns:
//   - bool: True if paused, false if running
func (pc *PauseController) IsPaused() bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.isPaused
}

// WaitIfPaused blocks the calling goroutine while the controller is paused.
func (pc *PauseController) WaitIfPaused() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for pc.isPaused {
		pc.cond.Wait()
	}
}
