package ticket

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTicketYAMLAndRecursion(t *testing.T) {
	// 1. Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "ticket-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	bugsDir := filepath.Join(tempDir, "bugs")
	tasksDir := filepath.Join(tempDir, "tasks")
	ignoredDir := filepath.Join(tempDir, "ignored")
	if err := os.Mkdir(bugsDir, 0755); err != nil {
		t.Fatalf("failed to create bugs dir: %v", err)
	}
	if err := os.Mkdir(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}
	if err := os.Mkdir(ignoredDir, 0755); err != nil {
		t.Fatalf("failed to create ignored dir: %v", err)
	}

	// 2. Write ticket.yaml in tempDir
	yamlConfig := `
sub_dirs:
  - bugs
  - tasks
extra_fields:
  - name: module
    required: true
  - name: severity
`
	if err := os.WriteFile(filepath.Join(tempDir, "ticket.yaml"), []byte(yamlConfig), 0644); err != nil {
		t.Fatalf("failed to write ticket.yaml: %v", err)
	}

	// 3. Create ticket in bugs directory (auto-corrects type to "bugs")
	filePath1, err := CreateTicket(bugsDir, "BUG-01", "Fix database connection", "", "open", "major", "dev-agent", map[string]string{"module": "auth", "severity": "high"}, "")
	if err != nil {
		t.Fatalf("failed to create ticket 1: %v", err)
	}

	// Create ticket in tasks directory (auto-corrects type to "tasks")
	filePath2, err := CreateTicket(tasksDir, "TASK-01", "Implement new button", "", "open", "minor", "dev-agent", nil, "")
	if err != nil {
		t.Fatalf("failed to create ticket 2: %v", err)
	}

	// Create ticket in ignored directory (should not be loaded since "ignored" is not in sub_dirs)
	_, err = CreateTicket(ignoredDir, "TASK-02", "This should be ignored", "", "open", "low", "dev-agent", nil, "")
	if err != nil {
		t.Fatalf("failed to create ticket 3: %v", err)
	}

	// 4. Test ListTickets
	tickets, err := ListTickets(tempDir, "", "", "", "")
	if err != nil {
		t.Fatalf("ListTickets failed: %v", err)
	}

	// Should list BUG-01 and TASK-01, but NOT TASK-02
	if len(tickets) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(tickets))
	}

	foundBug01 := false
	foundTask01 := false
	for _, tk := range tickets {
		if tk.Meta.ID == "BUG-01" {
			foundBug01 = true
			if tk.Meta.Type != "bugs" {
				t.Errorf("BUG-01: expected type 'bugs', got '%s'", tk.Meta.Type)
			}
		}
		if tk.Meta.ID == "TASK-01" {
			foundTask01 = true
			if tk.Meta.Type != "tasks" {
				t.Errorf("TASK-01: expected type 'tasks', got '%s'", tk.Meta.Type)
			}
		}
		if tk.Meta.ID == "TASK-02" {
			t.Errorf("TASK-02 should have been ignored from recursion scan")
		}
	}

	if !foundBug01 || !foundTask01 {
		t.Errorf("did not find expected tickets (BUG-01: %t, TASK-01: %t)", foundBug01, foundTask01)
	}

	// 5. Test ValidateTickets
	// Currently, TASK-01 is missing the required field "module" defined in root ticket.yaml
	validationErrors, err := ValidateTickets(tempDir)
	if err != nil {
		t.Fatalf("ValidateTickets failed: %v", err)
	}

	// We expect validation error for TASK-01 because "module" is required in tempDir's ticket.yaml,
	// but it is not defined in TASK-01.
	if len(validationErrors) != 1 {
		t.Errorf("expected 1 validation error, got %d", len(validationErrors))
		for _, ve := range validationErrors {
			t.Logf("Found error: %s - %s", ve.FilePath, ve.ErrorMsg)
		}
	} else {
		errPath := strings.ToLower(filepath.ToSlash(validationErrors[0].FilePath))
		expectedPathSub := strings.ToLower(filepath.ToSlash(filePath2))
		if !strings.Contains(errPath, expectedPathSub) {
			t.Errorf("expected validation error to contain path '%s', got '%s'", expectedPathSub, errPath)
		}
		if !strings.Contains(validationErrors[0].ErrorMsg, "missing required extra field 'module'") {
			t.Errorf("unexpected error message: %s", validationErrors[0].ErrorMsg)
		}
	}

	// 6. Fix TASK-01 by updating it
	_, err = UpdateTicket(filePath2, map[string]string{"module": "ui"})
	if err != nil {
		t.Fatalf("UpdateTicket failed: %v", err)
	}

	// Validate again: should pass now!
	validationErrors, err = ValidateTickets(tempDir)
	if err != nil {
		t.Fatalf("ValidateTickets failed: %v", err)
	}

	if len(validationErrors) != 0 {
		t.Errorf("expected 0 validation errors, got %d", len(validationErrors))
		for _, ve := range validationErrors {
			t.Logf("Unexpected error: %s - %s", ve.FilePath, ve.ErrorMsg)
		}
	}

	// 7. Test type check mismatch validation
	// Let's manually corrupt the type in BUG-01
	meta, body, err := ParseFile(filePath1)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	meta.Type = "not-bugs" // mismatch!
	if err := WriteFile(filePath1, meta, body); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Validate again: should fail because type is "not-bugs" instead of "bugs"
	validationErrors, err = ValidateTickets(tempDir)
	if err != nil {
		t.Fatalf("ValidateTickets failed: %v", err)
	}

	if len(validationErrors) != 1 {
		t.Errorf("expected 1 type mismatch validation error, got %d", len(validationErrors))
	} else {
		if !strings.Contains(validationErrors[0].ErrorMsg, "invalid ticket type") {
			t.Errorf("unexpected error message: %s", validationErrors[0].ErrorMsg)
		}
	}
}

