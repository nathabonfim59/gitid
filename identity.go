package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func encodeEmail(email string) string {
	return strings.ReplaceAll(strings.ReplaceAll(email, "@", "_at_"), ".", "_dot_")
}

func setNickname(email, nickname string) error {
	section := encodeEmail(email)
	nicknameCmd := fmt.Sprintf("identity.%s.nickname", section)
	return exec.Command("git", "config", "--global", nicknameCmd, nickname).Run()
}

func getNickname(email string) string {
	section := encodeEmail(email)
	nicknameCmd := fmt.Sprintf("identity.%s.nickname", section)
	out, err := exec.Command("git", "config", "--global", nicknameCmd).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func hasNickname(email string) bool {
	return getNickname(email) != ""
}

func getIdentityDisplay(identity Identity) string {
	if identity.Nickname != "" {
		return fmt.Sprintf("%s (%s <%s>)", identity.Nickname, identity.Name, identity.Email)
	}
	return fmt.Sprintf("%s <%s>", identity.Name, identity.Email)
}

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

func findIdentityByIdentifier(identifier string) (Identity, bool) {
	identities := getAllIdentities()

	for _, identity := range identities {
		if identity.Nickname == identifier {
			return identity, true
		}
	}

	for _, identity := range identities {
		if identity.Email == identifier {
			return identity, true
		}
	}

	for _, identity := range identities {
		if identity.Name == identifier {
			return identity, true
		}
	}

	for _, identity := range identities {
		if strings.Contains(identity.Email, identifier) {
			return identity, true
		}
	}

	for _, identity := range identities {
		if strings.Contains(identity.Name, identifier) {
			return identity, true
		}
	}

	return Identity{}, false
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

	if nickname != "" {
		if err := setNickname(email, nickname); err != nil {
			return fmt.Errorf("error setting nickname: %w", err)
		}
	}

	return nil
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

func switchIdentityByIdentifier(identifier string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	switchIdentity(identity.Name, identity.Email)
	return nil
}

func updateIdentity(oldEmail, newName, newEmail, newNickname string) error {
	if oldEmail != newEmail {
		if err := deleteIdentity(oldEmail); err != nil {
			return fmt.Errorf("error removing old identity: %w", err)
		}
	}

	if err := addIdentity(newName, newEmail, newNickname); err != nil {
		return fmt.Errorf("error adding updated identity: %w", err)
	}

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
	exec.Command("git", "config", "--global", "--unset", nicknameCmd).Run()

	return nil
}

func setLocalIdentity(name, email string) error {
	if err := exec.Command("git", "config", "user.name", name).Run(); err != nil {
		return fmt.Errorf("error setting local user name: %w", err)
	}
	if err := exec.Command("git", "config", "user.email", email).Run(); err != nil {
		return fmt.Errorf("error setting local user email: %w", err)
	}
	return nil
}

func setLocalIdentityByIdentifier(identifier string) error {
	identity, found := findIdentityByIdentifier(identifier)
	if !found {
		return fmt.Errorf("identity not found: %s", identifier)
	}

	return setLocalIdentity(identity.Name, identity.Email)
}

func addLocalIdentity(name, email, nickname string) error {
	// First add to global identities if not exists
	section := encodeEmail(email)
	nameCmd := fmt.Sprintf("identity.%s.name", section)

	// Check if identity already exists globally
	_, err := exec.Command("git", "config", "--global", nameCmd).Output()
	if err != nil {
		// Identity doesn't exist globally, add it
		if err := addIdentity(name, email, nickname); err != nil {
			return fmt.Errorf("error adding identity globally: %w", err)
		}
	}

	// Set as local identity
	return setLocalIdentity(name, email)
}

func getCurrentLocalIdentity() (string, string, error) {
	nameOut, err := exec.Command("git", "config", "--local", "user.name").Output()
	if err != nil {
		return "", "", fmt.Errorf("no local git identity configured")
	}
	emailOut, err := exec.Command("git", "config", "--local", "user.email").Output()
	if err != nil {
		return "", "", fmt.Errorf("no local git identity configured")
	}

	name := strings.TrimSpace(string(nameOut))
	email := strings.TrimSpace(string(emailOut))
	return name, email, nil
}

func isInGitRepository() bool {
	err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run()
	return err == nil
}

func hasLocalIdentity() bool {
	if !isInGitRepository() {
		return false
	}

	// Check if local config actually has user.name and user.email set
	// (not just falling back to global)
	nameOut, err := exec.Command("git", "config", "--local", "user.name").Output()
	if err != nil {
		return false
	}
	emailOut, err := exec.Command("git", "config", "--local", "user.email").Output()
	if err != nil {
		return false
	}

	name := strings.TrimSpace(string(nameOut))
	email := strings.TrimSpace(string(emailOut))

	return name != "" && email != ""
}
