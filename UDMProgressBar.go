package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UDMProgressTracker represents the progress data for UDM downloads
type UDMProgressTracker struct {
	Filename       string
	BytesCompleted int64
	TotalBytes     int64
	SpeedBps       float64
	Percentage     float64
	ETA            time.Duration
	StartTime      time.Time
	IsPaused       bool
	IsCompleted    bool
	OutputDir      string

	// Multi-stream specific
	IsMultiStream bool
	ChunkProgress []ChunkProgress // Progress for each chunk
}

// ChunkProgress represents progress for individual chunks in multi-stream downloads
type ChunkProgress struct {
	Index      int
	Percentage float64
	IsComplete bool
}

// UDMProgressModel represents the Bubble Tea model for UDM progress display
type UDMProgressModel struct {
	tracker     *UDMProgressTracker
	progressBar progress.Model
	width       int
	height      int
}

type progressTickMsg time.Time
type progressUpdateMsg UDMProgressTracker
type progressCompletionMsg struct{}

// NewUDMProgress creates a new UDM progress bar
func NewUDMProgress(tracker *UDMProgressTracker) *UDMProgressModel {
	p := progress.New(progress.WithGradient("#00d7af", "#5fafff"))
	p.Width = 50

	return &UDMProgressModel{
		tracker:     tracker,
		progressBar: p,
		width:       80,
		height:      20,
	}
}

// Init initializes the Bubble Tea model
func (m UDMProgressModel) Init() tea.Cmd {
	return progressTick()
}

func progressTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return progressTickMsg(t)
	})
}

// Update handles Bubble Tea messages
func (m UDMProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressTickMsg:
		if m.tracker.IsCompleted {
			return m, tea.Quit
		}
		return m, progressTick()

	case progressUpdateMsg:
		// Update tracker with new data
		*m.tracker = UDMProgressTracker(msg)
		if m.tracker.IsCompleted {
			return m, tea.Sequence(
				func() tea.Msg { return progressCompletionMsg{} },
				tea.Quit,
			)
		}
		return m, nil

	case progressCompletionMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = m.width - 20
		return m, nil
	}

	return m, nil
}

// View renders the progress bar
func (m UDMProgressModel) View() string {
	if m.tracker.IsCompleted {
		return m.renderCompletionView()
	}

	return m.renderProgressView()
}

// renderProgressView renders the active download progress
func (m UDMProgressModel) renderProgressView() string {
	// Style definitions
	filenameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00d7af")).Bold(true)
	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
	speedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5fafff")).Bold(true)
	etaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaf00")).Bold(true)
	chunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))

	// Header line with filename and size
	headerLine := fmt.Sprintf("filename :: %s            Size:: %s",
		filenameStyle.Render(m.tracker.Filename),
		sizeStyle.Render(formatProgressBytes(m.tracker.TotalBytes)),
	)

	// Progress bar with percentage
	progressPercent := m.tracker.Percentage / 100.0
	var progressBar string

	if m.tracker.IsPaused {
		// Yellow progress bar for paused state
		pausedBar := progress.New(progress.WithGradient("#ffff00", "#ffa500"))
		pausedBar.Width = m.progressBar.Width
		progressBar = pausedBar.ViewAs(progressPercent)

		// Add PAUSED text in the middle
		barLength := m.progressBar.Width
		pausedText := "PAUSED"
		padding := (barLength - len(pausedText)) / 2
		if padding > 0 {
			progressBar = progressBar[:padding] + pausedText + progressBar[padding+len(pausedText):]
		}
	} else {
		// Green progress bar for active state
		progressBar = m.progressBar.ViewAs(progressPercent)
	}

	progressLine := fmt.Sprintf("%s %.1f%%", progressBar, m.tracker.Percentage)

	// Details line
	detailsLine := fmt.Sprintf("completed : %s / %s      Speed :: %s   ETA:: %s",
		formatProgressBytes(m.tracker.BytesCompleted),
		formatProgressBytes(m.tracker.TotalBytes),
		speedStyle.Render(formatProgressSpeed(m.tracker.SpeedBps)),
		etaStyle.Render(formatProgressDuration(m.tracker.ETA)),
	)

	// Build the view
	var view strings.Builder
	view.WriteString(headerLine + "\n")
	view.WriteString(progressLine + "\n")
	view.WriteString(detailsLine + "\n")

	// Add chunk progress for multi-stream downloads
	if m.tracker.IsMultiStream && len(m.tracker.ChunkProgress) > 0 {
		view.WriteString("\n")

		// Group chunks in rows (show 4 chunks per row)
		chunksPerRow := 4
		for i := 0; i < len(m.tracker.ChunkProgress); i += chunksPerRow {
			var chunkLine strings.Builder

			for j := 0; j < chunksPerRow && i+j < len(m.tracker.ChunkProgress); j++ {
				chunk := m.tracker.ChunkProgress[i+j]
				chunkText := fmt.Sprintf("chunk %d:: %.1f%%", chunk.Index+1, chunk.Percentage)

				if chunk.IsComplete {
					chunkText = filenameStyle.Render(chunkText) // Green for completed
				} else {
					chunkText = chunkStyle.Render(chunkText) // Gray for in progress
				}

				chunkLine.WriteString(fmt.Sprintf("%-20s", chunkText))

				if j < chunksPerRow-1 && i+j+1 < len(m.tracker.ChunkProgress) {
					chunkLine.WriteString("     ")
				}
			}

			view.WriteString(chunkLine.String() + "\n")
		}
	}

	return view.String()
}

// renderCompletionView renders the final completion message
func (m UDMProgressModel) renderCompletionView() string {
	// Style definitions
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00d7af")).Bold(true)
	filenameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00d7af")).Bold(true)
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaf00")).Bold(true)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5fafff")).Bold(true)
	speedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5fff")).Bold(true)

	elapsed := time.Since(m.tracker.StartTime)
	avgSpeed := float64(m.tracker.TotalBytes) / elapsed.Seconds()

	border := strings.Repeat("=", 50)

	completion := fmt.Sprintf(`%s
%s
%s
Filename :: %s
Output dir :: %s
Time taken :: %s
Average speed :: %s
%s`,
		border,
		successStyle.Render("File downloaded Successfully::"),
		border,
		filenameStyle.Render(m.tracker.Filename),
		dirStyle.Render(m.tracker.OutputDir),
		timeStyle.Render(formatProgressDuration(elapsed)),
		speedStyle.Render(formatProgressSpeed(avgSpeed)),
		border,
	)

	return completion
}

// formatProgressBytes formats bytes into human readable format
func formatProgressBytes(bytes int64) string {
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

// formatProgressSpeed formats speed into human readable format
func formatProgressSpeed(speedBps float64) string {
	speedMBps := speedBps / (1024 * 1024)
	return fmt.Sprintf("%.2f MB/s", speedMBps)
}

// formatProgressDuration formats duration into human readable format
func formatProgressDuration(d time.Duration) string {
	if d < 0 {
		return "∞"
	}

	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
