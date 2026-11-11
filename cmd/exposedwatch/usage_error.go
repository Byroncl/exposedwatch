package main

// usageError is a flag validation failure that main must print itself
// (the flag package already prints its own parse errors).
type usageError struct {
	message string
}

// Error implements the error interface.
func (usage usageError) Error() string {
	return usage.message
}
