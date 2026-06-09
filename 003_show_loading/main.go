package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bgColor        = lipgloss.Color("#121214")
	accentCyan     = lipgloss.Color("#00F5FF")
	accentPurple   = lipgloss.Color("#BD93F9")
	accentPink     = lipgloss.Color("#FF55FF")
	successGreen   = lipgloss.Color("#50FA7B")
	errorRed       = lipgloss.Color("#FF5555")
	textWhite      = lipgloss.Color("#FFFFFF")
	textGray       = lipgloss.Color("#7F7F7F")
	borderGray     = lipgloss.Color("#2A2A30")
	lightSlateGray = lipgloss.Color("#44475A")

	windowStyle = lipgloss.NewStyle().
			Background(bgColor).
			Foreground(textWhite).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(textWhite).
			Background(accentPurple).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(textGray).
			Italic(true).
			MarginTop(1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(borderGray).
			Foreground(textGray)

	activeTabStyle = tabStyle.
			Background(accentCyan).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)

	borderBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderGray).
			Padding(1, 3).
			Width(72).
			Height(14)
)

type tab int

const (
	tabSpinners tab = iota
	tabMarquee
	tabSkeleton
	tabTaskRunner
)

type taskState int

const (
	taskPending taskState = iota
	taskRunning
	taskCompleted
)

type taskItem struct {
	name     string
	duration int
	elapsed  int
	status   taskState
	details  string
}

type taskRunnerState struct {
	tasks     []taskItem
	activeIdx int
	running   bool
}

type model struct {
	activeTab    tab
	spinners     []spinner.Model
	marqueePos   int
	marqueeDir   int
	shimmerTick  int
	taskRunner   taskRunnerState
	windowWidth  int
	windowHeight int
}

type animationTickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func initTasks() []taskItem {
	return []taskItem{
		{name: "Resolving environments settings", duration: 15, status: taskPending},
		{name: "Fetching registry metadata", duration: 12, status: taskPending},
		{name: "Downloading required packages", duration: 25, status: taskPending, details: "0.0 MB / 12.8 MB"},
		{name: "Compiling source modules", duration: 20, status: taskPending},
		{name: "Verifying cryptographic signatures", duration: 10, status: taskPending},
		{name: "Deploying artifacts to production", duration: 15, status: taskPending},
	}
}

func initialModel() model {

	spin1 := spinner.New()
	spin1.Spinner = spinner.Dot
	spin1.Style = lipgloss.NewStyle().Foreground(accentCyan)

	spin2 := spinner.New()
	spin2.Spinner = spinner.Line
	spin2.Style = lipgloss.NewStyle().Foreground(accentPurple)

	spin3 := spinner.New()
	spin3.Spinner = spinner.MiniDot
	spin3.Style = lipgloss.NewStyle().Foreground(accentPink)

	return model{
		activeTab:  tabSpinners,
		spinners:   []spinner.Model{spin1, spin2, spin3},
		marqueePos: 0,
		marqueeDir: 1,
		taskRunner: taskRunnerState{
			tasks:     initTasks(),
			activeIdx: 0,
			running:   true,
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinners[0].Tick,
		m.spinners[1].Tick,
		m.spinners[2].Tick,
		tickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "1":
			m.activeTab = tabSpinners
		case "2":
			m.activeTab = tabMarquee
		case "3":
			m.activeTab = tabSkeleton
		case "4":
			m.activeTab = tabTaskRunner

		case "tab":
			m.activeTab = (m.activeTab + 1) % 4

		case "r", "R":
			if m.activeTab == tabTaskRunner {
				m.taskRunner.tasks = initTasks()
				m.taskRunner.activeIdx = 0
				m.taskRunner.running = true
			}
		}

	case animationTickMsg:

		m.marqueePos += m.marqueeDir
		if m.marqueePos >= 32 {
			m.marqueePos = 32
			m.marqueeDir = -1
		} else if m.marqueePos <= 0 {
			m.marqueePos = 0
			m.marqueeDir = 1
		}

		m.shimmerTick = (m.shimmerTick + 1) % 1000

		if m.taskRunner.running {
			active := &m.taskRunner.tasks[m.taskRunner.activeIdx]
			active.status = taskRunning
			active.elapsed++

			if m.taskRunner.activeIdx == 2 {
				prog := float64(active.elapsed) / float64(active.duration)
				mb := prog * 12.8
				if mb > 12.8 {
					mb = 12.8
				}
				active.details = fmt.Sprintf("%.1f MB / 12.8 MB", mb)
			}

			if active.elapsed >= active.duration {
				active.status = taskCompleted
				active.details = ""
				m.taskRunner.activeIdx++
				if m.taskRunner.activeIdx >= len(m.taskRunner.tasks) {
					m.taskRunner.running = false
				}
			}
		}

		cmds = append(cmds, tickCmd())

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}

	for i := range m.spinners {
		var cmd tea.Cmd
		m.spinners[i], cmd = m.spinners[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("CLI LOADING ANIMATIONS HARNESS"))
	s.WriteString("\n\n")

	tabs := []string{"[1] Spinners", "[2] Marquee Bar", "[3] Shimmer Skeletons", "[4] Task Runner"}
	var renderedTabs []string
	for i, t := range tabs {
		if int(m.activeTab) == i {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(t))
		}
	}
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
	s.WriteString("\n\n")

	var panelContent string
	switch m.activeTab {
	case tabSpinners:
		panelContent = m.viewSpinners()
	case tabMarquee:
		panelContent = m.viewMarquee()
	case tabSkeleton:
		panelContent = m.viewSkeleton()
	case tabTaskRunner:
		panelContent = m.viewTaskRunner()
	}

	s.WriteString(borderBoxStyle.Render(panelContent))
	s.WriteString("\n")

	s.WriteString(footerStyle.Render("⌨️  Tab: Switch tabs  •  1/2/3/4: Direct jump  •  R: Restart simulator (Tab 4)  •  Q/Esc: Exit"))

	return windowStyle.Render(s.String())
}
func (m model) viewSpinners() string {
	var s strings.Builder

	s.WriteString(lipgloss.NewStyle().Foreground(accentCyan).Bold(true).Render("Classic Spinner Components"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("  %s  %s\n\n", m.spinners[0].View(), lipgloss.NewStyle().Foreground(textWhite).Render("Loading repository files...")))
	s.WriteString(fmt.Sprintf("  %s  %s\n\n", m.spinners[1].View(), lipgloss.NewStyle().Foreground(accentPurple).Render("Synchronizing database connections...")))
	s.WriteString(fmt.Sprintf("  %s  %s\n", m.spinners[2].View(), lipgloss.NewStyle().Foreground(accentPink).Render("Generating optimization profiles...")))

	return s.String()
}

func (m model) viewMarquee() string {
	var s strings.Builder

	s.WriteString(lipgloss.NewStyle().Foreground(accentCyan).Bold(true).Render("Indeterminate Marquee Progress Loader"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(textGray).Render("Best used when the duration of a background task is unknown."))
	s.WriteString("\n\n\n")

	trackWidth := 40
	blockWidth := 8
	trackStyle := lipgloss.NewStyle().Foreground(borderGray)
	blockStyle := lipgloss.NewStyle().Foreground(accentCyan).Bold(true)

	var bar strings.Builder
	bar.WriteString(trackStyle.Render("["))
	for i := 0; i < trackWidth; i++ {
		if i >= m.marqueePos && i < m.marqueePos+blockWidth {
			bar.WriteString(blockStyle.Render("█"))
		} else {
			bar.WriteString(trackStyle.Render("░"))
		}
	}
	bar.WriteString(trackStyle.Render("]"))

	s.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(bar.String()))
	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Foreground(accentCyan).Italic(true).PaddingLeft(4).Render("status: waiting for job slot..."))

	return s.String()
}

