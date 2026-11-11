package report

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Byroncl/exposedwatch/internal/app"
)

// writeDriftSection prints the baseline used, the new ports with their
// severity, the closed entries (informational) and the notes. With no drift it
// is a single line.
func (renderer *TableRenderer) writeDriftSection(builder *strings.Builder, driftReport app.DriftReport) {
	title := fmt.Sprintf(driftSectionTitleFormat, driftReport.BaselineFile, driftReport.CreatedAt.UTC().Format(driftCreatedAtLayout))
	builder.WriteString(newline + paint(renderer.colorMode, ansiBold, title) + newline)
	if len(driftReport.New) == 0 && len(driftReport.Closed) == 0 {
		builder.WriteString(driftNoDriftText)
	}
	if len(driftReport.New) > 0 {
		fmt.Fprintf(builder, driftNewTitleFormat, len(driftReport.New))
		renderer.writeNewPorts(builder, driftReport.New)
	}
	if len(driftReport.Closed) > 0 {
		fmt.Fprintf(builder, driftClosedTitleFormat, len(driftReport.Closed))
		writeClosedEntries(builder, driftReport)
	}
	for _, note := range driftReport.Notes {
		fmt.Fprintf(builder, driftNoteFormat, note)
	}
}

func (renderer *TableRenderer) writeNewPorts(builder *strings.Builder, assessments []app.NewPortAssessment) {
	headers := []string{headerSeverity, headerPort, headerProtocol, headerBind, headerSource, headerProcess, headerTitle}
	rows := make([][]string, 0, len(assessments))
	for _, assessment := range assessments {
		rows = append(rows, []string{
			assessment.Severity.String(),
			strconv.Itoa(assessment.Port.Number),
			string(assessment.Port.Protocol),
			orPlaceholder(assessment.Port.BindAddress),
			string(assessment.Port.Source),
			orPlaceholder(processOrContainer(assessment.Port)),
			orPlaceholder(assessment.Title),
		})
	}
	widths := columnWidths(headers, rows)
	builder.WriteString(driftRowIndent)
	writeRow(builder, headers, widths, func(_ int, cell string) string { return paint(renderer.colorMode, ansiBold, cell) })
	for rowIndex, row := range rows {
		severity := assessments[rowIndex].Severity
		builder.WriteString(driftRowIndent)
		writeRow(builder, row, widths, func(column int, cell string) string {
			if column == 0 {
				return paint(renderer.colorMode, severityColor(severity), cell)
			}
			return cell
		})
	}
}

func writeClosedEntries(builder *strings.Builder, driftReport app.DriftReport) {
	headers := []string{headerPort, headerProtocol, headerBind, headerSource}
	rows := make([][]string, 0, len(driftReport.Closed))
	for _, entry := range driftReport.Closed {
		rows = append(rows, []string{strconv.Itoa(entry.Port), string(entry.Protocol), entry.BindAddress, string(entry.Source)})
	}
	widths := columnWidths(headers, rows)
	plain := func(_ int, cell string) string { return cell }
	builder.WriteString(driftRowIndent)
	writeRow(builder, headers, widths, plain)
	for _, row := range rows {
		builder.WriteString(driftRowIndent)
		writeRow(builder, row, widths, plain)
	}
}

func columnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for column, header := range headers {
		widths[column] = utf8.RuneCountInString(header)
	}
	for _, row := range rows {
		for column, cell := range row {
			widths[column] = max(widths[column], utf8.RuneCountInString(cell))
		}
	}
	return widths
}
