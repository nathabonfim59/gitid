package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/posener/complete/v2"
	"github.com/posener/complete/v2/install"
	"github.com/posener/complete/v2/predict"
)

func predictIdentities(prefix string) []string {
	identities := getAllIdentities()
	var suggestions []string

	for _, identity := range identities {
		if identity.Nickname != "" {
			suggestions = append(suggestions, identity.Nickname)
		}
		suggestions = append(suggestions, identity.Name)
		suggestions = append(suggestions, identity.Email)
	}

	return suggestions
}

func setupCompletion() {
	cmd := &complete.Command{
		Sub: map[string]*complete.Command{
			"list":     {},
			"current":  {},
			"switch":   {Args: complete.PredictFunc(predictIdentities)},
			"use":      {Args: complete.PredictFunc(predictIdentities)},
			"add":      {},
			"delete":   {Args: complete.PredictFunc(predictIdentities)},
			"nickname": {Args: complete.PredictFunc(predictIdentities)},
			"repo": {
				Sub: map[string]*complete.Command{
					"current": {},
					"use":     {Args: complete.PredictFunc(predictIdentities)},
					"add":     {},
				},
			},
			"completion": {Args: predict.Set{"bash", "zsh", "fish"}, Flags: map[string]complete.Predictor{"r": predict.Nothing}},
			"help":       {},
		},
		Flags: map[string]complete.Predictor{
			"h":    predict.Nothing,
			"help": predict.Nothing,
		},
	}

	cmd.Complete("gitid")
}

func handleCLICommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	command := args[0]
	switch command {
	case "list":
		return listIdentitiesCLI()
	case "current":
		return getCurrentIdentityCLI()
	case "switch", "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: gitid %s <identifier>", command)
		}
		return switchIdentityCLI(args[1])
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: gitid add <name> <email> [nickname]")
		}
		nickname := ""
		if len(args) > 3 {
			nickname = args[3]
		}
		return addIdentityCLI(args[1], args[2], nickname)
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: gitid delete <identifier>")
		}
		return deleteIdentityCLI(args[1])
	case "nickname":
		if len(args) < 3 {
			return fmt.Errorf("usage: gitid nickname <identifier> <nickname>")
		}
		return setNicknameCLI(args[1], args[2])
	case "completion":
		return completionCLI(args[1:])
	case "repo":
		if len(args) < 2 {
			return fmt.Errorf("usage: gitid repo <list|use|current> [identifier]")
		}
		return repoCLI(args[1:])
	case "help", "--help", "-h":
		showHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nRun 'gitid help' for usage information", command)
	}
}

func listIdentitiesCLI() error {
	identities := getAllIdentities()
	if len(identities) == 0 {
		fmt.Println("No identities configured.")
		return nil
	}

	var localName, localEmail string
	hasLocal := hasLocalIdentity()
	if hasLocal {
		localName, localEmail, _ = getCurrentLocalIdentity()
	}

	for _, identity := range identities {
		nickname := identity.Nickname
		if nickname == "" {
			nickname = "-"
		}

		// Check if this identity is the current local identity
		isLocal := hasLocal && identity.Name == localName && identity.Email == localEmail
		localIndicator := ""
		if isLocal {
			localIndicator = " [current local]"
		}

		fmt.Printf("%-12s %s <%s>%s\n", nickname, identity.Name, identity.Email, localIndicator)
	}

	if isInGitRepository() && !hasLocal {
		fmt.Printf("\n(In git repository using global identity)\n")
	}

	return nil
}

func getCurrentIdentityCLI() error {
	nameOut, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		return fmt.Errorf("no global git identity configured")
	}
	emailOut, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		return fmt.Errorf("no global git identity configured")
	}

	name := strings.TrimSpace(string(nameOut))
	email := strings.TrimSpace(string(emailOut))

	nickname := getNickname(email)
	if nickname != "" {
		fmt.Printf("Global: %s (%s <%s>)\n", nickname, name, email)
	} else {
		fmt.Printf("Global: %s <%s>\n", name, email)
	}

	// If we're in a git repository and have a local identity, show it too
	if hasLocalIdentity() {
		localName, localEmail, err := getCurrentLocalIdentity()
		if err == nil {
			localNickname := getNickname(localEmail)
			if localNickname != "" {
				fmt.Printf("Local:  %s (%s <%s>)\n", localNickname, localName, localEmail)
			} else {
				fmt.Printf("Local:  %s <%s>\n", localName, localEmail)
			}
		}
	} else if isInGitRepository() {
		fmt.Printf("Local:  (using global identity)\n")
	}

	return nil
}

func switchIdentityCLI(identifier string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	switchIdentity(identity.Name, identity.Email)
	display := getIdentityDisplay(identity)
	fmt.Printf("Switched to %s\n", display)
	return nil
}

