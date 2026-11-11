package tui

// ViewState tells which screen the model currently shows.
type ViewState int

// View states.
const (
	ViewTable ViewState = iota
	ViewFilterInput
	ViewHelp
	// ViewFieldEdit is a row of the settings form being typed into.
	ViewFieldEdit
)
