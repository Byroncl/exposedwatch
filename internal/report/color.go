package report

import (
	"io"
	"os"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// DetectColorMode enables color only for terminals when NO_COLOR is unset.
// noColorValue is the value of the NO_COLOR environment variable.
func DetectColorMode(output io.Writer, noColorValue string) ColorMode {
	if noColorValue != noColorUnsetValue {
		return ColorDisabled
	}
	outputFile, isFile := output.(*os.File)
	if !isFile {
		return ColorDisabled
	}
	fileInfo, err := outputFile.Stat()
	if err != nil || fileInfo.Mode()&os.ModeCharDevice == 0 {
		return ColorDisabled
	}
	return ColorEnabled
}

func severityColor(severity domain.Severity) AnsiCode {
	switch severity {
	case domain.Critical:
		return ansiBoldRedOnDefault
	case domain.High:
		return ansiRed
	case domain.Medium:
		return ansiYellow
	case domain.Low:
		return ansiCyan
	default:
		return ansiGray
	}
}

func paint(mode ColorMode, code AnsiCode, text string) string {
	if mode == ColorDisabled {
		return text
	}
	return string(code) + text + string(ansiReset)
}
