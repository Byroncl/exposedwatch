package report

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Byroncl/exposedwatch/internal/app"
	"github.com/Byroncl/exposedwatch/internal/domain"
)

// TableRenderer prints a human readable table.
type TableRenderer struct {
	colorMode ColorMode
}

// NewTableRenderer returns a table renderer with the given color mode.
func NewTableRenderer(colorMode ColorMode) *TableRenderer {
	return &TableRenderer{colorMode: colorMode}
}

// Render writes the findings table, recommendations and failed sources.
func (renderer *TableRenderer) Render(output io.Writer, scanReport app.Report) error {
	var builder strings.Builder
	if len(scanReport.Findings) == 0 {
		builder.WriteString(noFindingsText + newline)
	} else {
		renderer.writeTable(&builder, scanReport.Findings)
		renderer.writeRecommendations(&builder, scanReport.Findings)
	}
	fmt.Fprintf(&builder, newline+summaryFormat, len(scanReport.Ports), len(scanReport.Findings))
	renderer.writeFailedSources(&builder, scanReport.SourceErrors)
	_, err := io.WriteString(output, builder.String())
	return err
}

func (renderer *TableRenderer) writeTable(builder *strings.Builder, findings []domain.Finding) {
	headers := []string{headerSeverity, headerPort, headerProtocol, headerBind, headerExposure, headerSource, headerProcess, headerTitle}
	rows := make([][]string, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, []string{
			finding.Severity.String(),
			strconv.Itoa(finding.Port.Number),
			string(finding.Port.Protocol),
			orPlaceholder(finding.Port.BindAddress),
			finding.Port.Exposure().String(),
			string(finding.Port.Source),
			orPlaceholder(processOrContainer(finding.Port)),
			finding.Title,
		})
	}
	widths := make([]int, len(headers))
	for column, header := range headers {
		widths[column] = utf8.RuneCountInString(header)
	}
	for _, row := range rows {
		for column, cell := range row {
			widths[column] = max(widths[column], utf8.RuneCountInString(cell))
		}
	}
	writeRow(builder, headers, widths, func(_ int, cell string) string { return paint(renderer.colorMode, ansiBold, cell) })
	for rowIndex, row := range rows {
		severity := findings[rowIndex].Severity
		writeRow(builder, row, widths, func(column int, cell string) string {
			if column == 0 {
				return paint(renderer.colorMode, severityColor(severity), cell)
			}
			return cell
		})
	}
}

// writeRow pads each plain cell before styling so ANSI codes do not skew widths.
func writeRow(builder *strings.Builder, cells []string, widths []int, style func(column int, cell string) string) {
	lastColumn := len(cells) - 1
	for column, cell := range cells {
		padded := cell
		if column != lastColumn {
			padded += strings.Repeat(paddingCharacter, widths[column]-utf8.RuneCountInString(cell))
		}
		builder.WriteString(style(column, padded))
		if column != lastColumn {
			builder.WriteString(columnSeparator)
		}
	}
	builder.WriteString(newline)
}

func (renderer *TableRenderer) writeRecommendations(builder *strings.Builder, findings []domain.Finding) {
	builder.WriteString(newline + paint(renderer.colorMode, ansiBold, recommendationsTitle) + newline)
	seenRules := make(map[string]bool, len(findings))
	for _, finding := range findings {
		if seenRules[finding.RuleID] {
			continue
		}
		seenRules[finding.RuleID] = true
		fmt.Fprintf(builder, recommendationFormat, finding.RuleID, finding.Recommendation)
	}
}

func (renderer *TableRenderer) writeFailedSources(builder *strings.Builder, sourceErrors []app.SourceError) {
	if len(sourceErrors) == 0 {
		return
	}
	builder.WriteString(newline + paint(renderer.colorMode, ansiYellow, failedSourcesTitle) + newline)
	for _, sourceError := range sourceErrors {
		fmt.Fprintf(builder, failedSourceFormat, sourceError.Error())
	}
}

func processOrContainer(port domain.Port) string {
	if port.Container != "" {
		return port.Container
	}
	return port.Process
}

func orPlaceholder(text string) string {
	if text == "" {
		return emptyCellText
	}
	return text
}
