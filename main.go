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

type model struct {
    identities []string
    cursor     int
}

func initialModel() model {
    identities := listIdentities()
    identities = append(identities, "Add new identity")
    return model{
        identities: identities,
        cursor:     0,
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
    }
    return m, nil
}

func (m model) View() string {
    style := lipgloss.NewStyle().Margin(0, 1)
    title := lipgloss.NewStyle().Bold(true).Render("Git Identity Manager")
    
    var items []string
    for i, identity := range m.identities {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
            identity = lipgloss.NewStyle().Bold(true).Render(identity)
        }
        items = append(items, fmt.Sprintf("%s %s", cursor, identity))
    }
    
    help := "\n↑/k up • ↓/j down • q quit"
    
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

    return m.(inputModel).value
}

type inputModel struct {
    textInput textinput.Model
    value     string
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
