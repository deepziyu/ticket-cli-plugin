package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/deepziyu/ticket-cli-plugin/ticket"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "list":
		handleList()
	case "update":
		handleUpdate()
	case "create":
		handleCreate()
	case "show":
		handleShow()
	case "validate":
		handleValidate()
	case "migrate":
		handleMigrate()
	case "dashboard":
		handleDashboard()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Ticket Markdown Management CLI (Go)")
	fmt.Println("Usage: ticket <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  list [dir]         List and filter tickets in a directory")
	fmt.Println("  update <file>      Update metadata fields of a ticket")
	fmt.Println("  create <dir> <id>  Create a new ticket markdown file")
	fmt.Println("  show <file>        Show details of a ticket (JSON/text)")
	fmt.Println("  validate [dir]     Validate ticket files in a directory")
	fmt.Println("  migrate [dir]      Migrate legacy markdown tickets to standard format")
	fmt.Println("  dashboard [dir]    Start web-based Kanban dashboard")
	fmt.Println("\nUse 'ticket <command> --help' for details on a specific command.")
}

func handleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	status := fs.String("status", "", "Filter by status (open, fixing, resolved, passed, rejected)")
	priority := fs.String("priority", "", "Filter by priority (critical, major, minor, low)")
	owner := fs.String("owner", "", "Filter by owner")
	ticketType := fs.String("type", "", "Filter by type (bug, task, feature, etc.)")
	format := fs.String("format", "table", "Output format (table, json)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket list [dir] [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected at most 1 directory, got: %v\n", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	tickets, err := ticket.ListTickets(dir, *status, *priority, *owner, *ticketType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if strings.ToLower(*format) == "json" {
		if err := ticket.PrintTicketsJSON(os.Stdout, tickets); err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		ticket.PrintTicketsTable(os.Stdout, tickets)
	}
}

func handleUpdate() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	status := fs.String("status", "", "Update status field")
	priority := fs.String("priority", "", "Update priority field")
	owner := fs.String("owner", "", "Update owner field")
	conclusion := fs.String("conclusion", "", "Update resolution/conclusion description")
	resolvedAt := fs.String("resolved-at", "", "Override resolved time")
	title := fs.String("title", "", "Update title")
	ticketType := fs.String("type", "", "Update type")
	format := fs.String("format", "text", "Output format (text, json)")

	var fields stringSlice
	fs.Var(&fields, "field", "Update custom key=value field (can be specified multiple times)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket update <file_path> [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 1 {
		if fs.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "Error: missing target file path")
		} else {
			fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected 1 file path, got: %v\n", fs.Args())
		}
		fs.Usage()
		os.Exit(1)
	}

	filePath := fs.Arg(0)

	oldMeta, oldBody, err := ticket.ParseFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ticket: %v\n", err)
		os.Exit(1)
	}

	updates := make(map[string]string)
	if *status != "" {
		updates["status"] = *status
	}
	if *priority != "" {
		updates["priority"] = *priority
	}
	if *owner != "" {
		updates["owner"] = *owner
	}
	if *conclusion != "" {
		updates["conclusion"] = *conclusion
	}
	if *resolvedAt != "" {
		updates["resolved_at"] = *resolvedAt
	}
	if *title != "" {
		updates["title"] = *title
	}
	if *ticketType != "" {
		updates["type"] = *ticketType
	}

	// Parse custom fields
	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			updates[parts[0]] = parts[1]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: invalid field format '%s', should be key=value\n", f)
		}
	}

	meta, err := ticket.UpdateTicket(filePath, updates)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	updatedFields := make(map[string][2]string)
	if oldMeta.Status != meta.Status {
		updatedFields["status"] = [2]string{oldMeta.Status, meta.Status}
	}
	if oldMeta.Priority != meta.Priority {
		updatedFields["priority"] = [2]string{oldMeta.Priority, meta.Priority}
	}
	if oldMeta.Owner != meta.Owner {
		updatedFields["owner"] = [2]string{oldMeta.Owner, meta.Owner}
	}
	if oldMeta.Conclusion != meta.Conclusion {
		updatedFields["conclusion"] = [2]string{oldMeta.Conclusion, meta.Conclusion}
	}
	if oldMeta.ResolvedAt != meta.ResolvedAt {
		updatedFields["resolved_at"] = [2]string{oldMeta.ResolvedAt, meta.ResolvedAt}
	}
	if oldMeta.Title != meta.Title {
		updatedFields["title"] = [2]string{oldMeta.Title, meta.Title}
	}
	if oldMeta.Type != meta.Type {
		updatedFields["type"] = [2]string{oldMeta.Type, meta.Type}
	}
	if oldMeta.UpdatedAt != meta.UpdatedAt {
		updatedFields["updated_at"] = [2]string{oldMeta.UpdatedAt, meta.UpdatedAt}
	}
	
	// Compare extra fields
	for k, v := range meta.ExtraFields {
		oldVal, ok := oldMeta.ExtraFields[k]
		newValStr := fmt.Sprintf("%v", v)
		oldValStr := ""
		if ok {
			oldValStr = fmt.Sprintf("%v", oldVal)
		}
		if !ok || oldValStr != newValStr {
			updatedFields[k] = [2]string{oldValStr, newValStr}
		}
	}
	for k, v := range oldMeta.ExtraFields {
		if _, ok := meta.ExtraFields[k]; !ok {
			updatedFields[k] = [2]string{fmt.Sprintf("%v", v), ""}
		}
	}

	if strings.ToLower(*format) == "json" {
		ticketInfo := ticket.TicketInfo{
			Meta:     meta,
			FilePath: filepath.ToSlash(filePath),
			FileName: filepath.Base(filePath),
		}
		
		type TicketUpdateJSON struct {
			ticket.TicketInfo
			Body string `json:"body"`
		}
		
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(TicketUpdateJSON{TicketInfo: ticketInfo, Body: oldBody})
	} else {
		fmt.Printf("Successfully updated ticket '%s' (Status: %s)\n", meta.ID, strings.ToUpper(meta.Status))
		if len(updatedFields) > 0 {
			fmt.Println("Updated Fields:")
			for k, vals := range updatedFields {
				if vals[0] == "" {
					fmt.Printf("  - %s: -> %q\n", k, vals[1])
				} else {
					fmt.Printf("  - %s: %q -> %q\n", k, vals[0], vals[1])
				}
			}
		}
	}
}

