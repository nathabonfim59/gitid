package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlightColor = lipgloss.Color("6")
	subtleColor    = lipgloss.Color("8")
	errorColor     = lipgloss.Color("1")
	successColor   = lipgloss.Color("2")
)

func runTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel() Model {
	identities := getAllIdentities()
	return Model{
		identities:       identities,
		cursor:           0,
		showConfirmation: false,
		confirmChoices:   []string{"Yes", "No"},
		confirmCursor:    1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.identities) {
				m.cursor++
			}
		case "enter":
			if m.showConfirmation {
				if m.confirmCursor == 0 {
					email := m.identities[m.cursor].Email
					if err := deleteIdentity(email); err != nil {
						fmt.Printf("Error deleting identity: %v\n", err)
					} else {
						m.identities = getAllIdentities()
						if m.cursor >= len(m.identities) {
							m.cursor = len(m.identities)
						}
					}
				}
				m.showConfirmation = false
				m.confirmCursor = 1
			} else {
				if m.cursor >= len(m.identities) {
					addIdentityTUI()
					m.identities = getAllIdentities()
				} else {
					identity := m.identities[m.cursor]
					switchIdentity(identity.Name, identity.Email)
					return m, tea.Quit
				}
			}
		case "D":
			if m.cursor < len(m.identities) {
				m.showConfirmation = true
			}
		case "e":
			if m.cursor < len(m.identities) {
				editNicknameTUI(m.identities[m.cursor])
				m.identities = getAllIdentities()
				return m, tea.ClearScreen
			}
		case "E":
			if m.cursor < len(m.identities) {
				editFullIdentityTUI(m.identities[m.cursor])
				m.identities = getAllIdentities()
				return m, tea.ClearScreen
			}
		case "left", "h":
			if m.showConfirmation && m.confirmCursor > 0 {
				m.confirmCursor--
			}
		case "right", "l":
			if m.showConfirmation && m.confirmCursor < len(m.confirmChoices)-1 {
				m.confirmCursor++
			}
		case "esc":
			if m.showConfirmation {
				m.showConfirmation = false
				m.confirmCursor = 1
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	style := lipgloss.NewStyle().Margin(0, 1)
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(highlightColor).
		Render("Git Identity Manager")

	var items []string
	for i, identity := range m.identities {
		cursor := "  "
		displayText := getIdentityDisplay(identity)
		if m.cursor == i {
			cursor = "▸ "
			displayText = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Render(displayText)
		}
		items = append(items, fmt.Sprintf("%s%s", cursor, displayText))
	}

	cursor := "  "
	displayText := "Add new identity"
	if m.cursor >= len(m.identities) {
		cursor = "▸ "
		displayText = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Render(displayText)
	}
	items = append(items, fmt.Sprintf("%s%s", cursor, displayText))

	if m.showConfirmation {
		confirmMsg := lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Render("\nAre you sure you want to delete this identity?")

		var choices []string
		for i, choice := range m.confirmChoices {
			if i == m.confirmCursor {
				choice = lipgloss.NewStyle().
					Background(highlightColor).
					Foreground(lipgloss.Color("0")).
					Bold(true).
					Render(" " + choice + " ")
			} else {
				choice = lipgloss.NewStyle().
					Foreground(subtleColor).
					Render(" " + choice + " ")
			}
			choices = append(choices, choice)
		}

		items = append(items,
			confirmMsg,
			"\n"+strings.Join(choices, " "),
		)
	}

	helpStyle := lipgloss.NewStyle().Foreground(subtleColor)
	help := helpStyle.Render("\n" +
		"↑/k up • ↓/j down • enter select • D delete • e edit nickname • E edit full • q quit\n" +
		"Confirmation: ←/→ navigate • enter confirm • esc cancel",
	)

	return style.Render(
		title + "\n\n" +
			strings.Join(items, "\n") +
			help,
	)
}

func addIdentityTUI() {
	name := prompt("Enter name")
	email := prompt("Enter email")
	nickname := prompt("Enter nickname (optional)")

	if err := addIdentity(name, email, nickname); err != nil {
		fmt.Printf("Error adding identity: %v\n", err)
	}
}

func editNicknameTUI(identity Identity) {
	currentNickname := getNickname(identity.Email)
	if currentNickname == "" {
		currentNickname = "(none)"
	}

	fmt.Printf("Current nickname for %s: %s\n", identity.Name, currentNickname)
	newNickname := prompt("Enter new nickname (leave empty to remove)")

	if err := setNickname(identity.Email, newNickname); err != nil {
		fmt.Printf("Error setting nickname: %v\n", err)
	}
}

func editFullIdentityTUI(identity Identity) {
	fmt.Printf("Editing identity: %s\n", getIdentityDisplay(identity))

	newName := prompt("Enter name (" + identity.Name + ")")
	if newName == "" {
		newName = identity.Name
	}

	newEmail := prompt("Enter email (" + identity.Email + ")")
	if newEmail == "" {
		newEmail = identity.Email
	}

	currentNickname := getNickname(identity.Email)
	nicknamePrompt := "Enter nickname"
	if currentNickname != "" {
		nicknamePrompt += " (" + currentNickname + ")"
	}
	nicknamePrompt += " (leave empty to keep current)"
	newNickname := prompt(nicknamePrompt)
	if newNickname == "" {
		newNickname = currentNickname
	}

	if err := updateIdentity(identity.Email, newName, newEmail, newNickname); err != nil {
		fmt.Printf("Error updating identity: %v\n", err)
	}
}

func prompt(placeholder string) string {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Focus()

	p := tea.NewProgram(InputModel{
		textInput: input,
	})

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running prompt: %v\n", err)
		os.Exit(1)
	}

	result := m.(InputModel)
	if result.interrupted {
		os.Exit(0)
	}
	return result.value
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.value = m.textInput.Value()
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.interrupted = true
			return m, tea.Quit
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.textInput.View())
}
