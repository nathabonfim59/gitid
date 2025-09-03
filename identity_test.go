package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestEncodeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test@example.com", "test_at_example_dot_com"},
		{"user.name@domain.co.uk", "user_dot_name_at_domain_dot_co_dot_uk"},
		{"simple@test.org", "simple_at_test_dot_org"},
		{"no-dots@nodots", "no-dots_at_nodots"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := encodeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("encodeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetIdentityDisplay(t *testing.T) {
	tests := []struct {
		name     string
		identity Identity
		expected string
	}{
		{
			name: "with nickname",
			identity: Identity{
				Name:     "John Doe",
				Email:    "john@example.com",
				Nickname: "johnny",
			},
			expected: "johnny (John Doe <john@example.com>)",
		},
		{
			name: "without nickname",
			identity: Identity{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			expected: "Jane Smith <jane@example.com>",
		},
		{
			name: "empty nickname",
			identity: Identity{
				Name:     "Bob Wilson",
				Email:    "bob@example.com",
				Nickname: "",
			},
			expected: "Bob Wilson <bob@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIdentityDisplay(tt.identity)
			if result != tt.expected {
				t.Errorf("getIdentityDisplay() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func setupTestGitConfig(t *testing.T) func() {
	originalHome := os.Getenv("HOME")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	exec.Command("git", "config", "--global", "init.defaultBranch", "main").Run()

	return func() {
		os.Setenv("HOME", originalHome)
	}
}

func TestSetAndGetNickname(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	email := "test@example.com"
	nickname := "testnick"

	err := setNickname(email, nickname)
	if err != nil {
		t.Fatalf("setNickname failed: %v", err)
	}

	result := getNickname(email)
	if result != nickname {
		t.Errorf("getNickname() = %q, want %q", result, nickname)
	}

	if !hasNickname(email) {
		t.Error("hasNickname() should return true after setting nickname")
	}
}

func TestGetNicknameNonExistent(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	email := "nonexistent@example.com"
	result := getNickname(email)
	if result != "" {
		t.Errorf("getNickname() for non-existent email = %q, want empty string", result)
	}

	if hasNickname(email) {
		t.Error("hasNickname() should return false for non-existent email")
	}
}

func TestAddIdentity(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	tests := []struct {
		name     string
		userName string
		email    string
		nickname string
	}{
		{"with nickname", "John Doe", "john@example.com", "johnny"},
		{"without nickname", "Jane Smith", "jane@example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addIdentity(tt.userName, tt.email, tt.nickname)
			if err != nil {
				t.Fatalf("addIdentity failed: %v", err)
			}

			section := encodeEmail(tt.email)
			nameCmd := fmt.Sprintf("identity.%s.name", section)
			emailCmd := fmt.Sprintf("identity.%s.email", section)

			nameOut, err := exec.Command("git", "config", "--global", nameCmd).Output()
			if err != nil {
				t.Fatalf("Failed to get name config: %v", err)
			}
			if strings.TrimSpace(string(nameOut)) != tt.userName {
				t.Errorf("Name not set correctly: got %q, want %q", strings.TrimSpace(string(nameOut)), tt.userName)
			}

			emailOut, err := exec.Command("git", "config", "--global", emailCmd).Output()
			if err != nil {
				t.Fatalf("Failed to get email config: %v", err)
			}
			if strings.TrimSpace(string(emailOut)) != tt.email {
				t.Errorf("Email not set correctly: got %q, want %q", strings.TrimSpace(string(emailOut)), tt.email)
			}

			if tt.nickname != "" {
				nickname := getNickname(tt.email)
				if nickname != tt.nickname {
					t.Errorf("Nickname not set correctly: got %q, want %q", nickname, tt.nickname)
				}
			}
		})
	}
}

func TestSwitchIdentity(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	name := "Test User"
	email := "test@example.com"

	switchIdentity(name, email)

	nameOut, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		t.Fatalf("Failed to get user.name: %v", err)
	}
	if strings.TrimSpace(string(nameOut)) != name {
		t.Errorf("user.name not set correctly: got %q, want %q", strings.TrimSpace(string(nameOut)), name)
	}

	emailOut, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		t.Fatalf("Failed to get user.email: %v", err)
	}
	if strings.TrimSpace(string(emailOut)) != email {
		t.Errorf("user.email not set correctly: got %q, want %q", strings.TrimSpace(string(emailOut)), email)
	}
}

func TestFindIdentityByIdentifier(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	addIdentity("John Doe", "john@example.com", "johnny")
	addIdentity("Jane Smith", "jane@example.com", "")
	addIdentity("Bob Wilson", "bob@company.org", "bobby")

	tests := []struct {
		name       string
		identifier string
		shouldFind bool
		expected   Identity
	}{
		{
			name:       "find by nickname",
			identifier: "johnny",
			shouldFind: true,
			expected:   Identity{Name: "John Doe", Email: "john@example.com", Nickname: "johnny"},
		},
		{
			name:       "find by email exact",
			identifier: "jane@example.com",
			shouldFind: true,
			expected:   Identity{Name: "Jane Smith", Email: "jane@example.com", Nickname: ""},
		},
		{
			name:       "find by name exact",
			identifier: "Bob Wilson",
			shouldFind: true,
			expected:   Identity{Name: "Bob Wilson", Email: "bob@company.org", Nickname: "bobby"},
		},
		{
			name:       "find by partial email",
			identifier: "company",
			shouldFind: true,
			expected:   Identity{Name: "Bob Wilson", Email: "bob@company.org", Nickname: "bobby"},
		},
		{
			name:       "find by partial name",
			identifier: "Jane",
			shouldFind: true,
			expected:   Identity{Name: "Jane Smith", Email: "jane@example.com", Nickname: ""},
		},
		{
			name:       "not found",
			identifier: "nonexistent",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, found := findIdentityByIdentifier(tt.identifier)
			if found != tt.shouldFind {
				t.Errorf("findIdentityByIdentifier(%q) found = %v, want %v", tt.identifier, found, tt.shouldFind)
			}
			if tt.shouldFind {
				if identity.Name != tt.expected.Name || identity.Email != tt.expected.Email || identity.Nickname != tt.expected.Nickname {
					t.Errorf("findIdentityByIdentifier(%q) = %+v, want %+v", tt.identifier, identity, tt.expected)
				}
			}
		})
	}
}

func TestSwitchIdentityByIdentifier(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	addIdentity("John Doe", "john@example.com", "johnny")

	err := switchIdentityByIdentifier("johnny")
	if err != nil {
		t.Fatalf("switchIdentityByIdentifier failed: %v", err)
	}

	nameOut, _ := exec.Command("git", "config", "--global", "user.name").Output()
	emailOut, _ := exec.Command("git", "config", "--global", "user.email").Output()

	if strings.TrimSpace(string(nameOut)) != "John Doe" {
		t.Errorf("user.name not switched correctly")
	}
	if strings.TrimSpace(string(emailOut)) != "john@example.com" {
		t.Errorf("user.email not switched correctly")
	}

	err = switchIdentityByIdentifier("nonexistent")
	if err == nil {
		t.Error("switchIdentityByIdentifier should fail for non-existent identifier")
	}
}

func TestDeleteIdentity(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	email := "test@example.com"
	addIdentity("Test User", email, "testnick")

	err := deleteIdentity(email)
	if err != nil {
		t.Fatalf("deleteIdentity failed: %v", err)
	}

	section := encodeEmail(email)
	nameCmd := fmt.Sprintf("identity.%s.name", section)
	emailCmd := fmt.Sprintf("identity.%s.email", section)

	_, err = exec.Command("git", "config", "--global", nameCmd).Output()
	if err == nil {
		t.Error("Name config should be removed after deleteIdentity")
	}

	_, err = exec.Command("git", "config", "--global", emailCmd).Output()
	if err == nil {
		t.Error("Email config should be removed after deleteIdentity")
	}

	if hasNickname(email) {
		t.Error("Nickname should be removed after deleteIdentity")
	}
}

func TestGetAllIdentities(t *testing.T) {
	cleanup := setupTestGitConfig(t)
	defer cleanup()

	addIdentity("John Doe", "john@example.com", "johnny")
	addIdentity("Jane Smith", "jane@example.com", "")
	addIdentity("Bob Wilson", "bob@company.org", "bobby")

	identities := getAllIdentities()

	if len(identities) != 3 {
		t.Errorf("getAllIdentities() returned %d identities, want 3", len(identities))
	}

	found := make(map[string]bool)
	for _, identity := range identities {
		key := identity.Email
		found[key] = true

		switch key {
		case "john@example.com":
			if identity.Name != "John Doe" || identity.Nickname != "johnny" {
				t.Errorf("John's identity incorrect: %+v", identity)
			}
		case "jane@example.com":
			if identity.Name != "Jane Smith" || identity.Nickname != "" {
				t.Errorf("Jane's identity incorrect: %+v", identity)
			}
		case "bob@company.org":
			if identity.Name != "Bob Wilson" || identity.Nickname != "bobby" {
				t.Errorf("Bob's identity incorrect: %+v", identity)
			}
		}
	}

	expectedEmails := []string{"john@example.com", "jane@example.com", "bob@company.org"}
	for _, email := range expectedEmails {
		if !found[email] {
			t.Errorf("Expected identity with email %s not found", email)
		}
	}
}

func setupTestGitRepo(t *testing.T) (string, func()) {
	cleanup := setupTestGitConfig(t)
	tempDir := t.TempDir()

	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	return tempDir, func() {
		os.Chdir(originalDir)
		cleanup()
	}
}

func TestSetLocalIdentity(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	name := "Local Test User"
	email := "local@example.com"

	err := setLocalIdentity(name, email)
	if err != nil {
		t.Fatalf("setLocalIdentity failed: %v", err)
	}

	nameOut, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		t.Fatalf("Failed to get local user.name: %v", err)
	}
	if strings.TrimSpace(string(nameOut)) != name {
		t.Errorf("Local user.name not set correctly: got %q, want %q", strings.TrimSpace(string(nameOut)), name)
	}

	emailOut, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		t.Fatalf("Failed to get local user.email: %v", err)
	}
	if strings.TrimSpace(string(emailOut)) != email {
		t.Errorf("Local user.email not set correctly: got %q, want %q", strings.TrimSpace(string(emailOut)), email)
	}
}

