package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI color scheme
var (
	highlightColor = lipgloss.Color("6") // Cyan
	subtleColor    = lipgloss.Color("8") // Gray
	errorColor     = lipgloss.Color("1") // Red
	successColor   = lipgloss.Color("2") // Green
)

// Identity represents a git identity with optional nickname
type Identity struct {
	Name     string
	Email    string
	Nickname string // Optional, falls back to name if empty
}

// encodeEmail converts email to git config section format
func encodeEmail(email string) string {
	return strings.ReplaceAll(strings.ReplaceAll(email, "@", "_at_"), ".", "_dot_")
}

// setNickname sets a nickname for an identity
func setNickname(email, nickname string) error {
	section := encodeEmail(email)
	nicknameCmd := fmt.Sprintf("identity.%s.nickname", section)
	return exec.Command("git", "config", "--global", nicknameCmd, nickname).Run()
}

// getNickname gets the nickname for an identity
func getNickname(email string) string {
	section := encodeEmail(email)
	nicknameCmd := fmt.Sprintf("identity.%s.nickname", section)
	out, err := exec.Command("git", "config", "--global", nicknameCmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// hasNickname checks if an identity has a nickname set
func hasNickname(email string) bool {
	return getNickname(email) != ""
}

// getIdentityDisplay formats an identity for display
func getIdentityDisplay(identity Identity) string {
	if identity.Nickname != "" {
		return fmt.Sprintf("%s (%s <%s>)", identity.Nickname, identity.Name, identity.Email)
	}
	return fmt.Sprintf("%s <%s>", identity.Name, identity.Email)
}

// getAllIdentities returns all identities as Identity structs
func getAllIdentities() []Identity {
	out, _ := exec.Command("git", "config", "--global", "--get-regexp", "^identity\\.").Output()
	var identities []Identity
	re := regexp.MustCompile(`identity\.(.+)\.name\s(.+)`)

	for _, line := range strings.Split(string(out), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 2 {
			section := matches[1]
			name := matches[2]
			emailCmd := fmt.Sprintf("identity.%s.email", section)
			emailOut, _ := exec.Command("git", "config", "--global", emailCmd).Output()
			email := strings.TrimSpace(string(emailOut))

			identity := Identity{
				Name:     name,
				Email:    email,
				Nickname: getNickname(email),
			}
			identities = append(identities, identity)
		}
	}
	return identities
}

// findIdentityByIdentifier finds an identity by nickname, name, or email with smart matching
func findIdentityByIdentifier(identifier string) (Identity, bool) {
	identities := getAllIdentities()

	// 1. Exact nickname match
	for _, identity := range identities {
		if identity.Nickname == identifier {
			return identity, true
		}
	}

	// 2. Exact email match
	for _, identity := range identities {
		if identity.Email == identifier {
			return identity, true
		}
	}

	// 3. Exact name match
	for _, identity := range identities {
		if identity.Name == identifier {
			return identity, true
		}
	}

	// 4. Partial email match (contains)
	for _, identity := range identities {
		if strings.Contains(identity.Email, identifier) {
			return identity, true
		}
	}

	// 5. Partial name match (contains)
	for _, identity := range identities {
		if strings.Contains(identity.Name, identifier) {
			return identity, true
		}
	}

	return Identity{}, false
}

type model struct {
	identities       []Identity
	cursor           int
	showConfirmation bool
	confirmChoices   []string
	confirmCursor    int
}

func initialModel() model {
	identities := getAllIdentities()
	return model{
		identities:       identities,
		cursor:           0,
		showConfirmation: false,
		confirmChoices:   []string{"Yes", "No"},
		confirmCursor:    1, // Default to "No"
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if m.confirmCursor == 0 { // Yes selected
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
				m.confirmCursor = 1 // Reset to "No"
			} else {
				if m.cursor >= len(m.identities) { // Add new identity
					addIdentityTUI()
					m.identities = getAllIdentities()
				} else {
					identity := m.identities[m.cursor]
					switchIdentity(identity.Name, identity.Email)
					return m, tea.Quit
				}
			}
		case "D":
			if m.cursor < len(m.identities) { // Not on "Add new identity"
				m.showConfirmation = true
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
				m.confirmCursor = 1 // Reset to "No"
			}
		}
	}
	return m, nil
}

func (m model) View() string {
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

	// Add "Add new identity" option
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

	// Confirmation dialog
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
		"↑/k up • ↓/j down • enter select • D delete • q quit\n" +
		"Confirmation: ←/→ navigate • enter confirm • esc cancel",
	)

	return style.Render(
		title + "\n\n" +
			strings.Join(items, "\n") +
			help,
	)
}

func addIdentity(name, email, nickname string) error {
	section := encodeEmail(email)

	nameCmd := fmt.Sprintf("identity.%s.name", section)
	emailCmd := fmt.Sprintf("identity.%s.email", section)

	if err := exec.Command("git", "config", "--global", nameCmd, name).Run(); err != nil {
		return fmt.Errorf("error setting name: %w", err)
	}
	if err := exec.Command("git", "config", "--global", emailCmd, email).Run(); err != nil {
		return fmt.Errorf("error setting email: %w", err)
	}

	// Set nickname if provided
	if nickname != "" {
		if err := setNickname(email, nickname); err != nil {
			return fmt.Errorf("error setting nickname: %w", err)
		}
	}

	return nil
}

func addIdentityTUI() {
	name := prompt("Enter name")
	email := prompt("Enter email")
	nickname := prompt("Enter nickname (optional)")

	if err := addIdentity(name, email, nickname); err != nil {
		fmt.Printf("Error adding identity: %v\n", err)
	}
}

func switchIdentity(name, email string) {
	if err := exec.Command("git", "config", "--global", "user.name", name).Run(); err != nil {
		fmt.Printf("Error setting user name: %v\n", err)
		return
	}
	if err := exec.Command("git", "config", "--global", "user.email", email).Run(); err != nil {
		fmt.Printf("Error setting user email: %v\n", err)
		return
	}
}

// switchIdentityByIdentifier switches to an identity using smart matching
func switchIdentityByIdentifier(identifier string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	switchIdentity(identity.Name, identity.Email)
	return nil
}

func deleteIdentity(email string) error {
	section := encodeEmail(email)

	nameCmd := fmt.Sprintf("identity.%s.name", section)
	emailCmd := fmt.Sprintf("identity.%s.email", section)
	nicknameCmd := fmt.Sprintf("identity.%s.nickname", section)

	if err := exec.Command("git", "config", "--global", "--unset", nameCmd).Run(); err != nil {
		return fmt.Errorf("error removing name: %w", err)
	}
	if err := exec.Command("git", "config", "--global", "--unset", emailCmd).Run(); err != nil {
		return fmt.Errorf("error removing email: %w", err)
	}
	// Remove nickname if it exists (ignore error if it doesn't exist)
	exec.Command("git", "config", "--global", "--unset", nicknameCmd).Run()

	return nil
}

func prompt(placeholder string) string {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Focus()

	p := tea.NewProgram(inputModel{
		textInput: input,
	})

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running prompt: %v\n", err)
		os.Exit(1)
	}

	result := m.(inputModel)
	if result.interrupted {
		os.Exit(0)
	}
	return result.value
}

type inputModel struct {
	textInput   textinput.Model
	value       string
	interrupted bool
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m inputModel) View() string {
	return lipgloss.NewStyle().
		Margin(0, 1).
		Render(m.textInput.View())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
