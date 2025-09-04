package main

// tui_intro.go
import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type model struct {
	choice      string
	bool_answer bool
}

// Define style: crimson red text, faded yellow background
var style = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")).
	Background(lipgloss.Color("228")).
	Bold(true).
	PaddingTop(1).
	PaddingLeft(1).
	Width(50).
	BorderStyle(lipgloss.RoundedBorder())

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c", "q":
			return m, tea.Quit
		case "y":
			m.choice = "Yes"
			m.bool_answer = true
			return m, tea.Quit
		case "n":
			m.choice = "No"
			m.bool_answer = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.choice == "" {
		return style.Render("Do you want to integrate GitAegis into your git? (y/n)")
	}
	return style.Render(fmt.Sprintf("You selected: %s", m.choice))
}

func show_menu() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