func TestSetLocalIdentityByIdentifier(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	// Add a global identity first
	addIdentity("Test User", "test@example.com", "testuser")

	err := setLocalIdentityByIdentifier("testuser")
	if err != nil {
		t.Fatalf("setLocalIdentityByIdentifier failed: %v", err)
	}

	name, email, err := getCurrentLocalIdentity()
	if err != nil {
		t.Fatalf("getCurrentLocalIdentity failed: %v", err)
	}

	if name != "Test User" {
		t.Errorf("Local user.name not set correctly: got %q, want %q", name, "Test User")
	}
	if email != "test@example.com" {
		t.Errorf("Local user.email not set correctly: got %q, want %q", email, "test@example.com")
	}

	// Test with non-existent identifier
	err = setLocalIdentityByIdentifier("nonexistent")
	if err == nil {
		t.Error("setLocalIdentityByIdentifier should fail for non-existent identifier")
	}
}

func TestAddLocalIdentity(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	name := "New Local User"
	email := "newlocal@example.com"
	nickname := "newlocal"

	err := addLocalIdentity(name, email, nickname)
	if err != nil {
		t.Fatalf("addLocalIdentity failed: %v", err)
	}

	// Check that identity was added globally
	identity, found := findIdentityByIdentifier(nickname)
	if !found {
		t.Error("Identity should be added globally")
	}
	if identity.Name != name || identity.Email != email || identity.Nickname != nickname {
		t.Errorf("Global identity incorrect: got %+v, want {%s %s %s}", identity, name, email, nickname)
	}

	// Check that identity is set locally
	localName, localEmail, err := getCurrentLocalIdentity()
	if err != nil {
		t.Fatalf("getCurrentLocalIdentity failed: %v", err)
	}
	if localName != name || localEmail != email {
		t.Errorf("Local identity not set correctly: got %q <%s>, want %q <%s>", localName, localEmail, name, email)
	}
}

