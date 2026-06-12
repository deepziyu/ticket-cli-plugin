package ticket

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// PrintTicketsJSON writes the ticket list in compact or formatted JSON.
func PrintTicketsJSON(w io.Writer, tickets []TicketInfo) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tickets)
}

// PrintTicketsTable writes the ticket list in a beautiful ASCII table.
func PrintTicketsTable(w io.Writer, tickets []TicketInfo) {
	if len(tickets) == 0 {
		fmt.Fprintln(w, "No tickets found.")
		return
	}

	headers := []string{"ID", "STATUS", "PRIORITY", "OWNER", "UPDATED AT", "TITLE"}
	rows := make([][]string, len(tickets))

	for i, t := range tickets {
		rows[i] = []string{
			t.Meta.ID,
			strings.ToUpper(t.Meta.Status),
			strings.ToUpper(t.Meta.Priority),
			t.Meta.Owner,
			t.Meta.UpdatedAt,
			t.Meta.Title,
		}
	}

	// Compute max column widths
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	for _, row := range rows {
		for i, val := range row {
			// Handle wide characters or simple length. Simple length is fine for standard Latin ASCII.
			width := len(val)
			if width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	// Print headers
	printRow(w, headers, colWidths)

	// Print separator
	sep := make([]string, len(headers))
	for i, w := range colWidths {
		sep[i] = strings.Repeat("-", w)
	}
	printRow(w, sep, colWidths)

	// Print rows
	for _, row := range rows {
		printRow(w, row, colWidths)
	}

	fmt.Fprintf(w, "\nTotal: %d tickets\n", len(tickets))
}

func printRow(w io.Writer, row []string, colWidths []int) {
	var line strings.Builder
	for i, val := range row {
		if i > 0 {
			line.WriteString("  |  ")
		}
		// Left align
		formatStr := fmt.Sprintf("%%-%ds", colWidths[i])
		line.WriteString(fmt.Sprintf(formatStr, val))
	}
	fmt.Fprintln(w, line.String())
}

// PrintValidationErrors formats validation results.
func PrintValidationErrors(w io.Writer, errors []ValidationError, format string) {
	if format == "json" {
		if errors == nil {
			errors = []ValidationError{}
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(errors)
		return
	}

	if len(errors) == 0 {
		fmt.Fprintln(w, "✅ All ticket files are valid!")
		return
	}

	fmt.Fprintf(w, "❌ Found %d validation errors:\n", len(errors))
	for _, err := range errors {
		fmt.Fprintf(w, "  - %s: %s\n", err.FilePath, err.ErrorMsg)
	}
}

// PrintMigrationResults formats and prints migration outcomes.
func PrintMigrationResults(w io.Writer, results []MigrationResult, format string) {
	if format == "json" {
		if results == nil {
			results = []MigrationResult{}
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(results)
		return
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "No legacy tickets found to migrate.")
		return
	}

	fmt.Fprintf(w, "Migrated %d ticket(s):\n", len(results))
	for _, res := range results {
		if res.Action == "skipped" {
			reasonStr := ""
			if res.Reason != "" {
				reasonStr = ": " + strings.ReplaceAll(res.Reason, "_", " ")
			}
			fmt.Fprintf(w, "  - [skipped%s] %s\n", reasonStr, res.OriginalPath)
		} else if res.OriginalPath == res.NewPath {
			fmt.Fprintf(w, "  - [%s] %s (ID: %s, Status: %s)\n", res.Action, res.OriginalPath, res.ID, res.Status)
		} else {
			fmt.Fprintf(w, "  - [%s] %s -> %s (ID: %s, Status: %s)\n", res.Action, res.OriginalPath, res.NewPath, res.ID, res.Status)
		}
	}
}
