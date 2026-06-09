package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	accentCyan   = lipgloss.Color("#00F5FF")
	accentPurple = lipgloss.Color("#230152")
	accentGreen  = lipgloss.Color("#50FA7B")
	accentYellow = lipgloss.Color("#F1FA8C")
	errorRed     = lipgloss.Color("#FF5555")
	textWhite    = lipgloss.Color("#FFFFFF")
	textGray     = lipgloss.Color("#7F7F7F")
	borderGray   = lipgloss.Color("#2A2A30")
	infoBlue     = lipgloss.Color("#58A6FF")

	// Styles
	windowStyle = lipgloss.NewStyle().
			Foreground(textWhite)

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2, 0, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(textWhite).
			Background(accentPurple).
			Bold(true).
			Padding(1, 1).
			MarginTop(2).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(textGray).
			Italic(true).
			MarginTop(1)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(accentCyan).
				Bold(true)

	urlBoxStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2A2A30")).
			Padding(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(accentGreen)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorRed)

	infoStyle = lipgloss.NewStyle().
			Foreground(infoBlue)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(accentYellow)

	contentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderGray).
			Padding(1, 2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(accentCyan)
)

// Application state
type status int

const (
	statusIdle status = iota
	statusFetching
)

type httpResult struct {
	StatusCode    int
	ContentType   string
	ContentLength int
	Content       string
	Error         error
	Duration      time.Duration
}

type model struct {
	textInput textinput.Model
	spinner   spinner.Model
	status    status
	result    *httpResult
	width     int
	height    int
}

type fetchResultMsg struct {
	result httpResult
}

func initialModel() model {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = spinnerStyle

	boxBg := lipgloss.Color("#2A2A30")

	ti := textinput.New()
	ti.Placeholder = "https://example.com"
	ti.Focus()
	ti.CharLimit = 2048
	ti.Prompt = "URL: "
	ti.PromptStyle = inputPromptStyle.Copy().Background(boxBg)
	ti.TextStyle = lipgloss.NewStyle().Foreground(textWhite).Background(boxBg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(textGray).Background(boxBg)
	ti.Cursor.Style = cursorStyle.Copy().Background(boxBg)
	ti.SetValue("")

	return model{
		textInput: ti,
		spinner:   spin,
		status:    statusIdle,
		result:    nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.status == statusIdle && strings.TrimSpace(m.textInput.Value()) != "" {
				urlToFetch := m.textInput.Value()
				m.textInput.SetValue("")
				m.status = statusFetching
				return m, m.fetchURL(urlToFetch)
			}
		}

	case fetchResultMsg:
		m.status = statusIdle
		m.result = &msg.result

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update textInput only when idle
	if m.status == statusIdle {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update spinner
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) fetchURL(urlStr string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		// If user didn't specify protocol, prepend https://
		if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
			urlStr = "https://" + urlStr
		}

		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return fetchResultMsg{result: httpResult{Error: err}}
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			return fetchResultMsg{result: httpResult{Error: err, Duration: time.Since(start)}}
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		contentLength := 0

		var contentBuilder strings.Builder
		r := io.LimitReader(resp.Body, 1024*1024)
		n, err := io.Copy(&contentBuilder, r)
		contentLength = int(n)

		if err != nil && err != io.EOF {
			return fetchResultMsg{result: httpResult{
				Error:       err,
				StatusCode:  resp.StatusCode,
				ContentType: contentType,
				Duration:    time.Since(start),
			}}
		}

		// Sanitize content for display
		content := contentBuilder.String()
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")
		if len(content) > 500 {
			content = content[:500] + "..."
		}

		return fetchResultMsg{result: httpResult{
			StatusCode:    resp.StatusCode,
			ContentType:   contentType,
			ContentLength: contentLength,
			Content:       content,
			Error:         nil,
			Duration:      time.Since(start),
		}}
	}
}

func (m model) View() string {
	var top strings.Builder

	// Header
	top.WriteString(headerStyle.Render("HTTP CLIENT HARNESS"))
	top.WriteString("\n\n")

	// Response Section
	top.WriteString(infoStyle.Render("Response:"))
	top.WriteString("\n")

	var responseContent string
	contentWidth := m.width - 10
	if contentWidth < 20 {
		contentWidth = 70
	}

	if m.status == statusFetching {
		// Loading state
		responseContent = "🔄 Fetching URL...\n"
		responseContent += fmt.Sprintf("  %s Please wait...\n", m.spinner.View())
	} else if m.result != nil {
		// Completed/Last request result
		if m.result.Error != nil {
			responseContent = errorStyle.Render(fmt.Sprintf("❌ Error: %v", m.result.Error))
			responseContent += "\n\n"
			responseContent += infoStyle.Render(fmt.Sprintf("Status Code: %d", m.result.StatusCode))
			responseContent += fmt.Sprintf("\nDuration: %v", m.result.Duration)
		} else {
			responseContent = successStyle.Render("✅ Success")
			responseContent += "\n\n"
			responseContent += infoStyle.Render(fmt.Sprintf("Status Code: %d", m.result.StatusCode))
			responseContent += fmt.Sprintf("\nContent-Type: %s", m.result.ContentType)
			responseContent += fmt.Sprintf("\nLength: %d bytes", m.result.ContentLength)
			responseContent += fmt.Sprintf("\nDuration: %v", m.result.Duration)
			responseContent += "\n\n"
			responseContent += lipgloss.NewStyle().Foreground(textGray).Render("Content Preview:")
			responseContent += "\n"
			responseContent += m.result.Content
		}
	} else {
		// Idle state
		responseContent = "Ready to fetch. Enter a URL below and press Enter."
	}

	// Render response content inside the content box style with dynamic wrapping width
	top.WriteString(contentBoxStyle.Copy().Width(contentWidth).Render(responseContent))
	top.WriteString("\n\n")

	// Render top container with left/right padding
	renderedTop := containerStyle.Render(top.String())

	// Bottom URL Input Section (takes full width of the terminal)
	inputWidth := m.width
	if inputWidth < 20 {
		inputWidth = 80
	}

	urlBoxContent := ""
	if m.status == statusFetching {
		urlBoxContent = fmt.Sprintf("URL: %s Fetching...", m.spinner.View())
	} else {
		urlBoxContent = m.textInput.View()
	}

	renderedInput := urlBoxStyle.Copy().Width(inputWidth).Render(urlBoxContent)

	// Pin the input box to the bottom of the screen if terminal height is known
	topHeight := lipgloss.Height(renderedTop)
	inputHeight := lipgloss.Height(renderedInput)

	spacer := "\n"
	if m.height > 0 {
		spaceHeight := m.height - topHeight - inputHeight
		if spaceHeight > 0 {
			spacer = strings.Repeat("-\n", spaceHeight)
		}
	}

	// Join top content and bottom input field
	return windowStyle.Render(renderedTop + spacer + renderedInput + footerStyle.Render("⌨️  Type to enter URL • Press Enter to fetch • Esc to quit"))
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running HTTP client: %v\n", err)
		os.Exit(1)
	}
}
