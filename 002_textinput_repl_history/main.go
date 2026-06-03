package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Harmonious color palette definitions
	secondaryColor = lipgloss.Color("#00F5FF") // Electric Cyan
	grayColor      = lipgloss.Color("#7F7F7F") // Dim Gray
	lightGrayColor = lipgloss.Color("#E5E5E5") // Light off-white

	// Box borders & container styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 0)

	historyBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, false, false, true, true).
			Padding(1, 2)

	promptSymbolStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true)
)

// Start the program
func main() {
	ta := textarea.New()
	ta.Placeholder = "Type command or message here..."
	ta.Focus()
	ta.Prompt = "" // No prompt inside textarea lines
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Disable default enter newline

	// Apply identical elegant styling via textarea API
	ta.FocusedStyle, ta.BlurredStyle = textarea.DefaultStyles()
	bgDark := lipgloss.Color("#1A1A1A")

	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(bgDark).Padding(0, 1)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lightGrayColor).Background(bgDark)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(bgDark)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(grayColor).Background(bgDark)

	ta.BlurredStyle.Base = lipgloss.NewStyle().Background(bgDark).Padding(0, 1)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(grayColor).Background(bgDark)
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle().Background(bgDark)
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(grayColor).Background(bgDark)

	ta.Cursor.Style = cursorStyle

	ta.SetHeight(3)
	ta.SetWidth(60)

	initialModel := model{
		history: []string{
			"hello world",
			"building custom inputs",
		},
		input: ta,
		width: 60,
	}

	// Create and start the bubbletea program
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error running the application: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	history []string
	input   textarea.Model
	width   int
}

func (m model) View() string {
	var s strings.Builder

	// --- 1. History Section ---
	s.WriteString(titleStyle.Render("HISTORY LOG"))
	s.WriteString("\n\n")

	if len(m.history) == 0 {
		s.WriteString(helpStyle.Render("   (Empty history. Type below to add items...)"))
		s.WriteString("\n")
	} else {
		for i, entry := range m.history {
			// Enumerate starting from 1 (aligned neatly with leading zeroes)
			numStr := fmt.Sprintf("%02d. ", i+1)

			// Highlight history content and indexes nicely
			styledNum := lipgloss.NewStyle().Foreground(secondaryColor).Render(numStr)
			styledEntry := lipgloss.NewStyle().Foreground(lightGrayColor).Render(entry)

			s.WriteString(styledNum)
			s.WriteString(styledEntry)
			s.WriteString("\n")
		}
	}

	// --- 2. Interactive Input Prompt ---
	bgDark := lipgloss.Color("#1A1A1A")
	textareaWidth := m.width - 3
	if textareaWidth < 10 {
		textareaWidth = 10
	}

	// 1. Wrap the view in a box with padding, but render the padding in the dark background color
	inputBoxStyle := lipgloss.NewStyle().
		Padding(1, 1).
		Background(bgDark)

	// 2. Apply the dark background color to the actual text content
	textareaContainerStyle := lipgloss.NewStyle().
		Background(bgDark).
		Width(textareaWidth)

	// We replace all ANSI reset sequences with reset + re-applying the background color
	// to make sure spaces padded internally by textarea carry the background color.
	rawView := m.input.View()
	textareaView := strings.ReplaceAll(rawView, "\x1b[0m", "\x1b[0m\x1b[48;2;26;26;26m")

	promptSymbol := promptSymbolStyle.PaddingTop(1).Render(" > ")
	// Render prompt and textarea side-by-side
	interactivePrompt := lipgloss.JoinHorizontal(
		lipgloss.Top,
		promptSymbol,
		inputBoxStyle.Render(textareaContainerStyle.Render(textareaView)),
	)

	// --- 3. Help Instructions ---
	helpText := helpStyle.Render("💡 Commands: Enter (Submit) • Ctrl+N (Newline) • /pop (delete last) • /shift (delete first) • /delete <num> • /clean (delete all) • /quit /exit (exit)")

	// Stitch everything together into a neat, unified terminal screen
	return lipgloss.JoinVertical(
		lipgloss.Left,
		historyBoxStyle.Render(s.String()),
		"",
		interactivePrompt,
		helpText,
		"",
	)
}

func (m model) Init() tea.Cmd {
	// Enable cursor blinking
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Gracefully quit the application
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Handle standard Enter: Submit
		case "enter":
			val := strings.TrimSpace(m.input.Value())
			m.input.Reset() // Clear input field

			if val == "" {
				return m, nil
			}

			if strings.HasPrefix(val, "/") {
				if val == "/quit" || val == "/exit" {
					return m, tea.Quit
				}
				m.handleCommand(val)
			} else {
				m.history = append(m.history, val)
			}
			return m, nil

		// Handle Shift+Enter or Ctrl+N: Insert Newline
		case "shift+enter", "ctrl+n":
			m.input.KeyMap.InsertNewline.SetEnabled(true)
			syntheticMsg := tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}}
			m.input, cmd = m.input.Update(syntheticMsg)
			m.input.KeyMap.InsertNewline.SetEnabled(false)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		// Take all the width that the screen allows, minus prompt offset (3 chars)
		m.input.SetWidth(msg.Width - 3)
	}

	// Forward key and timer (blink) events to textarea model
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) handleCommand(cmdStr string) {
	parts := strings.Fields(cmdStr)

	if len(parts) == 0 {
		return
	}

	command := parts[0]

	switch command {
	case "/pop":
		if len(m.history) > 0 {
			m.history = m.history[:len(m.history)-1]
		}

	case "/shift":
		if len(m.history) > 0 {
			m.history = m.history[1:]
		}

	case "/clean":
		m.history = []string{}

	case "/delete":
		if len(parts) < 2 {
			// ignore if no index provided
			return
		}

		index, err := strconv.Atoi(parts[1])
		if err != nil {
			// if is not a valid int, ignore it
			return
		}

		sliceIndex := index - 1

		if sliceIndex >= 0 && sliceIndex < len(m.history) {
			m.history = append(m.history[:sliceIndex], m.history[sliceIndex+1:]...)
		}
	}
}
