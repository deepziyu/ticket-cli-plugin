package ticket

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TicketInfo holds the metadata and file system attributes of a ticket.
type TicketInfo struct {
	Meta     *TicketMeta `json:"meta"`
	FilePath string      `json:"file_path"`
	FileName string      `json:"file_name"`
}

// ListTickets scans the target directory and returns filtered ticket info.
func ListTickets(dir string, statusFilter, priorityFilter, ownerFilter, typeFilter string) ([]TicketInfo, error) {
	var tickets []TicketInfo
	if err := scanDir(dir, statusFilter, priorityFilter, ownerFilter, typeFilter, &tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}

func scanDir(currentDir string, statusFilter, priorityFilter, ownerFilter, typeFilter string, tickets *[]TicketInfo) error {
	config, err := LoadConfig(currentDir)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(currentDir, name)
		meta, _, err := ParseFile(path)
		if err != nil {
			// Skip files that are not valid tickets
			continue
		}

		// If ID is missing, it's not a standard ticket
		if meta.ID == "" {
			continue
		}

		// Apply filters (case-insensitive)
		if statusFilter != "" && !strings.EqualFold(meta.Status, statusFilter) {
			continue
		}
		if priorityFilter != "" && !strings.EqualFold(meta.Priority, priorityFilter) {
			continue
		}
		if ownerFilter != "" && !strings.EqualFold(meta.Owner, ownerFilter) {
			continue
		}
		if typeFilter != "" && !strings.EqualFold(meta.Type, typeFilter) {
			continue
		}

		*tickets = append(*tickets, TicketInfo{
			Meta:     meta,
			FilePath: filepath.ToSlash(path),
			FileName: name,
		})
	}

	// Recurse into configured sub_dirs
	for _, subDir := range config.SubDirs {
		subPath := filepath.Join(currentDir, subDir)
		info, err := os.Stat(subPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}

		if err := scanDir(subPath, statusFilter, priorityFilter, ownerFilter, typeFilter, tickets); err != nil {
			return err
		}
	}

	return nil
}

// UpdateTicket applies programmatic updates to a ticket's frontmatter.
func UpdateTicket(filePath string, updates map[string]string) (*TicketMeta, error) {
	meta, body, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	nowStr := time.Now().Format("2006-01-02 15:04")
	meta.UpdatedAt = nowStr

	for k, v := range updates {
		switch strings.ToLower(k) {
		case "status":
			oldStatus := meta.Status
			meta.Status = strings.ToLower(v)
			
			// Auto handle resolved_at for terminal states
			isTerminal := meta.Status == "passed" || meta.Status == "rejected"
			wasTerminal := oldStatus == "passed" || oldStatus == "rejected"
			
			if isTerminal && !wasTerminal {
				if meta.ResolvedAt == "" {
					meta.ResolvedAt = nowStr
				}
			} else if !isTerminal && wasTerminal {
				meta.ResolvedAt = ""
			}
		case "owner":
			meta.Owner = v
		case "priority":
			meta.Priority = strings.ToLower(v)
		case "conclusion":
			meta.Conclusion = v
		case "resolved_at":
			meta.ResolvedAt = v
		case "title":
			meta.Title = v
		case "type":
			// Will be overridden by directory name base rule
		default:
			// Save in extra fields dynamically
			if meta.ExtraFields == nil {
				meta.ExtraFields = make(map[string]interface{})
			}
			meta.ExtraFields[k] = v
		}
	}

	// Auto-correct ticket type based on current parent directory base name
	dirPath := filepath.Dir(filePath)
	absD, err := filepath.Abs(dirPath)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(dirPath)
	} else {
		dirBase = filepath.Base(absD)
	}
	meta.Type = strings.ToLower(dirBase)

	if err := WriteFile(filePath, meta, body); err != nil {
		return nil, fmt.Errorf("failed to save updated ticket: %w", err)
	}

	return meta, nil
}

