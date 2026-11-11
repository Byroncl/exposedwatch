// Package report renders scan reports for humans and machines.
package report

import (
	"io"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// Renderer writes a report to an output stream.
type Renderer interface {
	Render(output io.Writer, scanReport app.Report) error
}