func TestAddLocalIdentityExisting(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	// Add identity globally first
	name := "Existing User"
	email := "existing@example.com"
	nickname := "existing"
	addIdentity(name, email, nickname)

	// Add as local identity
	err := addLocalIdentity(name, email, nickname)
	if err != nil {
		t.Fatalf("addLocalIdentity failed for existing identity: %v", err)
	}

	// Check that identity is set locally
	localName, localEmail, err := getCurrentLocalIdentity()
	if err != nil {
		t.Fatalf("getCurrentLocalIdentity failed: %v", err)
	}
	if localName != name || localEmail != email {
		t.Errorf("Local identity not set correctly: got %q <%s>, want %q <%s>", localName, localEmail, name, email)
	}
}

func TestGetCurrentLocalIdentity(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	// Test when no local identity is set
	_, _, err := getCurrentLocalIdentity()
	if err == nil {
		t.Error("getCurrentLocalIdentity should fail when no local identity is set")
	}

	// Set local identity and test
	name := "Local User"
	email := "local@example.com"
	setLocalIdentity(name, email)

	gotName, gotEmail, err := getCurrentLocalIdentity()
	if err != nil {
		t.Fatalf("getCurrentLocalIdentity failed: %v", err)
	}

	if gotName != name {
		t.Errorf("getCurrentLocalIdentity name: got %q, want %q", gotName, name)
	}
	if gotEmail != email {
		t.Errorf("getCurrentLocalIdentity email: got %q, want %q", gotEmail, email)
	}
}

func TestLocalIdentityIsolation(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	// Set global identity
	globalName := "Global User"
	globalEmail := "global@example.com"
	switchIdentity(globalName, globalEmail)

	// Set local identity
	localName := "Local User"
	localEmail := "local@example.com"
	setLocalIdentity(localName, localEmail)

	// Verify local identity
	gotLocalName, gotLocalEmail, err := getCurrentLocalIdentity()
	if err != nil {
		t.Fatalf("getCurrentLocalIdentity failed: %v", err)
	}
	if gotLocalName != localName || gotLocalEmail != localEmail {
		t.Errorf("Local identity incorrect: got %q <%s>, want %q <%s>", gotLocalName, gotLocalEmail, localName, localEmail)
	}

	// Verify global identity is unchanged
	nameOut, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		t.Fatalf("Failed to get global user.name: %v", err)
	}
	if strings.TrimSpace(string(nameOut)) != globalName {
		t.Errorf("Global user.name changed: got %q, want %q", strings.TrimSpace(string(nameOut)), globalName)
	}

	emailOut, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		t.Fatalf("Failed to get global user.email: %v", err)
	}
	if strings.TrimSpace(string(emailOut)) != globalEmail {
		t.Errorf("Global user.email changed: got %q, want %q", strings.TrimSpace(string(emailOut)), globalEmail)
	}
}