// CreateTicket creates a new ticket file with standard YAML Frontmatter and body templates.
func CreateTicket(dirPath string, id string, title, ticketType, status, priority, owner string, extraFields map[string]string, body string) (string, error) {
	// Standardize filename
	fileName := id
	if !strings.HasSuffix(strings.ToLower(fileName), ".md") {
		fileName += ".md"
	}
	filePath := filepath.Join(dirPath, fileName)

	// Check if already exists
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("ticket file already exists: %s", filePath)
	}

	nowStr := time.Now().Format("2006-01-02 15:04")
	
	if status == "" {
		status = "open"
	}

	// Auto-determine ticket type based on target directory base name
	absD, err := filepath.Abs(dirPath)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(dirPath)
	} else {
		dirBase = filepath.Base(absD)
	}
	dirBaseLower := strings.ToLower(dirBase)

	meta := &TicketMeta{
		ID:          id,
		Title:       title,
		Type:        dirBaseLower, // Auto force/correct to directory name
		Status:      strings.ToLower(status),
		Priority:    strings.ToLower(priority),
		Owner:       owner,
		CreatedAt:   nowStr,
		UpdatedAt:   nowStr,
		ExtraFields: make(map[string]interface{}),
	}

	for k, v := range extraFields {
		meta.ExtraFields[k] = v
	}

	// Auto resolved_at for initial terminal status
	if meta.Status == "passed" || meta.Status == "rejected" {
		meta.ResolvedAt = nowStr
	}

	if body == "" {
		body = fmt.Sprintf("# %s: %s\n\n## 描述\n在这里填写工单详细描述...\n", id, title)
	}

	if err := WriteFile(filePath, meta, body); err != nil {
		return "", fmt.Errorf("failed to create ticket: %w", err)
	}

	return filepath.ToSlash(filePath), nil
}

// ValidationError represents a specific schema violation in a ticket file.
type ValidationError struct {
	FilePath string
	ErrorMsg string
}

// ValidateTickets scans and validates all tickets in a directory.
func ValidateTickets(dir string) ([]ValidationError, error) {
	errors := []ValidationError{}
	if err := validateDir(dir, nil, &errors); err != nil {
		return nil, err
	}
	return errors, nil
}

func validateDir(currentDir string, parentConfig *TicketConfig, errors *[]ValidationError) error {
	config, err := LoadConfig(currentDir)
	if err != nil {
		return err
	}

	// Inherit ExtraFields from parent if local is empty
	if len(config.ExtraFields) == 0 && parentConfig != nil {
		config.ExtraFields = parentConfig.ExtraFields
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	absD, err := filepath.Abs(currentDir)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(currentDir)
	} else {
		dirBase = filepath.Base(absD)
	}
	dirBaseLower := strings.ToLower(dirBase)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(currentDir, name)

		// Check if it's named like a ticket
		isTicketName := strings.HasPrefix(name, "BUG-") || strings.HasPrefix(name, "TASK-") || strings.HasPrefix(name, "TICKET-")

		meta, _, err := ParseFile(path)
		if err != nil {
			if isTicketName {
				*errors = append(*errors, ValidationError{
					FilePath: filepath.ToSlash(path),
					ErrorMsg: fmt.Sprintf("failed to parse ticket file: %v", err),
				})
			}
			continue
		}

		if meta.ID == "" {
			if isTicketName {
				*errors = append(*errors, ValidationError{
					FilePath: filepath.ToSlash(path),
					ErrorMsg: "missing 'id' field in frontmatter",
				})
			}
			continue
		}

		// 1. Validate type matches parent directory base name
		if !strings.EqualFold(meta.Type, dirBaseLower) {
			*errors = append(*errors, ValidationError{
				FilePath: filepath.ToSlash(path),
				ErrorMsg: fmt.Sprintf("invalid ticket type '%s', must match directory name '%s'", meta.Type, dirBaseLower),
			})
		}

		// 2. Validate status values
		validStatuses := map[string]bool{
			"open":     true,
			"fixing":   true,
			"resolved": true,
			"passed":   true,
			"rejected": true,
		}
		if !validStatuses[meta.Status] {
			*errors = append(*errors, ValidationError{
				FilePath: filepath.ToSlash(path),
				ErrorMsg: fmt.Sprintf("invalid status '%s' (must be open, fixing, resolved, passed, or rejected)", meta.Status),
			})
		}

		// 3. Validate priority values if present
		if meta.Priority != "" {
			validPriorities := map[string]bool{
				"critical": true,
				"major":    true,
				"minor":    true,
				"low":      true,
			}
			if !validPriorities[meta.Priority] {
				*errors = append(*errors, ValidationError{
					FilePath: filepath.ToSlash(path),
					ErrorMsg: fmt.Sprintf("invalid priority '%s' (must be critical, major, minor, or low)", meta.Priority),
				})
			}
		}

		// 4. Validate required extra fields from ticket.yaml
		for _, extraF := range config.ExtraFields {
			if extraF.Required {
				val, exists := getFieldValue(meta, extraF.Name)
				if !exists || strings.TrimSpace(val) == "" {
					*errors = append(*errors, ValidationError{
						FilePath: filepath.ToSlash(path),
						ErrorMsg: fmt.Sprintf("missing required extra field '%s' specified in ticket.yaml", extraF.Name),
					})
				}
			}
		}
	}

	// Recurse into configured sub_dirs
	for _, subDir := range config.SubDirs {
		subPath := filepath.Join(currentDir, subDir)
		info, err := os.Stat(subPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}

		if err := validateDir(subPath, config, errors); err != nil {
			return err
		}
	}

	return nil
}