func (m model) viewSkeleton() string {
	var s strings.Builder

	s.WriteString(lipgloss.NewStyle().Foreground(accentCyan).Bold(true).Render("Shimmering Skeleton Loader"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(textGray).Render("Simulates a loading layout using glowing waves of character blocks."))
	s.WriteString("\n\n")

	avatarShimmer := m.renderShimmerBar(4, 0)
	nameShimmer := m.renderShimmerBar(16, 2)
	handleShimmer := m.renderShimmerBar(10, 4)

	headerBlock := lipgloss.JoinHorizontal(
		lipgloss.Top,
		avatarShimmer+"\n"+avatarShimmer,
		"  ",
		lipgloss.JoinVertical(lipgloss.Left, nameShimmer, handleShimmer),
	)
	s.WriteString(headerBlock)
	s.WriteString("\n\n")

	s.WriteString(m.renderShimmerBar(50, 6) + "\n")
	s.WriteString(m.renderShimmerBar(44, 8) + "\n")
	s.WriteString(m.renderShimmerBar(56, 10) + "\n")
	s.WriteString(m.renderShimmerBar(24, 12))

	return s.String()
}

func (m model) renderShimmerBar(length int, lineOffset int) string {
	shades := []string{
		"#1A1A1E",
		"#25252A",
		"#3B3B42",
		"#585863",
		"#3B3B42",
		"#25252A",
	}

	var s strings.Builder
	for i := 0; i < length; i++ {
		idx := int(math.Abs(float64((m.shimmerTick + (i / 2) + lineOffset) % len(shades))))
		col := lipgloss.Color(shades[idx])
		s.WriteString(lipgloss.NewStyle().Foreground(col).Render("█"))
	}
	return s.String()
}

func (m model) viewTaskRunner() string {
	var s strings.Builder

	s.WriteString(lipgloss.NewStyle().Foreground(accentCyan).Bold(true).Render("Step-by-Step Task Runner Sim"))
	s.WriteString("   ")
	if m.taskRunner.running {
		s.WriteString(lipgloss.NewStyle().Foreground(accentPurple).Render("(Executing workflow...)"))
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(successGreen).Bold(true).Render("(✔ Success!)"))
	}
	s.WriteString("\n\n")

	for i, task := range m.taskRunner.tasks {
		var icon string
		var nameStyle lipgloss.Style

		switch task.status {
		case taskPending:
			icon = lipgloss.NewStyle().Foreground(textGray).Render("○")
			nameStyle = lipgloss.NewStyle().Foreground(textGray)
		case taskRunning:
			icon = m.spinners[0].View() // Use standard cyan dots spinner
			nameStyle = lipgloss.NewStyle().Foreground(accentCyan).Bold(true)
		case taskCompleted:
			icon = lipgloss.NewStyle().Foreground(successGreen).Render("✔")
			nameStyle = lipgloss.NewStyle().Foreground(textWhite)
		}

		s.WriteString(fmt.Sprintf("  %s  %s", icon, nameStyle.Render(task.name)))
		if task.details != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(accentPurple).Render(fmt.Sprintf(" (%s)", task.details)))
		}
		s.WriteString("\n")
		// Limit display lines to keep it snug
		if i >= 4 && i < len(m.taskRunner.tasks)-1 && len(m.taskRunner.tasks) > 5 {
			// Compact view if needed, but here it's 6 tasks so it fits perfectly!
		}
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running loading showcase: %v\n", err)
		os.Exit(1)
	}
}
