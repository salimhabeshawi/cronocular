package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/salimhabeshawi/cronocular/internal/timer"
)

type tickMsg time.Time

type model struct {
	controller *timer.Controller
	progress   progress.Model
	width      int
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	hotkeyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	restStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	focusStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
)

var rainbow = []lipgloss.Color{
	lipgloss.Color("196"),
	lipgloss.Color("202"),
	lipgloss.Color("226"),
	lipgloss.Color("46"),
	lipgloss.Color("45"),
	lipgloss.Color("21"),
	lipgloss.Color("201"),
}

const wordmark = `                                 _            
  ___ _ __ ___  _ __   ___   ___ _   _| | __ _ _ __ 
 / __| '__/ _ \| '_ \ / _ \ / __| | | | |/ _' | '__|
| (__| | | (_) | | | | (_) | (__| |_| | | (_| | |   
 \___|_|  \___/|_| |_|\___/ \___|\__,_|_|\__,_|_|`

func Run(ctx context.Context, controller *timer.Controller) error {
	p := tea.NewProgram(newModel(controller), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

func newModel(controller *timer.Controller) model {
	bar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	bar.Width = 40
	return model{controller: controller, progress: bar, width: 72}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = clamp(msg.Width-8, 20, 64)
		return m, nil
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.controller.TogglePause()
		}
	case tickMsg:
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	s := m.controller.Snapshot()
	total := s.Focus
	phaseLabel := "Focus"
	phaseStyle := focusStyle
	message := "Next eye break starts in"

	if s.Phase == timer.PhaseRest {
		total = s.Rest
		phaseLabel = "Rest"
		phaseStyle = restStyle
		message = "Look 20 feet away for"
	}

	if s.Paused {
		phaseLabel += " paused"
		message = "Timer paused with"
	}

	remaining := roundUpSecond(s.Remaining)
	elapsed := total - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	percent := 0.0
	if total > 0 {
		percent = float64(elapsed) / float64(total)
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(rainbowWordmark(s.UpdatedAt))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("20-20-20 eye care timer"))
	b.WriteString("  ")
	b.WriteString(mutedStyle.Render("by https://github.com/salimhabeshawi"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s  %s\n\n", phaseStyle.Render(phaseLabel), mutedStyle.Render(fmt.Sprintf("cycle %d", s.Cycle+1))))
	b.WriteString(fmt.Sprintf("%s %s\n\n", message, titleStyle.Render(formatDuration(remaining))))
	b.WriteString(m.progress.ViewAs(percent))
	b.WriteString("\n\n")
	b.WriteString(helpLine(s.Paused))
	b.WriteString("\n")
	return b.String()
}

func rainbowWordmark(now time.Time) string {
	phase := int(now.UnixMilli()/120) % len(rainbow)
	lines := strings.Split(wordmark, "\n")
	var b strings.Builder

	for y, line := range lines {
		colorIndex := 0
		for _, r := range line {
			if r == ' ' {
				b.WriteRune(r)
				continue
			}
			color := rainbow[(colorIndex+y+phase)%len(rainbow)]
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(color).Render(string(r)))
			colorIndex++
		}
		if y < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func helpLine(paused bool) string {
	return mutedStyle.Render("[") +
		hotkeyStyle.Render("space") + mutedStyle.Render(" pause/resume") +
		mutedStyle.Render("]  [") +
		hotkeyStyle.Render("q") + mutedStyle.Render(" quit]")
}

func tick() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func formatDuration(d time.Duration) string {
	d = roundUpSecond(d)
	minutes := int(d / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func roundUpSecond(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return ((d + time.Second - 1) / time.Second) * time.Second
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
