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

