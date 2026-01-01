package cmd

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Table provides a simple table formatter for list output.
// It renders aligned columns with headers and auto-sizes columns based on content.
type Table struct {
	headers []string
	rows    [][]string
	maxCols int
}

// NewTable creates a new table with the given column headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		maxCols: len(headers),
	}
}

// AddRow adds a row of values to the table.
// If fewer values are provided than headers, remaining columns are left empty.
// If more values are provided than headers, they are ignored.
func (t *Table) AddRow(values ...string) {
	row := make([]string, t.maxCols)
	for i := 0; i < t.maxCols && i < len(values); i++ {
		row[i] = values[i]
	}
	t.rows = append(t.rows, row)
}

// Render writes the formatted table to the given writer.
// Columns are auto-sized based on content, with a maximum width to prevent
// overly wide tables. Long values are truncated with "...".
func (t *Table) Render(w io.Writer) {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths
	widths := t.calculateColumnWidths()

	// Print header row
	t.printRow(w, t.headers, widths)

	// Print separator line
	t.printSeparator(w, widths)

	// Print data rows
	for _, row := range t.rows {
		t.printRow(w, row, widths)
	}
}

// calculateColumnWidths determines the width for each column.
// Each column width is the maximum of the header width and all row values,
// capped at a maximum width.
func (t *Table) calculateColumnWidths() []int {
	const maxColumnWidth = 40

	widths := make([]int, len(t.headers))

	// Start with header widths
	for i, h := range t.headers {
		widths[i] = utf8.RuneCountInString(h)
	}

	// Check all rows for wider values
	for _, row := range t.rows {
		for i, val := range row {
			if i >= len(widths) {
				break
			}
			valWidth := utf8.RuneCountInString(val)
			if valWidth > widths[i] {
				widths[i] = valWidth
			}
		}
	}

	// Cap at maximum width
	for i := range widths {
		if widths[i] > maxColumnWidth {
			widths[i] = maxColumnWidth
		}
	}

	return widths
}

// printRow writes a single row of values with proper column alignment.
func (t *Table) printRow(w io.Writer, values []string, widths []int) {
	parts := make([]string, len(values))
	for i, val := range values {
		width := widths[i]
		parts[i] = padOrTruncate(val, width)
	}
	_, _ = fmt.Fprintln(w, strings.Join(parts, "  "))
}

// printSeparator writes a separator line using Unicode box-drawing dashes.
func (t *Table) printSeparator(w io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("\u2500", width) // Unicode box-drawing horizontal line
	}
	_, _ = fmt.Fprintln(w, strings.Join(parts, "  "))
}

// padOrTruncate ensures a string fits exactly within the given width.
// If the string is too long, it is truncated and "..." is appended.
// If the string is too short, it is padded with spaces.
func padOrTruncate(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)

	if runeCount <= width {
		// Pad with spaces
		return s + strings.Repeat(" ", width-runeCount)
	}

	// Truncate with ellipsis
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	// Take first (width-3) runes and add "..."
	runes := []rune(s)
	return string(runes[:width-3]) + "..."
}

// IsEmpty returns true if the table has no data rows.
func (t *Table) IsEmpty() bool {
	return len(t.rows) == 0
}

// RowCount returns the number of data rows in the table.
func (t *Table) RowCount() int {
	return len(t.rows)
}
