---
name: ticket-management
description: Manage, query, transition, and validate ticket-like Markdown files (e.g. bugs, tasks) using the 'ticket' CLI command.
---

# 🎫 Ticket & Bug Management (Ticket Management) Skill

[Chinese Version](./SKILL_ZH.md)

This skill is designed for AI Agents and human developers, providing a standard protocol to retrieve, transition, and validate local Markdown-based tickets, tasks, or bug files using the `ticket` CLI command.

By running the `ticket` command instead of manually reading and modifying files, the Agent avoids high directory traversal costs (reducing Token consumption) and prevents regex match errors during text modifications.

---

## 🎯 Activation Conditions

The Agent should activate this skill when encountering the following tasks:
1. **Get Ticket or Bug Lists**: When needing to know what tasks or bug tickets exist in a given folder (and filtering by status/priority).
2. **Update Ticket Status/Attributes**: When transitioning a ticket status (e.g., from `open` to `fixing` or `passed`) or adding resolution conclusions.
3. **Create a Ticket**: When a test fails or a new bug is detected, and a ticket file needs to be generated.
4. **Validate Ticket Syntax**: When checking if the Markdown files in a directory comply with YAML Frontmatter standards and schemas.

---

## 🛠️ CLI Command Reference

Always run the `ticket` command inside the environment instead of editing file frontmatter headers directly.

> [!TIP]
> **💡 Self-Healing Fallback**: If running `ticket` directly throws a "CommandNotFound" error:
> 1. Search the current Workspace or global plugin directories (e.g. `~/.codex/plugins/cache/`) for the `ticket` executable (such as `ticket.exe` / `ticket.cmd` on Windows, or `ticket` on Unix).
> 2. Use the **absolute or relative path** of the executable to run the command (e.g. `./plugin/commands/ticket list ./bugs`), instead of falling back to editing markdown headers manually.

### 1. List Tickets (`ticket list`)
*   **Syntax**: `ticket list [flags] [dir]`
*   **AI Recommendation**: Always use `--format json` to get a structured, compact JSON response.
    ```bash
    ticket list --format json --status open ./tests/02-cases/v0.1/bugs
    ```
*   **Filter Flags**:
    *   `--status`: Filter by status (`open` / `fixing` / `resolved` / `passed` / `rejected`)
    *   `--priority`: Filter by priority (`critical` / `major` / `minor` / `low`)
    *   `--owner`: Filter by owner
    *   `--type`: Filter by type (`bug` / `task` / `feature`)
    *   `--format`: Output formatting (`table` for humans / `json` for agents)

### 2. Update Ticket Attributes (`ticket update`)
*   **Syntax**: `ticket update [flags] <file_path>` (Note: flags can be placed after the file path for convenience, e.g., `ticket update ./bugs/BUG-01.md --status passed`)
*   **Flags**:
    *   `--status <status>`: Transition the state (moving to `passed` or `rejected` automatically sets the `resolved_at` time; moving back to `open` clears `resolved_at`).
    *   `--owner <owner>`: Assign to a developer or agent.
    *   `--priority <priority>`: Adjust priority levels.
    *   `--conclusion "<text>"`: Write details on how the ticket was resolved.
    *   `--field key=value`: Set or update custom YAML frontmatter fields (e.g. `--field module=auth --field bug_source=e2e`).
    *   `--format <text|json>`: Output format. `json` returns the full updated ticket structure, while `text` (default) displays a formatted diff of changed fields.
*   **Example**:
    ```bash
    ticket update --status passed --conclusion "Reused existing router configuration to fix" ./bugs/BUG-S1-03-20260611.md
    ```

### 3. Create a Ticket (`ticket create`)
*   **Syntax**: `ticket create [flags] <dir_path> <id>`
*   **Example**:
    ```bash
    ticket create --title "SOP L0 Validation returns 500 error" --type bug --status open --priority major --owner dev-agent ./bugs BUG-S1-05-20260611
    ```
