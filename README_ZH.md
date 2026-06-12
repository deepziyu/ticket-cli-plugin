# ticket-cli-plugin (工单/缺陷 Markdown 管理工具插件)

[English Version](./README.md)

本规范专为 **AI Agent + 研发工程师** 协同开发环境设计，是一个基于 Go 语言编译的通用工单/缺陷 Markdown 解析与管理 CLI 工具，并被无缝打包为 Google Antigravity CLI (`agy`) 的插件。

## 🎯 背景与痛点

在 AI 驱动的研发循环（如自动化测试 -> 自动修复 -> 自动化验证）中，Agent 常需对本地的工单或 Bug 报告（.md 文件）进行大量读写。这会带来以下痛点：
1. **遍历开销大**：如果目录下有数十个工单文件，Agent 为了获取未解决工单的列表，必须发起数十次 `view_file` 调用，严重消耗 API 配额并导致数十秒甚至数分钟的延迟。
2. **文本修改极易失败**：Agent 在使用 `replace_file_content` 替换 YAML frontmatter 或正文时，极易因格式微调或不可见字符而发生 Match Failure。
3. **不可变文件名诉求**：传统的“修改文件名”（如重命名为 `*-pass.md`）会打断 Git 历史并使指向该工单的引用链接（Markdown 链接/TAPD 引用）全部断开。

本工具通过 **「文件名 Immutable（不变） + 元数据 frontmatter 状态流转 + 极速 CLI 网关」** 的方式完美解决了上述痛点。

---

## 🛠️ 安装与使用指南

本工具支持 **「直接线上安装（推荐，自动自举下载）」** 和 **「本地源码编译安装」** 两种方式。

---

### 1. 🚀 直接线上安装（适合使用者，推荐）

此方式不需要您在本地配置 Go 语言环境。本插件内置了跨平台自举脚本（`ticket.js`），在您首次通过 AI 工具执行 `ticket` 命令时，它会自动从 GitHub Releases 线上下载适合您 CPU 和操作系统架构的预编译二进制包。

#### A. 在 OpenAI Codex CLI 中一键安装（最佳体验）
1. **添加远端 GitHub 市场**：
   ```bash
   codex plugin marketplace add https://github.com/deepziyu/ticket-cli-plugin.git
   ```
2. **安装并启用插件**：
   ```bash
   codex plugin add ticket-management-plugin@ticket-management-marketplace
   ```

#### B. 在 Google Antigravity CLI (`agy`) 中启用
如果您已通过上方 `codex` 安装，可直接关联其缓存目录（无需重复下载）：
```bash
# Windows
agy plugin install "$HOME\.codex\plugins\cache\ticket-management-marketplace\ticket-management-plugin\1.0.1"

# Unix / macOS
agy plugin install ~/.codex/plugins/cache/ticket-management-marketplace/ticket-management-plugin/1.0.1
```
*或者直接通过 Git 克隆仓库挂载：*
```bash
git clone https://github.com/deepziyu/ticket-cli-plugin.git
agy plugin install ./ticket-cli-plugin/plugin
```

#### C. 在 Claude Code (Anthropic) 中启用
如果您已通过上方 `codex` 安装，可直接关联其缓存目录运行：
```bash
# Windows (PowerShell)
claude --plugin-dir "$HOME\.codex\plugins\cache\ticket-management-marketplace\ticket-management-plugin\1.0.1"

# Unix / macOS
claude --plugin-dir ~/.codex/plugins/cache/ticket-management-marketplace/ticket-management-plugin/1.0.1
```
*或者直接通过 Git 克隆仓库挂载：*
```bash
git clone https://github.com/deepziyu/ticket-cli-plugin.git
claude --plugin-dir ./ticket-cli-plugin/plugin
```

---

### 2. 🛠️ 本地源码编译安装（适合二次开发者）

如果您需要修改本插件的 Go 语言源码或进行本地调试：

1. **克隆项目并进入根目录**：
   ```bash
   git clone https://github.com/deepziyu/ticket-cli-plugin.git
   cd ticket-cli-plugin
   ```
2. **本地编译当前平台二进制文件**：
   ```powershell
   # Windows (PowerShell 7)
   ./make.ps1 build

   # Linux / macOS
   make build
   ```
   这会在 `plugin/commands/` 目录下生成 `ticket` 或 `ticket.exe` 可执行文件。
3. **在各 CLI 中挂载本地插件目录**：
   - **Google Antigravity CLI (`agy`)**：
     ```bash
     agy plugin install ./plugin
     ```
   - **Claude Code**：
     ```bash
     claude --plugin-dir ./plugin
     ```
   - **OpenAI Codex CLI**：
     将本地目录注册为市场并安装：
     ```bash
     codex plugin marketplace add ./
     codex plugin add ticket-management-plugin@ticket-cli-plugin
     ```


---

## 📖 CLI 命令行指南 (人类与 AI 通用)

在任何已启用该插件的 Workspace 中，你可以直接运行 `ticket` 命令：

### 1. 检索工单列表 (`ticket list`)
*   **列表输出（面向人类）**：
    ```bash
    ticket list ./bugs
    ```
    输出示例：
    ```text
    ID                 |  STATUS    |  PRIORITY  |  OWNER      |  UPDATED AT        |  TITLE
    -------------------------------------------------------------------------------------------------------------------------
    BUG-S1-01-20260610  |  REJECTED  |  CRITICAL  |             |  2026-06-11 11:45  |  新建模板完整填写所有字段并发布
    BUG-S1-03-20260611  |  PASSED    |  MAJOR     |             |  2026-06-11 11:45  |  编辑已发布模板进入草稿态
    BUG-S1-04-20260611  |  PASSED    |  MAJOR     |  dev-agent  |  2026-06-11 11:45  |  存为草稿后再发布新版本
    ```

