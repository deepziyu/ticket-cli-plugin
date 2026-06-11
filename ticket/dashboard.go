package ticket

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed web/*
var webFS embed.FS

// Global workspace configuration for the running dashboard
var activeWorkspace string

// StartDashboard launches the web server and serves the Kanban UI.
func StartDashboard(dir string, port int) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}
	activeWorkspace = absDir

	// Ensure the directory exists
	info, err := os.Stat(activeWorkspace)
	if err != nil {
		return fmt.Errorf("directory not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", activeWorkspace)
	}

	// Sub-FS for web frontend static files
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to locate embedded web files: %w", err)
	}

	// Register Handlers
	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/api/config", handleGetConfig)
	http.HandleFunc("/api/tickets", handleGetTickets)
	http.HandleFunc("/api/tickets/detail", handleGetTicketDetail)
	http.HandleFunc("/api/tickets/create", handleCreateTicket)
	http.HandleFunc("/api/tickets/update", handleUpdateTicket)
	http.HandleFunc("/api/tickets/move", handleMoveTicket)
	http.HandleFunc("/api/tickets/delete", handleDeleteTicket)

	url := "http://localhost:" + strconv.Itoa(port)
	fmt.Printf("\n---------------------------------------------------\n")
	fmt.Printf("🚀 Ticket Dashboard Server is running at: %s\n", url)
	fmt.Printf("   Workspace: %s\n", activeWorkspace)
	fmt.Printf("   Press Ctrl+C to stop the server\n")
	fmt.Printf("---------------------------------------------------\n\n")

	// Trigger browser auto-open in a background goroutine
	go func() {
		time.Sleep(150 * time.Millisecond)
		openBrowser(url)
	}()

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// ==========================================================================
// Security & Path Resolution helpers
// ==========================================================================

func checkSafePath(targetPath string) (string, error) {
	// If path is absolute, check containment. If relative, join first.
	var fullPath string
	if filepath.IsAbs(targetPath) {
		fullPath = filepath.Clean(targetPath)
	} else {
		fullPath = filepath.Clean(filepath.Join(activeWorkspace, targetPath))
	}

	rel, err := filepath.Rel(activeWorkspace, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("access denied: path escapes the active workspace")
	}

	return fullPath, nil
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// ==========================================================================
// Web Controllers / Router Handlers
// ==========================================================================

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	config, err := LoadConfig(activeWorkspace)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func handleGetTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	tickets, err := ListTickets(activeWorkspace, "", "", "", "")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Normalize file paths relative to activeWorkspace before sending to UI
	for i := range tickets {
		rel, err := filepath.Rel(activeWorkspace, tickets[i].FilePath)
		if err == nil {
			tickets[i].FilePath = filepath.ToSlash(rel)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func handleGetTicketDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing path parameter")
		return
	}

	fullPath, err := checkSafePath(relPath)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	meta, body, err := ParseFile(fullPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse file: %v", err))
		return
	}

	type TicketDetailResponse struct {
		Meta     *TicketMeta `json:"meta"`
		FilePath string      `json:"file_path"`
		FileName string      `json:"file_name"`
		Body     string      `json:"body"`
	}

	resp := TicketDetailResponse{
		Meta:     meta,
		FilePath: filepath.ToSlash(relPath),
		FileName: filepath.Base(fullPath),
		Body:     body,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type CreateTicketRequest struct {
	Dir         string            `json:"dir"`
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	Owner       string            `json:"owner"`
	Body        string            `json:"body"`
	ExtraFields map[string]string `json:"extra_fields"`
}

func handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.ID == "" || req.Title == "" || req.Dir == "" {
		writeJSONError(w, http.StatusBadRequest, "ID, Title, and Dir are required")
		return
	}

	targetDir := filepath.Join(activeWorkspace, req.Dir)
	safeDir, err := checkSafePath(targetDir)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	// Verify target subdirectory exists
	if _, err := os.Stat(safeDir); os.IsNotExist(err) {
		// Auto-create directory if config lists it in sub_dirs
		err = os.MkdirAll(safeDir, 0755)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create subdirectory: %v", err))
			return
		}
	}

	createdPath, err := CreateTicket(safeDir, req.ID, req.Title, "", req.Status, req.Priority, req.Owner, req.ExtraFields, req.Body)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rel, err := filepath.Rel(activeWorkspace, createdPath)
	if err != nil {
		rel = createdPath
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"file_path": filepath.ToSlash(rel),
	})
}

type UpdateTicketRequest struct {
	FilePath string            `json:"file_path"`
	Updates  map[string]string `json:"updates"`
	Body     string            `json:"body"`
}

func handleUpdateTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	fullPath, err := checkSafePath(req.FilePath)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	meta, err := UpdateTicketWithBody(fullPath, req.Updates, req.Body)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"meta":   meta,
	})
}

type MoveTicketRequest struct {
	FilePath string `json:"file_path"`
	Status   string `json:"status"`
}

func handleMoveTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req MoveTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	fullPath, err := checkSafePath(req.FilePath)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	// Just update status (retains body automatically)
	meta, err := UpdateTicket(fullPath, map[string]string{"status": req.Status})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"meta":   meta,
	})
}

type DeleteTicketRequest struct {
	FilePath string `json:"file_path"`
}

func handleDeleteTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DeleteTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	fullPath, err := checkSafePath(req.FilePath)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	// Perform physical file removal
	if err := os.Remove(fullPath); err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove file: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// Helper to open browser URL on different host systems
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("⚠️ Auto-open browser failed. Please navigate to: %s\n", url)
	}
}