*   **Flags**:
    *   `--title`: Short title
    *   `--type`: Ticket type (bug, task, feature)
    *   `--status`: Initial status
    *   `--priority`: Priority
    *   `--owner`: Assigned owner
    *   `--body`: Initial markdown body
    *   `--field key=value`: Add custom frontmatter properties

### 4. Show Ticket Details (`ticket show`)
*   **Syntax**: `ticket show [flags] <file_path>`
*   **Flags**:
    *   `--format json`: Returns a single JSON containing all metadata and the full markdown body.
    *   `--format text`: Outputs a clean human-readable text view.
*   **Example**:
    ```bash
    ticket show --format json ./bugs/BUG-S1-03-20260611.md
    ```

### 5. Validate Directory Schema (`ticket validate`)
*   **Syntax**: `ticket validate [flags] [dir_path]`
*   **Flags**:
    *   `--format <text|json>`: Output formatting. Returns `[]` if no errors found under `json`.
*   **Example**:
    ```bash
    ticket validate --format json ./bugs
    ```

### 6. Migrate Legacy Tickets (`ticket migrate`)
*   **Syntax**: `ticket migrate [flags] [file_or_dir_path]`
*   **Description**: Scans a directory or targets a single file path for migration. It parses legacy state suffixes from filenames (e.g., `-pass`, `-passed`, `-fail`), transitions them to the frontmatter `status` field, and renames files to remove suffixes to achieve name immutability.
    *   **Idempotency & Minimal Write**: If the file is already standardized and has no metadata changes, it will be skipped entirely. No disk write is performed, and file modification times remain unchanged to prevent git workspace clutter.
*   **Flags**:
    *   `--dry-run`: Preview changes without modifying files on disk.
    *   `--only-invalid`: Only migrate ticket files that fail schema validation.
    *   `--format <text|json>`: Output format. Returns a detailed JSON array showing `file`, `action`, `reason`, and `changed_fields` when `json` is specified.
*   **Example**:
    ```bash
    # Migrate a directory
    ticket migrate ./bugs
    # Preview migration
    ticket migrate --dry-run ./bugs
    ```

### 7. Run Kanban Web Dashboard (`ticket dashboard`)
*   **Syntax**: `ticket dashboard [flags] [dir_path]`
*   **Description**: Launch a lightweight, aesthetic dark-mode Kanban board UI in your browser (default at `http://localhost:8080`). Supports drag-and-drop status transitions, adding/removing tickets, and live markdown editing/previewing.
*   **Flags**:
    *   `--port` / `-p`: Listen port (default: 8080)
*   **Example**:
    ```bash
    ticket dashboard --port 18080 ./bugs
    ```

---

## ⚙️ Custom Configurations (`ticket.yaml`)

Define custom subdirectories and metadata schemas by placing a `ticket.yaml` at your tickets root directory:

### 1. Configuration Example
```yaml
# ticket.yaml
sub_dirs:
  - bugs
  - tasks

extra_fields:
  - name: module
    required: true
  - name: severity
```

### 2. Validation Rules
1. **Explicit Subdirectory Scanning**: If `sub_dirs` is defined, the tool scans only those listed directories. If omitted, no recursive scanning is performed.
2. **Category is strictly bound to Type**: Frontmatter `type` is automatically forced to match the lowercase folder name. The CLI overrides this field dynamically during `create`/`update` calls.
3. **Required Field Verification**: `ticket validate` verifies that all extra fields marked as `required: true` are present and non-empty. Subdirectories inherit these validation schemas unless they have their own `ticket.yaml`.

---

## 🚫 Limitations & Notes

1. **Avoid Manual Frontmatter Editing**: The Agent must use CLI commands to change status or metadata. This prevents YAML formatting issues and typos.
2. **File Name Immutability**: Filenames must remain unchanged once created. All state transitions must be handled by updating the `status` field inside the frontmatter.
3. **Keep Type Synced**: Make sure to place tickets under correct physical directories so that the auto-aligned `type` remains correct.
