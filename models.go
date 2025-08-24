package main

import "github.com/charmbracelet/bubbles/textinput"

type Identity struct {
	Name     string
	Email    string
	Nickname string
}

type Model struct {
	identities       []Identity
	cursor           int
	showConfirmation bool
	confirmChoices   []string
	confirmCursor    int
}

type InputModel struct {
	textInput   textinput.Model
	value       string
	interrupted bool
}

type CompletionPromptModel struct {
	shell         string
	choices       []string
	cursor        int
	shouldInstall bool
	finished      bool
}
