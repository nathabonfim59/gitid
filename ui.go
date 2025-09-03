package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/posener/complete/v2/install"
)

var (
	highlightColor = lipgloss.Color("6")
	subtleColor    = lipgloss.Color("8")
	errorColor     = lipgloss.Color("1")
	successColor   = lipgloss.Color("2")
)

func runTUI() {
	// Check if user has no identities and completion not installed
	identities := getAllIdentities()
	if len(identities) == 0 && shouldPromptForCompletion() {
		if runCompletionPrompt() {
			// After completion prompt, continue to main TUI
		}
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func initialModel() Model {
	identities := getAllIdentities()

	isInRepo := isInGitRepository()
	hasLocal := hasLocalIdentity()
	var localName, localEmail string

	if hasLocal {
		localName, localEmail, _ = getCurrentLocalIdentity()
	}

	return Model{
		identities:       identities,
		cursor:           0,
		showConfirmation: false,
		confirmChoices:   []string{"Yes", "No"},
		confirmCursor:    1,
		isInGitRepo:      isInRepo,
		hasLocalIdentity: hasLocal,
		localName:        localName,
		localEmail:       localEmail,
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
		case "r":
			if m.cursor < len(m.identities) && m.isInGitRepo {
				identity := m.identities[m.cursor]
				if err := setLocalIdentity(identity.Name, identity.Email); err != nil {
					fmt.Printf("Error setting local identity: %v\n", err)
				} else {
					// Update local identity info
					m.hasLocalIdentity = true
					m.localName = identity.Name
					m.localEmail = identity.Email
				}
			}
		case "R":
			if m.hasLocalIdentity && m.isInGitRepo {
				// Unset local identity
				exec.Command("git", "config", "--unset", "user.name").Run()
				exec.Command("git", "config", "--unset", "user.email").Run()
				m.hasLocalIdentity = false
				m.localName = ""
				m.localEmail = ""
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

	// Add repository status
	var repoStatus string
	if m.isInGitRepo {
		if m.hasLocalIdentity {
			localNickname := getNickname(m.localEmail)
			if localNickname != "" {
				repoStatus = fmt.Sprintf("Repository Identity: %s (%s <%s>)", localNickname, m.localName, m.localEmail)
			} else {
				repoStatus = fmt.Sprintf("Repository Identity: %s <%s>", m.localName, m.localEmail)
			}
			repoStatus = lipgloss.NewStyle().
				Foreground(successColor).
				Render(repoStatus)
		} else {
			repoStatus = lipgloss.NewStyle().
				Foreground(subtleColor).
				Render("Repository: Using global identity")
		}
	}

	var items []string
	for i, identity := range m.identities {
		cursor := "  "
		displayText := getIdentityDisplay(identity)

		// Check if this is the current local identity
		isCurrentLocal := m.hasLocalIdentity && identity.Name == m.localName && identity.Email == m.localEmail
		if isCurrentLocal {
			displayText += lipgloss.NewStyle().
				Foreground(successColor).
				Render(" [local]")
		}

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
	var helpText string
	if m.isInGitRepo {
		helpText = "↑/k up • ↓/j down • enter select/set global • r set local • R unset local\n" +
			"D delete • e edit nickname • E edit full • q quit"
	} else {
		helpText = "↑/k up • ↓/j down • enter select/set global • D delete • e edit nickname • E edit full • q quit"
	}
	helpText += "\nConfirmation: ←/→ navigate • enter confirm • esc cancel"
	help := helpStyle.Render("\n" + helpText)

	result := title
	if repoStatus != "" {
		result += "\n" + repoStatus
	}
	result += "\n\n" + strings.Join(items, "\n") + help

	return style.Render(result)
}

func shouldPromptForCompletion() bool {
	// Check if completion is already installed
	if install.IsInstalled("gitid") {
		return false
	}

	// Check if we can detect the shell
	shell := detectCurrentShell()
	return shell != ""
}

func runCompletionPrompt() bool {
	shell := detectCurrentShell()
	if shell == "" {
		return false
	}

	model := CompletionPromptModel{
		shell:   shell,
		choices: []string{"Yes", "No"},
		cursor:  0,
	}

	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		fmt.Printf("Error running completion prompt: %v\n", err)
		return false
	}

	finalModel := result.(CompletionPromptModel)
	return finalModel.shouldInstall
}

func (m CompletionPromptModel) Init() tea.Cmd {
	return nil
}

func (m CompletionPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.finished = true
			return m, tea.Quit
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.shouldInstall = (m.cursor == 0)
			m.finished = true

			if m.shouldInstall {
				if err := install.Install("gitid"); err != nil {
					fmt.Printf("Failed to install completion: %v\n", err)
				}
			}

			return m, tea.Quit
		}
	}
	return m, nil
}

func (m CompletionPromptModel) View() string {
	if m.finished {
		if m.shouldInstall {
			return lipgloss.NewStyle().
				Foreground(successColor).
				Render("Shell completion installed for " + m.shell + "!\n")
		}
		return ""
	}

	style := lipgloss.NewStyle().Margin(0, 1)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(highlightColor).
		Render("Shell Completion Setup")

	message := fmt.Sprintf("Would you like to install shell completion for %s?", m.shell)

	var choices []string
	for i, choice := range m.choices {
		if i == m.cursor {
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

	help := lipgloss.NewStyle().
		Foreground(subtleColor).
		Render("\n←/→ navigate • enter confirm • q quit")

	return style.Render(
		title + "\n\n" +
			message + "\n\n" +
			strings.Join(choices, " ") +
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
