package ports

import "github.com/Byroncl/exposedwatch/internal/domain"

// BaselinePath is where a baseline file lives.
type BaselinePath string

// BaselineLoad is a trusted, validated baseline plus non-fatal warnings.
type BaselineLoad struct {
	Baseline domain.Baseline
	Warnings []string
}

// BaselineStore reads and writes the baseline file. Load refuses a file that
// may have been tampered with, because the baseline decides which ports count
// as accepted.
type BaselineStore interface {
	Load(path BaselinePath) (BaselineLoad, error)
	Save(path BaselinePath, baseline domain.Baseline) error
}
