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
    highlightColor = lipgloss.Color("6")  // Cyan
    subtleColor    = lipgloss.Color("8")  // Gray
    errorColor     = lipgloss.Color("1")  // Red
    successColor   = lipgloss.Color("2")  // Green
)

type model struct {
    identities []string
    cursor     int
    showConfirmation bool
    confirmChoices   []string
    confirmCursor    int
}

func initialModel() model {
    identities := listIdentities()
    identities = append(identities, "Add new identity")
    return model{
        identities:       identities,
        cursor:          0,
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
            if m.cursor < len(m.identities)-1 {
                m.cursor++
            }
        case "enter":
            if m.showConfirmation {
                if m.confirmCursor == 0 { // Yes selected
                    _, email := parseIdentity(m.identities[m.cursor])
                    if err := deleteIdentity(email); err != nil {
                        fmt.Printf("Error deleting identity: %v\n", err)
                    } else {
                        m.identities = listIdentities()
                        m.identities = append(m.identities, "Add new identity")
                        if m.cursor >= len(m.identities) {
                            m.cursor = len(m.identities) - 1
                        }
                    }
                }
                m.showConfirmation = false
                m.confirmCursor = 1 // Reset to "No"
            } else {
                if m.identities[m.cursor] == "Add new identity" {
                    addIdentity()
                    m.identities = listIdentities()
                    m.identities = append(m.identities, "Add new identity")
                } else {
                    name, email := parseIdentity(m.identities[m.cursor])
                    switchIdentity(name, email)
                    return m, tea.Quit
                }
            }
        case "D":
            if m.identities[m.cursor] != "Add new identity" {
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
        if m.cursor == i {
            cursor = "▸ "
            if identity == "Add new identity" {
                identity = lipgloss.NewStyle().
                    Foreground(successColor).
                    Bold(true).
                    Render(identity)
            } else {
                identity = lipgloss.NewStyle().
                    Foreground(highlightColor).
                    Bold(true).
                    Render(identity)
            }
        }
        items = append(items, fmt.Sprintf("%s%s", cursor, identity))
    }

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

func addIdentity() {
    name := prompt("Enter name")
    email := prompt("Enter email")
    section := strings.ReplaceAll(strings.ReplaceAll(email, "@", "_at_"), ".", "_dot_")

    nameCmd := fmt.Sprintf("identity.%s.name", section)
    emailCmd := fmt.Sprintf("identity.%s.email", section)

    if err := exec.Command("git", "config", "--global", nameCmd, name).Run(); err != nil {
        fmt.Printf("Error setting name: %v\n", err)
        return
    }
    if err := exec.Command("git", "config", "--global", emailCmd, email).Run(); err != nil {
        fmt.Printf("Error setting email: %v\n", err)
        return
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

func listIdentities() []string {
    out, _ := exec.Command("git", "config", "--global", "--get-regexp",
        "^identity\\.").Output()
    var identities []string
    re := regexp.MustCompile(`identity\.(.+)\.name\s(.+)`)

    for _, line := range strings.Split(string(out), "\n") {
        matches := re.FindStringSubmatch(line)
        if len(matches) > 2 {
            section := matches[1]
            name := matches[2]
            emailCmd := fmt.Sprintf("identity.%s.email", section)
            email, _ := exec.Command("git", "config", "--global", emailCmd).Output()
            identity := fmt.Sprintf("%s (%s)", name, strings.TrimSpace(string(email)))
            identities = append(identities, identity)
        }
    }
    return identities
}

func deleteIdentity(email string) error {
    section := strings.ReplaceAll(strings.ReplaceAll(email, "@", "_at_"), ".", "_dot_")

    nameCmd := fmt.Sprintf("identity.%s.name", section)
    emailCmd := fmt.Sprintf("identity.%s.email", section)

    if err := exec.Command("git", "config", "--global", "--unset", nameCmd).Run(); err != nil {
        return fmt.Errorf("error removing name: %w", err)
    }
    if err := exec.Command("git", "config", "--global", "--unset", emailCmd).Run(); err != nil {
        return fmt.Errorf("error removing email: %w", err)
    }
    return nil
}
func parseIdentity(identity string) (string, string) {
    parts := strings.Split(identity, "(")
    name := strings.TrimSpace(parts[0])
    email := strings.TrimSuffix(parts[1], ")")
    return name, email
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
    textInput  textinput.Model
    value      string
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
