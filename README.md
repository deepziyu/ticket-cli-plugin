# ticket-cli-plugin (Markdown Ticket & Bug Management CLI)

[中文版](./README_ZH.md)

This plugin is designed for collaborative environments combining **AI Agents + Human Software Engineers**. It is a Go-based CLI tool to parse and manage ticket/bug reports stored as Markdown files locally, and packaged natively as a Google Antigravity CLI (`agy`) plugin.

## 🎯 Background & Pain Points

In AI-driven development loops (e.g. auto test -> auto fix -> auto verify), AI agents frequently read and write local markdown ticket files. This often leads to:
1. **High File Traversal Cost**: Traversing dozens of ticket files locally requires multiple `view_file` calls from the Agent, wasting API tokens and causing delays of up to several minutes.
2. **Fragile Text Modification**: Replacing YAML frontmatter or markdown content using regex-based tools like `replace_file_content` easily fails due to invisible characters or minor format deviations.
3. **Broken Links due to Renaming**: Standard renaming workflows (e.g., appending `-pass.md`) break file names, disrupting Git histories and breaking Markdown or TAPD links pointing to the original ticket.

This tool resolves these issues through **"Immutable File Names + Status Transitions via Frontmatter + High-performance CLI Gateway"**.

---

## 🛠️ Installation & Setup Guide

This plugin supports both **"Direct Installation (Recommended, Self-Bootstrapping)"** and **"Local Source Build"** methods.

---

### 1. 🚀 Direct Installation (For Users, Recommended)

This mode does not require a local Go language environment. The plugin comes with a cross-platform bootstrap script (`ticket.js`) that automatically downloads the pre-compiled binary matching your system architecture from GitHub Releases when an AI tool runs the `ticket` command for the first time.

#### A. One-click Installation in OpenAI Codex CLI
1. **Register the remote GitHub marketplace**:
   ```bash
   codex plugin marketplace add https://github.com/deepziyu/ticket-cli-plugin.git
   ```
2. **Install and enable the plugin**:
   ```bash
   codex plugin add ticket-management-plugin@ticket-management-marketplace
   ```

#### B. Enable in Google Antigravity CLI (`agy`)
If you have installed via `codex` above, you can directly link its cached directory:
```bash
# Windows
agy plugin install "$HOME\.codex\plugins\cache\ticket-management-marketplace\ticket-management-plugin\1.0.1"

# Unix / macOS
agy plugin install ~/.codex/plugins/cache/ticket-management-marketplace/ticket-management-plugin/1.0.1
```
*Or clone the repository and install directly:*
```bash
git clone https://github.com/deepziyu/ticket-cli-plugin.git
agy plugin install ./ticket-cli-plugin/plugin
```

#### C. Enable in Claude Code (Anthropic)
If you have installed via `codex` above, link its cache directory to start:
```bash
# Windows (PowerShell)
claude --plugin-dir "$HOME\.codex\plugins\cache\ticket-management-marketplace\ticket-management-plugin\1.0.1"

# Unix / macOS
claude --plugin-dir ~/.codex/plugins/cache/ticket-management-marketplace/ticket-management-plugin/1.0.1
```
*Or clone the repository and install directly:*
```bash
git clone https://github.com/deepziyu/ticket-cli-plugin.git
claude --plugin-dir ./ticket-cli-plugin/plugin
```

---

### 2. 🛠️ Local Source Build & Installation (For Contributors & Developers)

If you need to customize the Go source code or debug the plugin locally:

1. **Clone the repository and enter the root directory**:
   ```bash
   git clone https://github.com/deepziyu/ticket-cli-plugin.git
   cd ticket-cli-plugin
   ```
2. **Build the binary matching your host OS**:
   ```powershell
   # Windows (PowerShell 7)
   ./make.ps1 build

   # Linux / macOS
   make build
   ```
   This generates the executable `ticket` (or `ticket.exe`) under `plugin/commands/`.
3. **Link the local plugin directory in your CLI of choice**:
   - **Google Antigravity CLI (`agy`)**:
     ```bash
     agy plugin install ./plugin
     ```
   - **Claude Code**:
     ```bash
     claude --plugin-dir ./plugin
     ```
   - **OpenAI Codex CLI**:
     Register the local directory as a marketplace and install:
     ```bash
     codex plugin marketplace add ./
     codex plugin add ticket-management-plugin@ticket-cli-plugin
     ```


---

## 📖 CLI Commands (For Humans & AI Agents)

Run the `ticket` command inside any enabled workspace:

### 1. List Tickets (`ticket list`)
*   **Table view (Human-friendly)**:
    ```bash
    ticket list ./bugs
    ```
    Output:
    ```text
    ID                 |  STATUS    |  PRIORITY  |  OWNER      |  UPDATED AT        |  TITLE
    -------------------------------------------------------------------------------------------------------------------------
    BUG-S1-01-20260610  |  REJECTED  |  CRITICAL  |             |  2026-06-11 11:45  |  Fill all template fields and publish
    BUG-S1-03-20260611  |  PASSED    |  MAJOR     |             |  2026-06-11 11:45  |  Move published template back to draft
    BUG-S1-04-20260611  |  PASSED    |  MAJOR     |  dev-agent  |  2026-06-11 11:45  |  Save draft and publish new version
    ```