func TestMigrateTickets(t *testing.T) {
	// 1. Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "ticket-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	bugsDir := filepath.Join(tempDir, "bugs")
	tasksDir := filepath.Join(tempDir, "tasks")
	if err := os.Mkdir(bugsDir, 0755); err != nil {
		t.Fatalf("failed to create bugs dir: %v", err)
	}
	if err := os.Mkdir(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Write ticket.yaml
	yamlConfig := `
sub_dirs:
  - bugs
  - tasks
`
	if err := os.WriteFile(filepath.Join(tempDir, "ticket.yaml"), []byte(yamlConfig), 0644); err != nil {
		t.Fatalf("failed to write ticket.yaml: %v", err)
	}

	// Create legacy files without frontmatter
	bugContent := `# BUG-01: Correct layout problem

## Details
Some bug details here.
`
	bugLegacyPath := filepath.Join(bugsDir, "BUG-01-pass.md")
	if err := os.WriteFile(bugLegacyPath, []byte(bugContent), 0644); err != nil {
		t.Fatalf("failed to write bug legacy file: %v", err)
	}

	taskContent := `# TASK-02: Setup pipeline

## Details
Some task details here.
`
	taskLegacyPath := filepath.Join(tasksDir, "TASK-02-failed.md")
	if err := os.WriteFile(taskLegacyPath, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task legacy file: %v", err)
	}

	// 2. Run MigrateTickets
	results, err := MigrateTickets(tempDir)
	if err != nil {
		t.Fatalf("MigrateTickets failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 migration results, got %d", len(results))
	}

	// Check that legacy files are gone and new files exist
	bugNewPath := filepath.Join(bugsDir, "BUG-01.md")
	taskNewPath := filepath.Join(tasksDir, "TASK-02.md")

	if _, err := os.Stat(bugLegacyPath); !os.IsNotExist(err) {
		t.Errorf("expected legacy BUG-01-pass.md to be deleted")
	}
	if _, err := os.Stat(taskLegacyPath); !os.IsNotExist(err) {
		t.Errorf("expected legacy TASK-02-failed.md to be deleted")
	}

	if _, err := os.Stat(bugNewPath); err != nil {
		t.Errorf("expected migrated BUG-01.md to exist")
	}
	if _, err := os.Stat(taskNewPath); err != nil {
		t.Errorf("expected migrated TASK-02.md to exist")
	}

	// 3. Verify BUG-01 contents
	metaBug, bodyBug, err := ParseFile(bugNewPath)
	if err != nil {
		t.Fatalf("failed to parse migrated BUG-01.md: %v", err)
	}
	if metaBug.ID != "BUG-01" {
		t.Errorf("BUG-01: expected ID 'BUG-01', got '%s'", metaBug.ID)
	}
	if metaBug.Status != "passed" {
		t.Errorf("BUG-01: expected status 'passed', got '%s'", metaBug.Status)
	}
	if metaBug.Title != "Correct layout problem" {
		t.Errorf("BUG-01: expected title 'Correct layout problem', got '%s'", metaBug.Title)
	}
	if metaBug.Type != "bugs" {
		t.Errorf("BUG-01: expected type 'bugs', got '%s'", metaBug.Type)
	}
	if metaBug.ResolvedAt == "" {
		t.Errorf("BUG-01: expected resolved_at to be populated for passed status")
	}
	if !strings.Contains(bodyBug, "## Details") {
		t.Errorf("BUG-01: body should retain original content, got: %s", bodyBug)
	}

	// 4. Verify TASK-02 contents
	metaTask, bodyTask, err := ParseFile(taskNewPath)
	if err != nil {
		t.Fatalf("failed to parse migrated TASK-02.md: %v", err)
	}
	if metaTask.ID != "TASK-02" {
		t.Errorf("TASK-02: expected ID 'TASK-02', got '%s'", metaTask.ID)
	}
	if metaTask.Status != "rejected" {
		t.Errorf("TASK-02: expected status 'rejected' (from failed suffix), got '%s'", metaTask.Status)
	}
	if metaTask.Title != "Setup pipeline" {
		t.Errorf("TASK-02: expected title 'Setup pipeline', got '%s'", metaTask.Title)
	}
	if metaTask.Type != "tasks" {
		t.Errorf("TASK-02: expected type 'tasks', got '%s'", metaTask.Type)
	}
	if metaTask.ResolvedAt == "" {
		t.Errorf("TASK-02: expected resolved_at to be populated for rejected status")
	}
	if !strings.Contains(bodyTask, "## Details") {
		t.Errorf("TASK-02: body should retain original content")
	}
}

func TestMigrateIdempotencyAndOptions(t *testing.T) {
	// 1. Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "ticket-migrate-opts-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	bugsDir := filepath.Join(tempDir, "bugs")
	if err := os.Mkdir(bugsDir, 0755); err != nil {
		t.Fatalf("failed to create bugs dir: %v", err)
	}

	// Write ticket.yaml config to allow recursion
	yamlConfig := `
sub_dirs:
  - bugs
`
	if err := os.WriteFile(filepath.Join(tempDir, "ticket.yaml"), []byte(yamlConfig), 0644); err != nil {
		t.Fatalf("failed to write ticket.yaml: %v", err)
	}

	// Create a standard valid ticket
	stdTicketContent := `---
id: BUG-001
title: Valid Ticket
type: bugs
status: resolved
priority: major
owner: dev-agent
created_at: 2026-06-11 10:00
updated_at: 2026-06-11 17:00
---
# Valid Ticket
`
	stdPath := filepath.Join(bugsDir, "BUG-001.md")
	if err := os.WriteFile(stdPath, []byte(stdTicketContent), 0644); err != nil {
		t.Fatalf("failed to write std ticket: %v", err)
	}

	// Create a legacy file needing migration
	legacyContent := `# BUG-002: Legacy Ticket

## Details
Legacy issue here.
`
	legacyPath := filepath.Join(bugsDir, "BUG-002-pass.md")
	if err := os.WriteFile(legacyPath, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("failed to write legacy ticket: %v", err)
	}

	// Record original modification time and hash of standard ticket
	stdInfoInit, err := os.Stat(stdPath)
	if err != nil {
		t.Fatalf("failed to stat std ticket: %v", err)
	}
	stdModTimeInit := stdInfoInit.ModTime()

	stdBytesInit, err := os.ReadFile(stdPath)
	if err != nil {
		t.Fatalf("failed to read std ticket: %v", err)
	}

	// TEST 1: Dry run migration
	dryResults, err := MigrateTicketsWithOptions(tempDir, MigrateOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	// Dry run shouldn't write BUG-002.md or delete BUG-002-pass.md
	if _, err := os.Stat(filepath.Join(bugsDir, "BUG-002.md")); !os.IsNotExist(err) {
		t.Errorf("BUG-002.md should not be written during dry run")
	}
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		t.Errorf("BUG-002-pass.md should not be deleted during dry run")
	}

	// Verify dry run results contain "renamed & updated" for BUG-002-pass.md
	foundDryLegacy := false
	for _, res := range dryResults {
		if strings.Contains(res.OriginalPath, "BUG-002-pass.md") {
			foundDryLegacy = true
			if res.Action != "renamed & updated" {
				t.Errorf("expected dry action 'renamed & updated', got '%s'", res.Action)
			}
		}
	}
	if !foundDryLegacy {
		t.Errorf("dry run did not return results for BUG-002-pass.md")
	}

	// TEST 2: Only Invalid migration (normal execution)
	optsOnlyInvalid := MigrateOptions{OnlyInvalid: true}
	resultsOnlyInvalid, err := MigrateTicketsWithOptions(tempDir, optsOnlyInvalid)
	if err != nil {
		t.Fatalf("only invalid migration failed: %v", err)
	}

	// Since BUG-001.md is valid, it should be skipped. BUG-002-pass.md is invalid, so it should be migrated.
	var found001, found002 bool
	for _, res := range resultsOnlyInvalid {
		if strings.Contains(res.OriginalPath, "BUG-001.md") {
			found001 = true
			if res.Action != "skipped" || res.Reason != "already_standardized" {
				t.Errorf("BUG-001: expected action 'skipped' with reason 'already_standardized', got action='%s', reason='%s'", res.Action, res.Reason)
			}
		}
		if strings.Contains(res.OriginalPath, "BUG-002-pass.md") {
			found002 = true
			if res.Action != "renamed & updated" {
				t.Errorf("BUG-002: expected action 'renamed & updated', got '%s'", res.Action)
			}
		}
	}
	if !found001 || !found002 {
		t.Errorf("expected both BUG-001 and BUG-002 in only-invalid results (BUG-001: %t, BUG-002: %t)", found001, found002)
	}

	// Check if BUG-001 was untouched
	stdInfoAfterOpts, err := os.Stat(stdPath)
	if err != nil {
		t.Fatalf("failed to stat BUG-001: %v", err)
	}
	if !stdInfoAfterOpts.ModTime().Equal(stdModTimeInit) {
		t.Errorf("BUG-001.md modification time changed during only-invalid migration")
	}
	stdBytesAfterOpts, err := os.ReadFile(stdPath)
	if err != nil {
		t.Fatalf("failed to read BUG-001: %v", err)
	}
	if string(stdBytesInit) != string(stdBytesAfterOpts) {
		t.Errorf("BUG-001.md content changed during only-invalid migration")
	}

	// BUG-002-pass.md should be migrated to BUG-002.md
	migratedPath := filepath.Join(bugsDir, "BUG-002.md")
	if _, err := os.Stat(migratedPath); os.IsNotExist(err) {
		t.Errorf("BUG-002.md should have been created")
	}

	// TEST 3: Idempotency (continuous normal migration on already migrated directory)
	// Get file modification time of BUG-002.md after first migration
	migInfo, err := os.Stat(migratedPath)
	if err != nil {
		t.Fatalf("failed to stat BUG-002: %v", err)
	}
	migModTime := migInfo.ModTime()
	migBytes, err := os.ReadFile(migratedPath)
	if err != nil {
		t.Fatalf("failed to read BUG-002: %v", err)
	}

	// Run migration again (idempotency check)
	resultsIdempotency2, err := MigrateTicketsWithOptions(tempDir, MigrateOptions{})
	if err != nil {
		t.Fatalf("third migration failed: %v", err)
	}

	// All files should be skipped as already standardized
	for _, res := range resultsIdempotency2 {
		if res.Action != "skipped" || res.Reason != "already_standardized" {
			t.Errorf("idempotency check: expected action 'skipped' with reason 'already_standardized' for %s, got action='%s', reason='%s'", res.OriginalPath, res.Action, res.Reason)
		}
	}

	// Verify BUG-002.md mod time and content remained unchanged
	migInfoAfter, err := os.Stat(migratedPath)
	if err != nil {
		t.Fatalf("failed to stat BUG-002: %v", err)
	}
	if !migInfoAfter.ModTime().Equal(migModTime) {
		t.Errorf("BUG-002.md modification time changed during second migration (violates minimal write)")
	}
	migBytesAfter, err := os.ReadFile(migratedPath)
	if err != nil {
		t.Fatalf("failed to read BUG-002: %v", err)
	}
	if string(migBytes) != string(migBytesAfter) {
		t.Errorf("BUG-002.md content changed during second migration")
	}

	// TEST 4: Single file migration target
	// Create another legacy file
	legacyContent3 := `# BUG-003: Legacy Ticket 3

## Details
Legacy issue here 3.
`
	legacyFile3 := filepath.Join(bugsDir, "BUG-003-pass.md")
	if err := os.WriteFile(legacyFile3, []byte(legacyContent3), 0644); err != nil {
		t.Fatalf("failed to write BUG-003-pass.md: %v", err)
	}

	singleResults, err := MigrateTicketsWithOptions(legacyFile3, MigrateOptions{})
	if err != nil {
		t.Fatalf("single file migration failed: %v", err)
	}

	if len(singleResults) != 1 {
		t.Errorf("expected 1 result from single file migration, got %d", len(singleResults))
	} else if singleResults[0].Action != "renamed & updated" {
		t.Errorf("expected action 'renamed & updated' for single file migration, got '%s'", singleResults[0].Action)
	}

	if _, err := os.Stat(filepath.Join(bugsDir, "BUG-003.md")); os.IsNotExist(err) {
		t.Errorf("single file migration target BUG-003.md should be created")
	}
}

