package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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

	for _, identity := range identities {
		if identity.Nickname != "" {
			fmt.Printf("%-12s %s <%s>\n", identity.Nickname, identity.Name, identity.Email)
		} else {
			fmt.Printf("%-12s %s <%s>\n", "-", identity.Name, identity.Email)
		}
	}
	return nil
}

func getCurrentIdentityCLI() error {
	nameOut, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		return fmt.Errorf("no git identity configured")
	}
	emailOut, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		return fmt.Errorf("no git identity configured")
	}

	name := strings.TrimSpace(string(nameOut))
	email := strings.TrimSpace(string(emailOut))

	nickname := getNickname(email)
	if nickname != "" {
		fmt.Printf("%s (%s <%s>)\n", nickname, name, email)
	} else {
		fmt.Printf("%s <%s>\n", name, email)
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

func showHelp() {
	fmt.Println(`GitID - Git Identity Manager

USAGE:
    gitid                           Launch interactive TUI
    gitid list                      List all identities
    gitid current                   Show current git identity
    gitid switch <identifier>       Switch to identity by nickname, name, or email
    gitid use <identifier>          Alias for switch
    gitid add <name> <email> [nick] Add new identity with optional nickname
    gitid delete <identifier>       Delete identity
    gitid nickname <id> <nickname>  Set/update nickname for identity
    gitid help                      Show this help

EXAMPLES:
    gitid list
    gitid current
    gitid switch work
    gitid add "John Doe" "john@company.com" work
    gitid nickname john@company.com work
    gitid delete work`)
}