func addIdentityCLI(name, email, nickname string) error {
	if err := addIdentity(name, email, nickname); err != nil {
		return err
	}

	identity := Identity{Name: name, Email: email, Nickname: nickname}
	display := getIdentityDisplay(identity)
	fmt.Printf("Added identity: %s\n", display)
	return nil
}

func deleteIdentityCLI(identifier string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	if err := deleteIdentity(identity.Email); err != nil {
		return err
	}

	display := getIdentityDisplay(identity)
	fmt.Printf("Deleted identity: %s\n", display)
	return nil
}

func setNicknameCLI(identifier, nickname string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	if err := setNickname(identity.Email, nickname); err != nil {
		return err
	}

	fmt.Printf("Set nickname \"%s\" for %s <%s>\n", nickname, identity.Name, identity.Email)
	return nil
}

func completionCLI(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gitid completion <shell> [-r]\nSupported shells: bash, zsh, fish")
	}

	shell := args[0]
	remove := false

	// Check for -r flag
	if len(args) > 1 && args[1] == "-r" {
		remove = true
	}

	// Validate shell
	validShells := map[string]bool{"bash": true, "zsh": true, "fish": true}
	if !validShells[shell] {
		return fmt.Errorf("unsupported shell: %s\nSupported shells: bash, zsh, fish", shell)
	}

	if remove {
		if !install.IsInstalled("gitid") {
			fmt.Printf("Completion not installed for %s\n", shell)
			return nil
		}
		if err := install.Uninstall("gitid"); err != nil {
			return fmt.Errorf("failed to uninstall completion: %v", err)
		}
		fmt.Printf("Successfully removed completion for %s\n", shell)
	} else {
		if install.IsInstalled("gitid") {
			fmt.Printf("Completion already installed for %s\n", shell)
			return nil
		}
		if err := install.Install("gitid"); err != nil {
			return fmt.Errorf("failed to install completion: %v", err)
		}
		fmt.Printf("Successfully installed completion for %s\n", shell)
		fmt.Println("Please restart your shell or run: source ~/.bashrc (or ~/.zshrc)")
	}

	return nil
}

func detectCurrentShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}

	shellName := filepath.Base(shell)
	switch shellName {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	default:
		return ""
	}
}

func repoCLI(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gitid repo <list|use|current|add> [args]")
	}

	subcommand := args[0]
	switch subcommand {
	case "current":
		return getCurrentLocalIdentityCLI()
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: gitid repo use <identifier>")
		}
		return useLocalIdentityCLI(args[1])
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: gitid repo add <name> <email> [nickname]")
		}
		nickname := ""
		if len(args) > 3 {
			nickname = args[3]
		}
		return addLocalIdentityCLI(args[1], args[2], nickname)
	default:
		return fmt.Errorf("unknown repo subcommand: %s", subcommand)
	}
}

func getCurrentLocalIdentityCLI() error {
	name, email, err := getCurrentLocalIdentity()
	if err != nil {
		return err
	}

	nickname := getNickname(email)
	if nickname != "" {
		fmt.Printf("%s (%s <%s>) [local]\n", nickname, name, email)
	} else {
		fmt.Printf("%s <%s> [local]\n", name, email)
	}
	return nil
}

func useLocalIdentityCLI(identifier string) error {
	if err := setLocalIdentityByIdentifier(identifier); err != nil {
		return err
	}

	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	display := getIdentityDisplay(identity)
	fmt.Printf("Set local repository identity to %s\n", display)
	return nil
}

func addLocalIdentityCLI(name, email, nickname string) error {
	if err := addLocalIdentity(name, email, nickname); err != nil {
		return err
	}

	identity := Identity{Name: name, Email: email, Nickname: nickname}
	display := getIdentityDisplay(identity)
	fmt.Printf("Added and set local repository identity to %s\n", display)
	return nil
}

func showHelp() {
	fmt.Println(`GitID - Git Identity Manager

USAGE:
    gitid                           Launch interactive TUI
    gitid list                      List all identities
    gitid current                   Show current global git identity
    gitid switch <identifier>       Switch global identity by nickname, name, or email
    gitid use <identifier>          Alias for switch
    gitid add <name> <email> [nick] Add new identity with optional nickname
    gitid delete <identifier>       Delete identity
    gitid nickname <id> <nickname>  Set/update nickname for identity
    gitid repo current              Show current local repository identity
    gitid repo use <identifier>     Set local repository identity
    gitid repo add <name> <email> [nick] Add and set new local repository identity
    gitid completion <shell>        Install shell completion (bash/zsh/fish)
    gitid completion <shell> -r     Remove shell completion
    gitid help                      Show this help

EXAMPLES:
    gitid list
    gitid current
    gitid switch work
    gitid add "John Doe" "john@company.com" work
    gitid nickname john@company.com work
    gitid delete work
    gitid repo current
    gitid repo use work
    gitid repo add "Jane Smith" "jane@company.com" work-jane
    gitid completion bash
    gitid completion zsh -r`)
}
