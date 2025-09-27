package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices  []string
	cursor   int
	selected int // -1 means nothing selected yet
}

func initialModel() model {
	return model{
		choices:  []string{"Yes", "No"},
		selected: -1,
	}
}

// simple style (you can adjust as needed)
var style = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")).
	Background(lipgloss.Color("228")).
	Bold(true).
	Padding(1, 2).
	Border(lipgloss.RoundedBorder())

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true)

	// Apply it to your header
	s := headerStyle.Render("We recognise a gitignore file, do you want us to use this ?") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if m.selected == i {
			checked = "*"
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"
	return style.Render(s)
}

func integrate_gitignore() {
	_, err := os.Stat(".gitignore")
	// use alternate screen
	if err == nil {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}

		// cast model back
		m := finalModel.(model)

		if m.selected != -1 {
			fmt.Println("You chose:", m.choices[m.selected])
		}
	}
}