func handleCreate() {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	title := fs.String("title", "Untitled Ticket", "Ticket title")
	ticketType := fs.String("type", "task", "Ticket type (e.g. bug, task, feature)")
	status := fs.String("status", "open", "Initial ticket status")
	priority := fs.String("priority", "minor", "Ticket priority (critical, major, minor, low)")
	owner := fs.String("owner", "", "Ticket owner")
	body := fs.String("body", "", "Custom markdown body content")

	var fields stringSlice
	fs.Var(&fields, "field", "Add custom key=value field (can be specified multiple times)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket create <dir_path> <id> [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 2 {
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "Error: missing target directory or ticket ID")
		} else {
			fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected 1 directory path and 1 ticket ID, got: %v\n", fs.Args())
		}
		fs.Usage()
		os.Exit(1)
	}

	dirPath := fs.Arg(0)
	id := fs.Arg(1)

	extraFields := make(map[string]string)
	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			extraFields[parts[0]] = parts[1]
		}
	}

	createdPath, err := ticket.CreateTicket(dirPath, id, *title, *ticketType, *status, *priority, *owner, extraFields, *body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created ticket file: %s\n", createdPath)
}

func handleShow() {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket show <file_path> [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 1 {
		if fs.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "Error: missing target file path")
		} else {
			fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected 1 file path, got: %v\n", fs.Args())
		}
		fs.Usage()
		os.Exit(1)
	}

	filePath := fs.Arg(0)

	meta, body, err := ticket.ParseFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if strings.ToLower(*format) == "json" {
		ticketInfo := ticket.TicketInfo{
			Meta:     meta,
			FilePath: filepath.ToSlash(filePath),
			FileName: filepath.Base(filePath),
		}
		
		type TicketShowJSON struct {
			ticket.TicketInfo
			Body string `json:"body"`
		}
		
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(TicketShowJSON{TicketInfo: ticketInfo, Body: body})
	} else {
		// Output human-readable plain text
		fmt.Printf("Ticket ID:    %s\n", meta.ID)
		fmt.Printf("Title:        %s\n", meta.Title)
		fmt.Printf("Type:         %s\n", meta.Type)
		fmt.Printf("Status:       %s\n", strings.ToUpper(meta.Status))
		fmt.Printf("Priority:     %s\n", strings.ToUpper(meta.Priority))
		fmt.Printf("Owner:        %s\n", meta.Owner)
		fmt.Printf("Created At:   %s\n", meta.CreatedAt)
		fmt.Printf("Updated At:   %s\n", meta.UpdatedAt)
		if meta.ResolvedAt != "" {
			fmt.Printf("Resolved At:  %s\n", meta.ResolvedAt)
			fmt.Printf("Conclusion:   %s\n", meta.Conclusion)
		}
		
		if len(meta.ExtraFields) > 0 {
			fmt.Println("\nExtra Metadata:")
			for k, v := range meta.ExtraFields {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}

		fmt.Println("\nMarkdown Body:")
		fmt.Println("-------------------------------------------")
		fmt.Println(body)
	}
}

func handleValidate() {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket validate [dir] [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected at most 1 directory, got: %v\n", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	errors, err := ticket.ValidateTickets(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	ticket.PrintValidationErrors(os.Stdout, errors, strings.ToLower(*format))
	
	if len(errors) > 0 {
		os.Exit(1) // Exit with non-zero if validation fails
	}
}

func handleMigrate() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket migrate [dir] [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected at most 1 directory, got: %v\n", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	results, err := ticket.MigrateTickets(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running migration: %v\n", err)
		os.Exit(1)
	}

	ticket.PrintMigrationResults(os.Stdout, results, strings.ToLower(*format))
}

func reorderArgs(args []string) []string {
	var flags []string
	var positionals []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") || arg == "--help" || arg == "-h" {
				flags = append(flags, arg)
				i++
			} else {
				if i+1 < len(args) {
					flags = append(flags, arg, args[i+1])
					i += 2
				} else {
					flags = append(flags, arg)
					i++
				}
			}
		} else {
			positionals = append(positionals, arg)
			i++
		}
	}
	return append(flags, positionals...)
}

func handleDashboard() {
	fs := flag.NewFlagSet("dashboard", flag.ExitOnError)
	port := fs.Int("port", 8080, "Web server port")
	fs.IntVar(port, "p", 8080, "Web server port (shorthand)")

	fs.Usage = func() {
		fmt.Println("Usage: ticket dashboard [dir] [flags]")
		fs.PrintDefaults()
	}

	args := reorderArgs(os.Args[2:])
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "Error: too many arguments. Expected at most 1 directory, got: %v\n", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	if err := ticket.StartDashboard(dir, *port); err != nil {
		fmt.Fprintf(os.Stderr, "Server Error: %v\n", err)
		os.Exit(1)
	}
}

