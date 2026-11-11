package tui

import (
	"fmt"
	"time"
)

// formatConnectionAge renders how long a connection has been seen as its
// largest whole unit: 12s, 4m, 1h, 2d. Negative ages count as zero.
func formatConnectionAge(age time.Duration) string {
	switch {
	case age < time.Minute:
		return fmt.Sprintf(ageFormatSeconds, int(max(age, 0)/time.Second))
	case age < time.Hour:
		return fmt.Sprintf(ageFormatMinutes, int(age/time.Minute))
	case age < hoursPerDay*time.Hour:
		return fmt.Sprintf(ageFormatHours, int(age/time.Hour))
	default:
		return fmt.Sprintf(ageFormatDays, int(age/(hoursPerDay*time.Hour)))
	}
}
