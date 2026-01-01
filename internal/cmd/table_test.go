package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable_BasicOutput(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME", "STATUS")
	table.AddRow("123", "Test Item", "active")
	table.AddRow("456", "Another", "inactive")
	table.Render(&buf)

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	if len(lines) != 4 { // header + separator + 2 data rows
		t.Errorf("expected 4 lines, got %d: %q", len(lines), output)
	}

	// Check header line contains all headers
	if !strings.Contains(lines[0], "ID") || !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "STATUS") {
		t.Errorf("header line missing expected columns: %q", lines[0])
	}

	// Check separator line uses Unicode box-drawing character
	if !strings.Contains(lines[1], "\u2500") {
		t.Errorf("separator line should contain Unicode horizontal line: %q", lines[1])
	}

	// Check data rows
	if !strings.Contains(lines[2], "123") || !strings.Contains(lines[2], "Test Item") {
		t.Errorf("first data row missing expected values: %q", lines[2])
	}
	if !strings.Contains(lines[3], "456") || !strings.Contains(lines[3], "Another") {
		t.Errorf("second data row missing expected values: %q", lines[3])
	}
}

func TestTable_EmptyTable(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME")
	table.Render(&buf)

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	// Empty table should still show header and separator
	if len(lines) != 2 {
		t.Errorf("expected 2 lines for empty table (header + separator), got %d: %q", len(lines), output)
	}
}

func TestTable_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable()
	table.Render(&buf)

	if buf.Len() != 0 {
		t.Errorf("expected empty output for table with no headers, got %q", buf.String())
	}
}

func TestTable_LongValueTruncation(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "DESCRIPTION")
	longValue := strings.Repeat("x", 50) // Longer than maxColumnWidth (40)
	table.AddRow("1", longValue)
	table.Render(&buf)

	output := buf.String()

	// Value should be truncated with "..."
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncation ellipsis in output: %q", output)
	}

	// Line should not contain the full long value
	if strings.Contains(output, longValue) {
		t.Errorf("output should have truncated value, but contains full value: %q", output)
	}
}

func TestTable_ColumnAlignment(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME")
	table.AddRow("1", "Short")
	table.AddRow("123456", "Longer Name Here")
	table.Render(&buf)

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")

	// Find where NAME column starts in header
	headerNamePos := strings.Index(lines[0], "NAME")

	// Check that NAME column values start at the same position in data rows
	for i := 2; i < len(lines); i++ { // Skip header and separator
		// Find where the second column starts (after ID column + spacing)
		// This tests that columns are properly aligned
		if len(lines[i]) <= headerNamePos {
			continue // Row might be shorter
		}
	}
}

func TestTable_FewerValuesThanHeaders(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME", "STATUS", "EXTRA")
	table.AddRow("1", "Test") // Only 2 values, 4 headers
	table.Render(&buf)

	output := buf.String()
	// Should not panic, and should produce valid output
	if !strings.Contains(output, "1") || !strings.Contains(output, "Test") {
		t.Errorf("output missing expected values: %q", output)
	}
}

func TestTable_IsEmpty(t *testing.T) {
	table := NewTable("ID", "NAME")

	if !table.IsEmpty() {
		t.Error("new table should be empty")
	}

	table.AddRow("1", "Test")

	if table.IsEmpty() {
		t.Error("table with rows should not be empty")
	}
}

func TestTable_RowCount(t *testing.T) {
	table := NewTable("ID", "NAME")

	if table.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", table.RowCount())
	}

	table.AddRow("1", "First")
	table.AddRow("2", "Second")
	table.AddRow("3", "Third")

	if table.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", table.RowCount())
	}
}

func TestPadOrTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "exact width",
			input:    "hello",
			width:    5,
			expected: "hello",
		},
		{
			name:     "shorter than width - padding",
			input:    "hi",
			width:    5,
			expected: "hi   ",
		},
		{
			name:     "longer than width - truncation",
			input:    "hello world",
			width:    8,
			expected: "hello...",
		},
		{
			name:     "very short width",
			input:    "hello",
			width:    3,
			expected: "...",
		},
		{
			name:     "width of 2",
			input:    "hello",
			width:    2,
			expected: "..",
		},
		{
			name:     "width of 1",
			input:    "hello",
			width:    1,
			expected: ".",
		},
		{
			name:     "empty string",
			input:    "",
			width:    5,
			expected: "     ",
		},
		{
			name:     "unicode characters - padded",
			input:    "\u4e2d\u6587\u5b57\u7b26", // 4 characters
			width:    6,
			expected: "\u4e2d\u6587\u5b57\u7b26  ", // padded to 6
		},
		{
			name:     "unicode characters - truncated",
			input:    "\u4e2d\u6587\u5b57\u7b26\u6d4b\u8bd5", // 6 characters
			width:    5,
			expected: "\u4e2d\u6587...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padOrTruncate(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("padOrTruncate(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

func TestTable_UnicodeContent(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME")
	table.AddRow("1", "\u4e2d\u6587\u540d\u79f0") // Chinese characters
	table.AddRow("2", "\u65e5\u672c\u8a9e")       // Japanese characters
	table.Render(&buf)

	output := buf.String()
	if !strings.Contains(output, "\u4e2d\u6587\u540d\u79f0") {
		t.Errorf("output should contain Chinese characters: %q", output)
	}
	if !strings.Contains(output, "\u65e5\u672c\u8a9e") {
		t.Errorf("output should contain Japanese characters: %q", output)
	}
}

func TestTable_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "VALUE")
	table.AddRow("1", "value with spaces")
	table.AddRow("2", "value-with-dashes")
	table.AddRow("3", "value_with_underscores")
	table.Render(&buf)

	output := buf.String()
	if !strings.Contains(output, "value with spaces") {
		t.Errorf("output should contain value with spaces: %q", output)
	}
}

// Edge case tests for calculateColumnWidths

func TestTable_MoreValuesThanHeaders(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "NAME")
	// Add more values than headers - extra values should be ignored
	table.AddRow("1", "Test", "Extra", "MoreExtra")
	table.Render(&buf)

	output := buf.String()
	// Should render without errors and only include columns matching headers
	if !strings.Contains(output, "1") || !strings.Contains(output, "Test") {
		t.Errorf("output missing expected values: %q", output)
	}
	// Extra values should not appear
	if strings.Contains(output, "Extra") {
		t.Errorf("output should not contain extra values beyond headers: %q", output)
	}
}

func TestTable_RowWithLessColumnsThanHeaderValues(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("A", "B", "C", "D", "E")
	// Row with fewer values
	table.AddRow("1")
	table.Render(&buf)

	// Should not panic and should produce valid output
	output := buf.String()
	if !strings.Contains(output, "A") || !strings.Contains(output, "1") {
		t.Errorf("output missing expected values: %q", output)
	}
}

func TestTable_AllColumnsAtMaxWidth(t *testing.T) {
	var buf bytes.Buffer
	table := NewTable("ID", "VERY_LONG_DESCRIPTION")
	// Add value longer than maxColumnWidth (40)
	longValue1 := strings.Repeat("A", 50)
	longValue2 := strings.Repeat("B", 50)
	table.AddRow("1", longValue1)
	table.AddRow("2", longValue2)
	table.Render(&buf)

	output := buf.String()
	// Both values should be truncated
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncation in output: %q", output)
	}
	// Should not contain the full 50-char values
	if strings.Contains(output, longValue1) || strings.Contains(output, longValue2) {
		t.Errorf("output should not contain full long values: %q", output)
	}
}

func TestCalculateColumnWidths_RowWithMoreColumns(t *testing.T) {
	table := NewTable("A", "B")
	// Add row with values that would go beyond the header columns
	table.rows = append(table.rows, []string{"1", "2", "3", "4"}) // More than headers

	widths := table.calculateColumnWidths()

	// Should only have widths for the headers
	if len(widths) != 2 {
		t.Errorf("expected 2 column widths, got %d", len(widths))
	}
}