func getFieldValue(meta *TicketMeta, name string) (string, bool) {
	switch strings.ToLower(name) {
	case "id":
		return meta.ID, meta.ID != ""
	case "title":
		return meta.Title, meta.Title != ""
	case "type":
		return meta.Type, meta.Type != ""
	case "status":
		return meta.Status, meta.Status != ""
	case "priority":
		return meta.Priority, meta.Priority != ""
	case "owner":
		return meta.Owner, meta.Owner != ""
	case "created_at":
		return meta.CreatedAt, meta.CreatedAt != ""
	case "updated_at":
		return meta.UpdatedAt, meta.UpdatedAt != ""
	case "resolved_at":
		return meta.ResolvedAt, meta.ResolvedAt != ""
	case "conclusion":
		return meta.Conclusion, meta.Conclusion != ""
	}
	if meta.ExtraFields != nil {
		if val, ok := meta.ExtraFields[name]; ok {
			valStr := fmt.Sprintf("%v", val)
			return valStr, valStr != ""
		}
	}
	return "", false
}

// MigrationResult represents the outcome of migrating a single ticket file.
type MigrationResult struct {
	OriginalPath string `json:"original_path"`
	NewPath      string `json:"new_path"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Action       string `json:"action"` // "renamed & updated", "updated frontmatter", "skipped (conflict)"
}

// MigrateTickets scans the target directory and migrates legacy markdown tickets to standard format.
func MigrateTickets(dir string) ([]MigrationResult, error) {
	var results []MigrationResult
	if err := migrateDir(dir, nil, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []MigrationResult{}
	}
	return results, nil
}

func migrateDir(currentDir string, parentConfig *TicketConfig, results *[]MigrationResult) error {
	config, err := LoadConfig(currentDir)
	if err != nil {
		return err
	}

	if len(config.ExtraFields) == 0 && parentConfig != nil {
		config.ExtraFields = parentConfig.ExtraFields
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	absD, err := filepath.Abs(currentDir)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(currentDir)
	} else {
		dirBase = filepath.Base(absD)
	}
	dirBaseLower := strings.ToLower(dirBase)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		// Standard tickets start with BUG-, TASK-, TICKET- (case-insensitive)
		nameUpper := strings.ToUpper(name)
		isTicketName := strings.HasPrefix(nameUpper, "BUG-") || strings.HasPrefix(nameUpper, "TASK-") || strings.HasPrefix(nameUpper, "TICKET-")
		if !isTicketName {
			continue
		}

		filePath := filepath.Join(currentDir, name)
		
		// Parse legacy filename and potential status
		extractedID, legacyStatus, renamed := parseLegacyFilename(name)
		
		meta, body, err := ParseFile(filePath)
		if err != nil {
			// Even if parse fails, we can try to recover if it was due to invalid YAML frontmatter.
			// Let's read the file content directly.
			contentBytes, readErr := os.ReadFile(filePath)
			if readErr != nil {
				continue
			}
			// Treat the whole file as body
			body = string(contentBytes)
			meta = &TicketMeta{}
		}

		// Initialize metadata if empty
		if meta == nil {
			meta = &TicketMeta{}
		}

		originalMeta := *meta

		// Set or enrich metadata
		if meta.ID == "" {
			meta.ID = extractedID
		}
		
		// Extract title from body if empty
		if meta.Title == "" {
			meta.Title = extractTitleFromBody(body, meta.ID)
			if meta.Title == "" {
				meta.Title = "Untitled Ticket"
			}
		}

		// Set type based on directory
		meta.Type = dirBaseLower

		// Determine status
		if legacyStatus != "" {
			meta.Status = legacyStatus
		} else if meta.Status == "" {
			meta.Status = "open"
		}

		// Set priority if empty
		if meta.Priority == "" {
			meta.Priority = "minor"
		}

		// Times
		nowStr := time.Now().Format("2006-01-02 15:04")
		if meta.CreatedAt == "" {
			// Get file creation/modification time as a fallback
			fileInfo, statErr := os.Stat(filePath)
			if statErr == nil {
				meta.CreatedAt = fileInfo.ModTime().Format("2006-01-02 15:04")
			} else {
				meta.CreatedAt = nowStr
			}
		}
		
		meta.UpdatedAt = nowStr

		// Handle resolved_at terminal status logic
		isTerminal := meta.Status == "passed" || meta.Status == "rejected"
		if isTerminal {
			if meta.ResolvedAt == "" {
				meta.ResolvedAt = nowStr
			}
		} else {
			meta.ResolvedAt = ""
		}

		// Check if we need to write changes
		metaChanged := originalMeta.ID != meta.ID ||
			originalMeta.Title != meta.Title ||
			originalMeta.Type != meta.Type ||
			originalMeta.Status != meta.Status ||
			originalMeta.Priority != meta.Priority ||
			originalMeta.ResolvedAt != meta.ResolvedAt ||
			originalMeta.CreatedAt != meta.CreatedAt

		targetFileName := meta.ID + ".md"
		targetFilePath := filepath.Join(currentDir, targetFileName)

		result := MigrationResult{
			OriginalPath: filepath.ToSlash(filePath),
			ID:           meta.ID,
			Title:        meta.Title,
			Status:       meta.Status,
		}

		if renamed && targetFilePath != filePath {
			// Check conflict
			if _, err := os.Stat(targetFilePath); err == nil {
				result.NewPath = filepath.ToSlash(filePath)
				result.Action = "skipped (conflict: target file already exists)"
				*results = append(*results, result)
				continue
			}

			// Write to new file and delete old file
			if err := WriteFile(targetFilePath, meta, body); err != nil {
				return fmt.Errorf("failed to write migrated file %s: %w", targetFilePath, err)
			}
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove legacy file %s: %w", filePath, err)
			}

			result.NewPath = filepath.ToSlash(targetFilePath)
			result.Action = "renamed & updated"
		} else {
			// No rename, just update metadata if changed
			if metaChanged || !strings.Contains(body, "id:") { // write anyway if frontmatter was missing
				if err := WriteFile(filePath, meta, body); err != nil {
					return fmt.Errorf("failed to write migrated file %s: %w", filePath, err)
				}
				result.Action = "updated frontmatter"
			} else {
				result.Action = "no change needed"
			}
			result.NewPath = filepath.ToSlash(filePath)
		}

		*results = append(*results, result)
	}

	// Recurse into configured sub_dirs
	for _, subDir := range config.SubDirs {
		subPath := filepath.Join(currentDir, subDir)
		info, err := os.Stat(subPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}

		if err := migrateDir(subPath, config, results); err != nil {
			return err
		}
	}

	return nil
}

func parseLegacyFilename(filename string) (id string, status string, renamed bool) {
	base := filename
	if strings.HasSuffix(strings.ToLower(base), ".md") {
		base = base[:len(base)-3]
	}

	suffixes := []struct {
		suffix string
		status string
	}{
		{"-passed", "passed"},
		{"-pass", "passed"},
		{"-failed", "rejected"},
		{"-fail", "rejected"},
		{"-rejected", "rejected"},
		{"-fixing", "fixing"},
		{"-resolved", "resolved"},
		{"-open", "open"},
	}

	for _, s := range suffixes {
		if strings.HasSuffix(strings.ToLower(base), s.suffix) {
			id = base[:len(base)-len(s.suffix)]
			return id, s.status, true
		}
	}

	return base, "", false
}

func extractTitleFromBody(body string, id string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimPrefix(title, "缺陷报告：")
			title = strings.TrimPrefix(title, "缺陷报告:")
			title = strings.TrimPrefix(title, "缺陷报告")
			
			idColon1 := id + ":"
			idColon2 := id + "："
			if strings.HasPrefix(title, idColon1) {
				title = strings.TrimPrefix(title, idColon1)
			} else if strings.HasPrefix(title, idColon2) {
				title = strings.TrimPrefix(title, idColon2)
			} else if strings.HasPrefix(title, id) {
				title = strings.TrimPrefix(title, id)
			}
			
			title = strings.TrimSpace(title)
			title = strings.TrimLeft(title, ":：- ")
			return title
		}
	}
	return ""
}

// UpdateTicketWithBody applies programmatic updates to frontmatter and overwrites the body content.
func UpdateTicketWithBody(filePath string, updates map[string]string, newBody string) (*TicketMeta, error) {
	meta, _, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	nowStr := time.Now().Format("2006-01-02 15:04")
	meta.UpdatedAt = nowStr

	for k, v := range updates {
		switch strings.ToLower(k) {
		case "status":
			oldStatus := meta.Status
			meta.Status = strings.ToLower(v)
			
			// Auto handle resolved_at for terminal states
			isTerminal := meta.Status == "passed" || meta.Status == "rejected"
			wasTerminal := oldStatus == "passed" || oldStatus == "rejected"
			
			if isTerminal && !wasTerminal {
				if meta.ResolvedAt == "" {
					meta.ResolvedAt = nowStr
				}
			} else if !isTerminal && wasTerminal {
				meta.ResolvedAt = ""
			}
		case "owner":
			meta.Owner = v
		case "priority":
			meta.Priority = strings.ToLower(v)
		case "conclusion":
			meta.Conclusion = v
		case "resolved_at":
			meta.ResolvedAt = v
		case "title":
			meta.Title = v
		case "type":
			// Will be overridden by directory name base rule
		default:
			// Save in extra fields dynamically
			if meta.ExtraFields == nil {
				meta.ExtraFields = make(map[string]interface{})
			}
			meta.ExtraFields[k] = v
		}
	}

	// Auto-correct ticket type based on current parent directory base name
	dirPath := filepath.Dir(filePath)
	absD, err := filepath.Abs(dirPath)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(dirPath)
	} else {
		dirBase = filepath.Base(absD)
	}
	meta.Type = strings.ToLower(dirBase)

	if err := WriteFile(filePath, meta, newBody); err != nil {
		return nil, fmt.Errorf("failed to save updated ticket: %w", err)
	}

	return meta, nil
}

