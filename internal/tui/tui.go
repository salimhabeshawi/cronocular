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

var greenGradient = []rgb{
	{r: 0x14, g: 0xB8, b: 0xA6},
	{r: 0x22, g: 0xC5, b: 0x5E},
	{r: 0xA3, g: 0xE6, b: 0x35},
}

const wordmark = `                                                                      
 ████  █████   ████  █    █  ████   ████  █    █ █        ██   █████  
█    █ █    █ █    █ ██   █ █    █ █    █ █    █ █       █  █  █    █ 
█      █    █ █    █ █ █  █ █    █ █      █    █ █      █    █ █    █ 
█      █████  █    █ █  █ █ █    █ █      █    █ █      ██████ █████  
█    █ █   █  █    █ █   ██ █    █ █    █ █    █ █      █    █ █   █  
 ████  █    █  ████  █    █  ████   ████   ████  ██████ █    █ █    █ 
                                                                      `

func Run(ctx context.Context, controller *timer.Controller) error {
	p := tea.NewProgram(newModel(controller), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

func newModel(controller *timer.Controller) model {
	bar := progress.New(
		progress.WithGradient("#14B8A6", "#A3E635"),
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
	b.WriteString(gradientWordmark())
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

type rgb struct {
	r int
	g int
	b int
}

func gradientWordmark() string {
	lines := strings.Split(wordmark, "\n")
	width := 1
	for _, line := range lines {
		lineWidth := len([]rune(line))
		if lineWidth > width {
			width = lineWidth
		}
	}

	var b strings.Builder

	for y, line := range lines {
		x := 0
		for _, r := range line {
			if r == ' ' {
				b.WriteRune(r)
				x++
				continue
			}
			color := gradientColor(float64(x) / float64(width-1))
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(color).Render(string(r)))
			x++
		}
		if y < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func gradientColor(t float64) lipgloss.Color {
	if t <= 0 {
		return greenGradient[0].color()
	}
	if t >= 1 {
		return greenGradient[len(greenGradient)-1].color()
	}

	scaled := t * float64(len(greenGradient)-1)
	i := int(scaled)
	local := scaled - float64(i)
	return mix(greenGradient[i], greenGradient[i+1], local).color()
}

func mix(a, b rgb, t float64) rgb {
	return rgb{
		r: int(float64(a.r) + (float64(b.r)-float64(a.r))*t),
		g: int(float64(a.g) + (float64(b.g)-float64(a.g))*t),
		b: int(float64(a.b) + (float64(b.b)-float64(a.b))*t),
	}
}

func (c rgb) color() lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b))
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
