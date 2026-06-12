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

## 🛠️ Installation & Usage

### 1. Build
Build the executable binary inside the project root using PowerShell 7 or make:
```powershell
# Windows PowerShell
./make.ps1 build

# Unix / macOS
make build
```
This produces the executable `ticket` (or `ticket.exe` on Windows) under `plugin/commands/`.

### 2. Integration with AI Agent CLIs

Choose the integration method based on your AI toolchain:

#### A. As a Google Antigravity CLI (`agy`) Plugin
Install using the `agy` plugin system by specifying the local plugin path:
```bash
# Replace /path/to/ with the actual absolute path to this repo on your machine
agy plugin install /path/to/ticket-cli-plugin/plugin
```
Once installed:
- `agy` automatically adds the `commands/` directory to the shell session `PATH` for the Agent.
- `agy` automatically loads and activates the skill defined in `skills/ticket-management/SKILL.md`.

#### B. As a Claude Code (Anthropic) Plugin
Claude Code natively supports loading plugins via `.claude-plugin/plugin.json`:
1. **Load via command line**:
   Start Claude Code pointing to the local plugin directory:
   ```bash
   claude --plugin-dir /path/to/ticket-cli-plugin/plugin
   ```
2. **Add inside an active session**:
   Add the plugin directly in a running interactive session:
   ```bash
   /plugin add /path/to/ticket-cli-plugin/plugin
   ```

#### C. As an OpenAI Codex CLI Plugin

Codex CLI features a native plugin and marketplace manager. We recommend installing it using one of the following methods:

##### Method 1: Direct Install from GitHub Repo (Recommended, Self-Bootstrapping)
This is the easiest installation method. The plugin includes a lightweight cross-platform JS bootstrap script that automatically downloads the pre-compiled binary matching your platform architecture from GitHub Releases on first run:
1. **Register the Git repository as a marketplace**:
   ```bash
   codex plugin marketplace add https://github.com/deepziyu/ticket-cli-plugin.git
   ```
2. **Install and enable the plugin**:
   ```bash
   codex plugin add ticket-management-plugin@ticket-management-marketplace
   ```

##### Method 2: Local Source Build & Install (For Developer Debugging)
If you want to modify or compile the Go source code yourself:
1. **Build the binary locally**:
   Build the executable in the project root:
   ```powershell
   # Windows
   ./make.ps1 build

   # macOS / Linux
   make build
   ```
   This generates the platform-specific binary under `plugin/commands/`.
2. **Register the local project root directory as a marketplace**:
   ```bash
   # Replace /path/to/ with the actual absolute path to the repo root folder
   codex plugin marketplace add /path/to/ticket-cli-plugin
   ```
3. **Install and enable the plugin**:
   ```bash
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
Scans a folder, parses legacy state suffixes from filenames (e.g., `-pass`, `-passed`, `-fail`), writes them to the Frontmatter `status` field, and renames files to achieve name immutability.
```bash
ticket migrate ./bugs
```
If `--format json` is set, a JSON array detailing every migration action is returned.

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
