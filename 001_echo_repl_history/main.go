package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
			Background(secondaryColor).
			Foreground(lipgloss.Color("#000000"))

	helpStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true)
)

// Start the program
func main() {
	initialModel := model{
		history: []string{
			"hello world",
			"building custom inputs",
		},
	}

	// Create and start the bubbletea program
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error running the application: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	history      []string
	currentInput string
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

	// Wrap history items inside our beautiful rounded box
	//historyBox := historyBoxStyle.Render(s.String())

	// --- 2. Interactive Input Prompt ---
	promptSymbol := promptSymbolStyle.Render(" > ")

	// Create an active typing block with a simulated cursor
	inputText := m.currentInput
	// Render an accent color cursor block at the end of the text
	interactivePrompt := promptSymbol + inputText + cursorStyle.Render(" ")

	// --- 3. Help Instructions ---
	helpText := helpStyle.Render("💡 Commands: /pop (delete last) • /shift (delete first) • /delete <num> • /clean (delete all) • /quit /exit (exit)")

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

func (model model) Init() tea.Cmd {

	return nil

}

func (model model) Update(message tea.Msg) (tea.Model, tea.Cmd) {

	switch message := message.(type) {

	case tea.KeyMsg:
		switch message.Type {
		// Gracefully quit the application
		case tea.KeyCtrlC, tea.KeyEsc:
			return model, tea.Quit

		// Safe blackspace:
		case tea.KeyBackspace:
			runes := []rune(model.currentInput)
			if len(runes) > 0 {
				model.currentInput = string(runes[:len(runes)-1])
			}

		// Handle Enter Key
		case tea.KeyEnter:
			input := strings.TrimSpace(model.currentInput)
			model.currentInput = ""

			if input == "" {
				return model, nil
			}

			if strings.HasPrefix(input, "/") {
				if input == "/quit" || input == "/exit" {
					return model, tea.Quit
				}
				model.handleCommand(input)
			} else {
				model.history = append(model.history, input)
			}

		// Handler especial characters and symbols
		case tea.KeyRunes, tea.KeySpace:
			// ensure we accumulate runes properly
			model.currentInput += string(message.Runes)

		}
	}

	return model, nil

}

func (model *model) handleCommand(cmdStr string) {
	parts := strings.Fields(cmdStr)

	if len(parts) == 0 {
		return
	}

	command := parts[0]

	switch command {
	case "/pop":
		if len(model.history) > 0 {
			model.history = model.history[:len(model.history)-1]
		}

	case "/shift":
		if len(model.history) > 0 {
			model.history = model.history[1:]
		}

	case "/clean":
		model.history = []string{}

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

		if sliceIndex >= 0 && sliceIndex < len(model.history) {
			model.history = append(model.history[:sliceIndex], model.history[sliceIndex+1:]...)
		}
	}
}
