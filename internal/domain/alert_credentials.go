package domain

import "fmt"

// BotToken is the secret of a chat bot. Printing it through fmt never
// reveals it; the only way to read it is Reveal, which makes every use
// greppable.
type BotToken string

// ChatID identifies the chat that receives alerts.
type ChatID string

// String hides the secret.
func (token BotToken) String() string {
	return redactedCredentialText
}

// GoString hides the secret in %#v output.
func (token BotToken) GoString() string {
	return redactedCredentialText
}

// Format hides the secret for every fmt verb.
func (token BotToken) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(redactedCredentialText))
}

// Reveal returns the secret. Call it only where the token is sent.
func (token BotToken) Reveal() string {
	return string(token)
}

// IsEmpty reports whether no token is set.
func (token BotToken) IsEmpty() bool {
	return token == ""
}