*   **JSON output (Agent-optimized)**:
    ```bash
    ticket list ./bugs --format json --status open
    ```
    Returns all matching tickets in a single lightweight JSON payload without markdown body noises.

### 2. Create a Ticket (`ticket create`)
```bash
ticket create ./bugs BUG-S1-05-20260611 --title "SOP L0 Validation 500 Server Error" --type bug --status open --priority major
```

### 3. Transition & Update Attributes (`ticket update`)
Update status, owner, priority, conclusion, or append custom metadata via `--field`. Flags can be placed after the target file path:
```bash
ticket update ./bugs/BUG-S1-04-20260611.md --status passed --conclusion "Reused existing router dispatch to fix" --field test_data_notes="draft_v1"
```
*Behind the scenes, the tool will:*
- Analyze the metadata changes and display a diff. If `--format json` is set, the full updated ticket structure is returned in JSON.
- Automatically refresh the `updated_at` timestamp.
- Auto-populate `resolved_at` when entering final states (`passed`/`rejected`), and clear it when moving back to open states (e.g. `open`).

### 4. Show Details (`ticket show`)
```bash
# JSON output containing the metadata and full markdown body content
ticket show ./bugs/BUG-S1-04-20260611.md --format json

# Human-readable formatted console layout
ticket show ./bugs/BUG-S1-04-20260611.md
```

### 5. Validate Ticket Schema (`ticket validate`)
Validates that all markdown files in the folder contain valid YAML Frontmatter headers, and checks if status and priority values match standard enums:
```bash
ticket validate ./bugs
```
If `--format json` is provided and verification passes, it returns a blank `[]` array.

### 6. Migrate Legacy Tickets (`ticket migrate`)
Scans a directory recursively, or targets a single file path directly. It parses legacy state suffixes from filenames (e.g., `-pass`, `-passed`, `-fail`), transitions them to the Frontmatter `status` field, and renames files to achieve name immutability.
```bash
# Migrate a directory or a single file
ticket migrate ./bugs
ticket migrate ./bugs/BUG-001-pass.md

# Preview changes without modifying files on disk
ticket migrate ./bugs --dry-run

# Only migrate ticket files that fail schema validation
ticket migrate ./bugs --only-invalid
```
*The migrate command guarantees the following key features:*
- **Idempotency & Minimal Write**: If the file is already standardized and has no metadata changes, it will be skipped entirely. No disk write is performed, and file modification times remain unchanged to prevent git workspace clutter.
- **Rich Output Options**: When `--format json` is provided, a detailed JSON array is returned, showing `file`, `action`, `reason`, and `changed_fields` for each file. On console output, skipped files display cleanly as `[skipped: already standardized]`.

---

## ⚙️ Custom Configurations (`ticket.yaml`)

Define custom subdirectories and metadata schemas by placing a `ticket.yaml` at your tickets root directory:

```yaml
# ticket.yaml
# 1. Subdirectories to recursively scan (scans only current dir if omitted)
sub_dirs:
  - bugs
  - tasks

# 2. Extra metadata fields validation rules
extra_fields:
  - name: module        # Field key
    required: true      # Triggers validation errors if missing or empty
  - name: severity
    required: false
```

### Validation Mechanisms:
1. **Category is tied to Folder Names**: The ticket `type` property is strictly validated against the lowercase folder name. The CLI automatically updates or corrects this field when running `ticket create` or `ticket update`.
2. **Inheritance**: Subdirectories inherit custom fields validations (`extra_fields`) from their parents unless a nested `ticket.yaml` is defined. `sub_dirs` configurations do not inherit.

---

## 💎 Standard YAML Frontmatter Schema

Markdown tickets must start with a valid YAML Frontmatter block:
```yaml
---
id: BUG-S1-04-20260611              # Unique ticket ID
title: Save draft and publish version  # Short title
type: bug                           # Type (bug/task/feature)
status: passed                      # Status (open/fixing/resolved/passed/rejected)
priority: major                     # Priority (critical/major/minor/low)
owner: dev-agent                    # Assigned developer or Agent
created_at: 2026-06-11 10:24        # Creation timestamp (YYYY-MM-DD HH:MM)
updated_at: 2026-06-11 11:45        # Modification timestamp (YYYY-MM-DD HH:MM)
resolved_at: 2026-06-11 11:45       # Time resolved (populated only for final statuses)
conclusion: Reused dispatch router   # Resolution/conclusion details
---

# BUG-S1-04-20260611

## Steps & Actual Results
... markdown body content here ...
```
*(Note: Custom metadata fields defined during dev cycles will be fully preserved in frontmatter headers by the parser.)*
