package tui

// styleRenderer is what the views need from a Lip Gloss style.
type styleRenderer interface {
	Render(...string) string
}
