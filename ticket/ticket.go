package ticket

import (
	"bytes"
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

// ValidateSingleTicket validates a single ticket file against schema rules.
func ValidateSingleTicket(filePath string, config *TicketConfig) ([]ValidationError, error) {
	var errors []ValidationError

	name := filepath.Base(filePath)
	currentDir := filepath.Dir(filePath)

	absD, err := filepath.Abs(currentDir)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(currentDir)
	} else {
		dirBase = filepath.Base(absD)
	}
	dirBaseLower := strings.ToLower(dirBase)

	// Check if it's named like a ticket
	nameUpper := strings.ToUpper(name)
	isTicketName := strings.HasPrefix(nameUpper, "BUG-") || strings.HasPrefix(nameUpper, "TASK-") || strings.HasPrefix(nameUpper, "TICKET-")

	meta, _, err := ParseFile(filePath)
	if err != nil {
		if isTicketName {
			errors = append(errors, ValidationError{
				FilePath: filepath.ToSlash(filePath),
				ErrorMsg: fmt.Sprintf("failed to parse ticket file: %v", err),
			})
		}
		return errors, nil
	}

	if meta.ID == "" {
		if isTicketName {
			errors = append(errors, ValidationError{
				FilePath: filepath.ToSlash(filePath),
				ErrorMsg: "missing 'id' field in frontmatter",
			})
		}
		return errors, nil
	}

	// 1. Validate type matches parent directory base name
	if !strings.EqualFold(meta.Type, dirBaseLower) {
		errors = append(errors, ValidationError{
			FilePath: filepath.ToSlash(filePath),
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
		errors = append(errors, ValidationError{
			FilePath: filepath.ToSlash(filePath),
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
			errors = append(errors, ValidationError{
				FilePath: filepath.ToSlash(filePath),
				ErrorMsg: fmt.Sprintf("invalid priority '%s' (must be critical, major, minor, or low)", meta.Priority),
			})
		}
	}

	// 4. Validate required extra fields from ticket.yaml
	if config != nil {
		for _, extraF := range config.ExtraFields {
			if extraF.Required {
				val, exists := getFieldValue(meta, extraF.Name)
				if !exists || strings.TrimSpace(val) == "" {
					errors = append(errors, ValidationError{
						FilePath: filepath.ToSlash(filePath),
						ErrorMsg: fmt.Sprintf("missing required extra field '%s' specified in ticket.yaml", extraF.Name),
					})
				}
			}
		}
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

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(currentDir, name)
		fileErrors, err := ValidateSingleTicket(path, config)
		if err != nil {
			return err
		}
		*errors = append(*errors, fileErrors...)
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
	OriginalPath  string   `json:"original_path"`
	NewPath       string   `json:"new_path"`
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Status        string   `json:"status"`
	Action        string   `json:"action"` // "renamed & updated", "updated frontmatter", "skipped"
	File          string   `json:"file"`
	Reason        string   `json:"reason,omitempty"` // "already_standardized", "conflict", 等
	ChangedFields []string `json:"changed_fields"`
}

// MigrateOptions defines parameters for the ticket migration process.
type MigrateOptions struct {
	DryRun      bool
	OnlyInvalid bool
}

// MigrateTickets scans the target directory and migrates legacy markdown tickets to standard format.
func MigrateTickets(dir string) ([]MigrationResult, error) {
	return MigrateTicketsWithOptions(dir, MigrateOptions{})
}

// MigrateTicketsWithOptions scans the target path (file or directory) and migrates tickets.
func MigrateTicketsWithOptions(targetPath string, opts MigrateOptions) ([]MigrationResult, error) {
	var results []MigrationResult

	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access target path %s: %w", targetPath, err)
	}

	if !fi.IsDir() {
		// It's a single file.
		currentDir := filepath.Dir(targetPath)
		config, err := LoadConfig(currentDir)
		if err != nil {
			return nil, err
		}

		res, processed, err := migrateSingleFile(targetPath, config, opts)
		if err != nil {
			return nil, err
		}
		if processed {
			results = append(results, res)
		}
		return results, nil
	}

	// It's a directory.
	if err := migrateDirWithOptions(targetPath, nil, opts, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = []MigrationResult{}
	}
	return results, nil
}

func migrateDirWithOptions(currentDir string, parentConfig *TicketConfig, opts MigrateOptions, results *[]MigrationResult) error {
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

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		filePath := filepath.Join(currentDir, name)
		res, processed, err := migrateSingleFile(filePath, config, opts)
		if err != nil {
			return err
		}
		if processed {
			*results = append(*results, res)
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

		if err := migrateDirWithOptions(subPath, config, opts, results); err != nil {
			return err
		}
	}

	return nil
}

func migrateSingleFile(filePath string, config *TicketConfig, opts MigrateOptions) (MigrationResult, bool, error) {
	name := filepath.Base(filePath)
	currentDir := filepath.Dir(filePath)
	nameUpper := strings.ToUpper(name)

	isTicketName := strings.HasPrefix(nameUpper, "BUG-") || strings.HasPrefix(nameUpper, "TASK-") || strings.HasPrefix(nameUpper, "TICKET-")
	if !isTicketName {
		return MigrationResult{}, false, nil
	}

	// 1. If OnlyInvalid, check if already valid
	if opts.OnlyInvalid {
		errors, err := ValidateSingleTicket(filePath, config)
		if err != nil {
			return MigrationResult{}, false, err
		}
		if len(errors) == 0 {
			meta, _, _ := ParseFile(filePath)
			if meta == nil {
				meta = &TicketMeta{}
			}
			return MigrationResult{
				OriginalPath:  filepath.ToSlash(filePath),
				NewPath:       filepath.ToSlash(filePath),
				ID:            meta.ID,
				Title:         meta.Title,
				Status:        meta.Status,
				Action:        "skipped",
				Reason:        "already_standardized",
				File:          name,
				ChangedFields: []string{},
			}, true, nil
		}
	}

	// Read original bytes for comparison
	originalBytes, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return MigrationResult{}, false, fmt.Errorf("failed to read file: %w", readErr)
	}

	// Parse legacy filename and potential status
	extractedID, legacyStatus, renamed := parseLegacyFilename(name)

	meta, body, err := ParseFile(filePath)
	hasOriginalFrontmatter := err == nil && meta != nil && meta.ID != ""
	if err != nil {
		body = string(originalBytes)
		meta = &TicketMeta{}
	}

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
	absD, err := filepath.Abs(currentDir)
	var dirBase string
	if err != nil {
		dirBase = filepath.Base(currentDir)
	} else {
		dirBase = filepath.Base(absD)
	}
	dirBaseLower := strings.ToLower(dirBase)
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
		fileInfo, statErr := os.Stat(filePath)
		if statErr == nil {
			meta.CreatedAt = fileInfo.ModTime().Format("2006-01-02 15:04")
		} else {
			meta.CreatedAt = nowStr
		}
	}

	// Auto resolved_at terminal status logic
	isTerminal := meta.Status == "passed" || meta.Status == "rejected"
	if isTerminal {
		if meta.ResolvedAt == "" {
			meta.ResolvedAt = nowStr
		}
	} else {
		meta.ResolvedAt = ""
	}

	// Prepare for comparison (initially keep UpdatedAt as is)
	meta.UpdatedAt = originalMeta.UpdatedAt

	renderedBytes, renderErr := RenderTicket(meta, body)
	if renderErr != nil {
		return MigrationResult{}, false, fmt.Errorf("failed to render ticket: %w", renderErr)
	}

	targetFileName := meta.ID + ".md"
	targetFilePath := filepath.Join(currentDir, targetFileName)

	isContentUnchanged := bytes.Equal(originalBytes, renderedBytes)
	needsRename := renamed && targetFilePath != filePath

	var result MigrationResult
	result.OriginalPath = filepath.ToSlash(filePath)
	result.ID = meta.ID
	result.Title = meta.Title
	result.Status = meta.Status
	result.File = name

	// Build changed fields list
	var changedFields []string
	if originalMeta.ID != meta.ID {
		changedFields = append(changedFields, "id")
	}
	if originalMeta.Title != meta.Title {
		changedFields = append(changedFields, "title")
	}
	if originalMeta.Type != meta.Type {
		changedFields = append(changedFields, "type")
	}
	if originalMeta.Status != meta.Status {
		changedFields = append(changedFields, "status")
	}
	if originalMeta.Priority != meta.Priority {
		changedFields = append(changedFields, "priority")
	}
	if originalMeta.ResolvedAt != meta.ResolvedAt {
		changedFields = append(changedFields, "resolved_at")
	}
	if originalMeta.CreatedAt != meta.CreatedAt {
		changedFields = append(changedFields, "created_at")
	}

	if isContentUnchanged && !needsRename && hasOriginalFrontmatter {
		result.NewPath = filepath.ToSlash(filePath)
		result.Action = "skipped"
		result.Reason = "already_standardized"
		result.ChangedFields = []string{}
		return result, true, nil
	}

	// Update UpdatedAt for writing
	meta.UpdatedAt = nowStr
	changedFields = append(changedFields, "updated_at")
	if !hasOriginalFrontmatter {
		changedFields = append(changedFields, "frontmatter")
	}
	result.ChangedFields = changedFields

	finalRenderedBytes, renderErr := RenderTicket(meta, body)
	if renderErr != nil {
		return MigrationResult{}, false, fmt.Errorf("failed to render ticket with updated_at: %w", renderErr)
	}

	if needsRename {
		if _, err := os.Stat(targetFilePath); err == nil {
			result.NewPath = filepath.ToSlash(filePath)
			result.Action = "skipped"
			result.Reason = "conflict"
			result.ChangedFields = []string{}
			return result, true, nil
		}

		if !opts.DryRun {
			tempPath := targetFilePath + ".tmp"
			if err := os.WriteFile(tempPath, finalRenderedBytes, 0644); err != nil {
				return MigrationResult{}, false, fmt.Errorf("failed to write temp file: %w", err)
			}
			if err := os.Rename(tempPath, targetFilePath); err != nil {
				os.Remove(tempPath)
				return MigrationResult{}, false, fmt.Errorf("failed to rename temp file to target: %w", err)
			}
			if err := os.Remove(filePath); err != nil {
				return MigrationResult{}, false, fmt.Errorf("failed to remove legacy file %s: %w", filePath, err)
			}
		}

		result.NewPath = filepath.ToSlash(targetFilePath)
		result.Action = "renamed & updated"
	} else {
		if !opts.DryRun {
			tempPath := filePath + ".tmp"
			if err := os.WriteFile(tempPath, finalRenderedBytes, 0644); err != nil {
				return MigrationResult{}, false, fmt.Errorf("failed to write temp file: %w", err)
			}
			if err := os.Rename(tempPath, filePath); err != nil {
				os.Remove(tempPath)
				return MigrationResult{}, false, fmt.Errorf("failed to rename temp file to target: %w", err)
			}
		}
		result.NewPath = filepath.ToSlash(filePath)
		result.Action = "updated frontmatter"
	}

	return result, true, nil
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