*   **JSON 输出（面向 AI Agent）**：
    ```bash
    ticket list ./bugs --format json --status open
    ```
    在一词工具调用内极速返回所有匹配工单的 JSON 数据，没有正文噪音。

### 2. 创建新工单 (`ticket create`)
```bash
ticket create ./bugs BUG-S1-05-20260611 --title "SOP 验收规则 L0 审核 500 报错" --type bug --status open --priority major
```

### 3. 流转与更新属性 (`ticket update`)
支持更新 status, owner, priority, conclusion 等字段，或通过 `--field` 更新自定义 YAML frontmatter 键值对。参数（例如 flags）也可以置于文件路径之后，如：
```bash
ticket update ./bugs/BUG-S1-04-20260611.md --status passed --conclusion "复用现有派发配置修复" --field test_data_notes="现场为生效v0草稿v1"
```
*更新工单时，本工具会：*
- 自动分析修改前后的元数据，并输出已修改的字段清单。若指定了 `--format json`，将返回更新后完整工单数据的 JSON。
- 自动更新 `updated_at` 字段为当前时间。
- 在状态流转到 `passed` 或 `rejected` 终态时，自动填充 `resolved_at` 时间；如果从终态流转回 `open` 等状态，则会自动清空 `resolved_at`。

### 4. 查看工单详情 (`ticket show`)
```bash
# 以 JSON 格式输出详情（含 markdown body 字段）
ticket show ./bugs/BUG-S1-04-20260611.md --format json

# 以控制台友好的排版输出
ticket show ./bugs/BUG-S1-04-20260611.md
```

### 5. 校验工单规范 (`ticket validate`)
检查目录下所有 `.md` 文件是否包含了符合 YAML 规范 of Frontmatter 头，并校验状态值和优先级是否在标准枚举内：
```bash
ticket validate ./bugs
```
若指定了 `--format json` 且校验通过，将返回 `[]` 数组。

### 6. 批量迁移旧工单 (`ticket migrate`)
扫描指定目录下所有工单文件，或直接指定单个工单文件路径。解析文件名中的 legacy 状态后缀（如 `-pass`、`-passed`、`-fail` 等）写入 Frontmatter `status` 字段中，并自动重命名文件，补齐缺失的元数据（如 `title`、`type`、时间字段等）。
```bash
# 扫描目录或单个文件进行迁移
ticket migrate ./bugs
ticket migrate ./bugs/BUG-001-pass.md

# 仅预览迁移计划而不修改磁盘
ticket migrate ./bugs --dry-run

# 仅迁移在校验中为无效（即不符合规范）的工单文件
ticket migrate ./bugs --only-invalid
```
*本工具的迁移保证以下两个优秀特性：*
- **幂等性与最小写入**：如果工单已标准化且无任何内容或状态变化，它在二次运行时将被完全跳过（Skipped），不触发任何磁盘写回，亦不会改变文件的修改时间，避免污染 Git 工作区。
- **丰富的输出格式**：若指定了 `--format json`，将返回每个文件迁移状态的 JSON 报告，包含 `file`、`action`、`reason`、`changed_fields` 等属性。若在控制台输出，跳过的文件会清晰打印为 `[skipped: already standardized]`。

---

## ⚙️ 自定义配置 (`ticket.yaml`)

本工具支持在工单管理根目录（以及子目录）下创建 `ticket.yaml` 进行自定义属性与分类层级校验：

```yaml
# ticket.yaml
# 1. 递归扫描的子目录配置（若不配置 sub_dirs，默认不递归扫描任何子目录）
sub_dirs:
  - bugs
  - tasks

# 2. 额外自定义字段配置
extra_fields:
  - name: module        # 字段名称
    required: true      # 在 validate 时强制校验是否存在且非空
  - name: severity
    required: false
```

### 核心校验机制：
1. **分类 (Category) 即是目录名**：工单的 `type` 属性必须与其所在的物理文件夹目录名（转换为小写）完全一致。在调用 `ticket create` 或 `ticket update` 时，工具会根据当前所在的路径自动修正/覆盖 `type` 字段的值。
2. **继承规则**：如果子目录中没有独立的 `ticket.yaml` 配置文件，它将默认继承父级文件夹配置的 `extra_fields` 规则，而 `sub_dirs` 不会继承。

---

## 💎 标准 YAML Frontmatter Schema 规范

Markdown 工单文件顶部必须包含 `---` 包含的 YAML Frontmatter：
```yaml
---
id: BUG-S1-04-20260611              # 工单/缺陷唯一ID
title: 存为草稿后再发布新版本           # 标题/简短描述
type: bug                           # 类型 (bug/task/feature)
status: passed                      # 状态 (open/fixing/resolved/passed/rejected)
priority: major                     # 优先级 (critical/major/minor/low)
owner: dev-agent                    # 负责人 / Agent 名称
created_at: 2026-06-11 10:24        # 创建时间 (YYYY-MM-DD HH:MM)
updated_at: 2026-06-11 11:45        # 更新时间 (YYYY-MM-DD HH:MM)
resolved_at: 2026-06-11 11:45       # 解决/验证通过时间 (仅在终态时填充)
conclusion: 复用现有派发配置修复       # 解决结论 / 修复摘要
---

# 缺陷报告：BUG-S1-04-20260611

## 测试步骤与实际结果
... markdown body content here ...
```
*(注意：Go 解析器会自动捕获并完整保留开发过程中开发者/Agent 写入的前述字段以外的自定义属性，如关联模块等)*
