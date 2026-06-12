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

## 🛠️ 安装与使用

### 1. 编译
在项目根目录下，使用 PowerShell 7 或 Makefile 进行编译：
```powershell
# Windows PowerShell
./make.ps1 build

# Unix / macOS
make build
```
这会在 `plugin/commands/` 目录下生成对应平台的 `ticket` 或 `ticket.exe` 二进制文件。

### 2. 在主流 AI Agent 中安装与集成

根据您使用的 AI 命令行工具，选择以下方式之一进行集成：

#### A. 在 Google Antigravity CLI (`agy`) 中作为插件安装
使用 `agy` 的插件系统从本地路径进行安装：
```bash
# 请将 /path/to/ 替换为项目在您本地的实际绝对路径
agy plugin install /path/to/ticket-cli-plugin/plugin
```
安装成功后：
- `agy` 会自动把 `commands/` 目录放入 Agent 执行会话的系统 `PATH` 中。
- `agy` 会自动加载并激活 `skills/ticket-management/SKILL.md`，使 AI Agent 学会使用该 CLI 命令。

#### B. 在 Claude Code (Anthropic) 中安装与集成
Claude Code 支持通过插件配置文件（`.claude-plugin/plugin.json`）原生加载插件：
1. **本地开发加载**：
   在拉起 Claude Code 命令行时，指定本地插件目录路径：
   ```bash
   claude --plugin-dir /path/to/ticket-cli-plugin/plugin
   ```
2. **会话中直接加载**：
   在已经运行的 Claude Code 交互式会话中运行：
   ```bash
   /plugin add /path/to/ticket-cli-plugin/plugin
   ```

#### C. 在 OpenAI Codex CLI 中作为插件安装

Codex CLI 具备原生的插件与市场管理系统，支持通过本地路径或预编译归档包进行安装。

> [!WARNING]
> **关于 Git 仓库直接安装的限制**：
> 由于可执行二进制文件（如 `ticket` / `ticket.exe`）被排除在 Git 仓库外，**请勿**使用 `codex plugin marketplace add <git-url>` 直接添加 Git 仓库作为市场，否则安装后将由于缺少可执行程序而无法运行。
> 请使用以下两种推荐方法之一进行安装。

##### 方法 1：开发者本地编译安装（推荐本地开发使用）
如果您已经克隆了本仓库：
1. **本地编译二进制文件**：
   在项目根目录下，使用 PowerShell 7 或 make 进行编译：
   ```powershell
   # Windows
   ./make.ps1 build

   # macOS / Linux
   make build
   ```
   这将在 `plugin/commands/` 下生成当前平台对应的可执行文件。
2. **将项目根目录添加为插件市场 (Marketplace)**：
   ```bash
   # 请将 /path/to/ 替换为项目在您本地的实际绝对路径（指向包含 plugin/ 文件夹的根目录，不可带 \plugin 尾缀）
   codex plugin marketplace add /path/to/ticket-cli-plugin
   ```
3. **安装并启用插件**：
   ```bash
   codex plugin add ticket-management-plugin@ticket-cli-plugin
   ```

##### 方法 2：使用 pre-compiled Release 包安装（推荐终端用户使用）
如果您是普通用户，不想手动安装 Go 和编译：
1. **下载 Release 包**：
   前往 GitHub Releases 页面，下载对应您系统平台的预编译 zip 或 tar.gz 插件包（例如 `ticket-management-plugin-windows-amd64.zip`）。
2. **解压到本地目录**：
   将压缩包解压到您指定的目录中（例如 `D:\tools\ticket-plugin\`），解压后该目录下会包含一个 `plugin` 文件夹（其中已包含为您编译好的 `commands/ticket` 或 `ticket.exe` 应用程序）。
3. **将解压目录添加为插件市场 (Marketplace)**：
   ```bash
   codex plugin marketplace add D:\tools\ticket-plugin
   ```
4. **安装并启用插件**：
   ```bash
   codex plugin add ticket-management-plugin@ticket-plugin
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
扫描指定目录下所有工单文件，解析文件名中的 legacy 状态后缀（如 `-pass`、`-passed`、`-fail` 等）写入 Frontmatter `status` 字段中，并自动重命名文件，补齐缺失的元数据（如 `title`、`type`、时间字段等）。
```bash
ticket migrate ./bugs
```
若指定了 `--format json`，将以 JSON 数组返回每份文件的详细迁移结果（包括新旧路径、ID、动作等）。

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
